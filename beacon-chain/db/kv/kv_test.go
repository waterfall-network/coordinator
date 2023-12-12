package kv

import (
	"context"
	"encoding/binary"
	"math/big"
	"sync"
	"testing"

	"github.com/pkg/errors"
	"github.com/prysmaticlabs/go-bitfield"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/async"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/core/blocks"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/core/helpers"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/core/signing"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/state"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/state/stateutil"
	v1 "gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/state/v1"
	stateAltair "gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/state/v2"
	fieldparams "gitlab.waterfall.network/waterfall/protocol/coordinator/config/fieldparams"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/config/params"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/container/trie"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/crypto/bls"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/crypto/hash"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/encoding/bytesutil"
	ethpb "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/require"
	gwatCommon "gitlab.waterfall.network/waterfall/protocol/gwat/common"
)

// DeterministicGenesisStateAltair returns a genesis state in hard fork 1 format made using the deterministic deposits.
func DeterministicGenesisStateAltair(t testing.TB, numValidators uint64) (state.BeaconStateAltair, []bls.SecretKey) {
	deposits, privKeys, err := DeterministicDepositsAndKeys(numValidators)
	if err != nil {
		t.Fatal(errors.Wrapf(err, "failed to get %d deposits", numValidators))
	}
	eth1Data, err := DeterministicEth1Data(len(deposits))
	if err != nil {
		t.Fatal(errors.Wrapf(err, "failed to get eth1data for %d deposits", numValidators))
	}
	beaconState, err := GenesisBeaconState(context.Background(), deposits, uint64(0), eth1Data)
	if err != nil {
		t.Fatal(errors.Wrapf(err, "failed to get genesis beacon state of %d validators", numValidators))
	}
	resetCache()
	return beaconState, privKeys
}

// DeterministicallyGenerateKeys creates BLS private keys using a fixed curve order according to
// the algorithm specified in the Ethereum beacon chain specification interop mock start section found here:
// https://github.com/ethereum/eth2.0-pm/blob/a085c9870f3956d6228ed2a40cd37f0c6580ecd7/interop/mocked_start/README.md
func DeterministicallyGenerateKeys(startIndex, numKeys uint64) ([]bls.SecretKey, []bls.PublicKey, error) {
	privKeys := make([]bls.SecretKey, numKeys)
	pubKeys := make([]bls.PublicKey, numKeys)
	type keys struct {
		secrets []bls.SecretKey
		publics []bls.PublicKey
	}
	// lint:ignore uintcast -- this is safe because we can reasonably expect that the number of keys is less than max int64.
	results, err := async.Scatter(int(numKeys), func(offset int, entries int, _ *sync.RWMutex) (interface{}, error) {
		secs, pubs, err := deterministicallyGenerateKeys(uint64(offset)+startIndex, uint64(entries))
		return &keys{secrets: secs, publics: pubs}, err
	})
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to generate keys")
	}
	for _, result := range results {
		if keysExtent, ok := result.Extent.(*keys); ok {
			copy(privKeys[result.Offset:], keysExtent.secrets)
			copy(pubKeys[result.Offset:], keysExtent.publics)
		} else {
			return nil, nil, errors.New("extent not of expected type")
		}
	}
	return privKeys, pubKeys, nil
}

func deterministicallyGenerateKeys(startIndex, numKeys uint64) ([]bls.SecretKey, []bls.PublicKey, error) {
	privKeys := make([]bls.SecretKey, numKeys)
	pubKeys := make([]bls.PublicKey, numKeys)
	for i := startIndex; i < startIndex+numKeys; i++ {
		enc := make([]byte, 32)
		binary.LittleEndian.PutUint32(enc, uint32(i))
		hash := hash.Hash(enc)
		// Reverse byte order to big endian for use with big ints.
		b := bytesutil.ReverseByteOrder(hash[:])
		num := new(big.Int)
		num = num.SetBytes(b)
		order := new(big.Int)
		var ok bool
		order, ok = order.SetString(bls.CurveOrder, 10)
		if !ok {
			return nil, nil, errors.New("could not set bls curve order as big int")
		}
		num = num.Mod(num, order)
		numBytes := num.Bytes()
		// pad key at the start with zero bytes to make it into a 32 byte key
		if len(numBytes) < 32 {
			emptyBytes := make([]byte, 32-len(numBytes))
			numBytes = append(emptyBytes, numBytes...)
		}
		priv, err := bls.SecretKeyFromBytes(numBytes)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "could not create bls secret key at index %d from raw bytes", i)
		}
		privKeys[i-startIndex] = priv
		pubKeys[i-startIndex] = priv.PublicKey()
	}
	return privKeys, pubKeys, nil
}

// DeterministicDepositsAndKeys returns the entered amount of deposits and secret keys.
// The deposits are configured such that for deposit n the validator
// account is key n and the withdrawal account is key n+1.  As such,
// if all secret keys for n validators are required then numDeposits
// should be n+1.
func DeterministicDepositsAndKeys(numDeposits uint64) ([]*ethpb.Deposit, []bls.SecretKey, error) {
	resetCache()
	lock.Lock()
	defer lock.Unlock()
	var err error

	// Populate trie cache, if not initialized yet.
	if t == nil {
		t, err = trie.NewTrie(params.BeaconConfig().DepositContractTreeDepth)
		if err != nil {
			return nil, nil, errors.Wrap(err, "failed to create new trie")
		}
	}

	// If more deposits requested than cached, generate more.
	if numDeposits > uint64(len(cachedDeposits)) {
		numExisting := uint64(len(cachedDeposits))
		numRequired := numDeposits - uint64(len(cachedDeposits))
		// Fetch the required number of keys.
		secretKeys, publicKeys, err := DeterministicallyGenerateKeys(numExisting, numRequired+1)
		if err != nil {
			return nil, nil, errors.Wrap(err, "could not create deterministic keys: ")
		}
		privKeys = append(privKeys, secretKeys[:len(secretKeys)-1]...)

		// Create the new deposits and add them to the trie.
		for i := uint64(0); i < numRequired; i++ {
			balance := params.BeaconConfig().MaxEffectiveBalance
			deposit, err := signedDeposit(secretKeys[i], publicKeys[i].Marshal(), publicKeys[i+1].Marshal(), balance)
			if err != nil {
				return nil, nil, errors.Wrap(err, "could not create signed deposit")
			}
			cachedDeposits = append(cachedDeposits, deposit)

			hashedDeposit, err := deposit.Data.HashTreeRoot()
			if err != nil {
				return nil, nil, errors.Wrap(err, "could not tree hash deposit data")
			}

			if err = t.Insert(hashedDeposit[:], int(numExisting+i)); err != nil {
				return nil, nil, err
			}
		}
	}

	depositTrie, _, err := DeterministicDepositTrie(int(numDeposits)) // lint:ignore uintcast
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to create deposit trie")
	}
	requestedDeposits := cachedDeposits[:numDeposits]
	for i := range requestedDeposits {
		proof, err := depositTrie.MerkleProof(i)
		if err != nil {
			return nil, nil, errors.Wrap(err, "could not create merkle proof")
		}
		requestedDeposits[i].Proof = proof
	}

	return requestedDeposits, privKeys[0:numDeposits], nil
}

var lock sync.Mutex

// Caches
var cachedDeposits []*ethpb.Deposit
var privKeys []bls.SecretKey
var t *trie.SparseMerkleTrie

// resetCache clears out the old trie, private keys and deposits.
func resetCache() {
	lock.Lock()
	defer lock.Unlock()
	t = nil
	privKeys = []bls.SecretKey{}
	cachedDeposits = []*ethpb.Deposit{}
}

func signedDeposit(
	secretKey bls.SecretKey,
	publicKey,
	withdrawalKey []byte,
	balance uint64,
) (*ethpb.Deposit, error) {
	withdrawalCreds := gwatCommon.BytesToAddress(withdrawalKey)
	creatorAddr := gwatCommon.BytesToAddress(withdrawalKey)
	depositMessage := &ethpb.DepositMessage{
		PublicKey:             publicKey,
		CreatorAddress:        creatorAddr[:],
		WithdrawalCredentials: withdrawalCreds[:],
	}

	domain, err := signing.ComputeDomain(params.BeaconConfig().DomainDeposit, nil, nil)
	if err != nil {
		return nil, errors.Wrap(err, "could not compute domain")
	}
	root, err := depositMessage.HashTreeRoot()
	if err != nil {
		return nil, errors.Wrap(err, "could not get signing root of deposit data")
	}

	sigRoot, err := (&ethpb.SigningData{ObjectRoot: root[:], Domain: domain}).HashTreeRoot()
	if err != nil {
		return nil, err
	}
	depositData := &ethpb.Deposit_Data{
		PublicKey:             publicKey,
		Amount:                balance,
		CreatorAddress:        creatorAddr[:],
		WithdrawalCredentials: withdrawalCreds[:],
		Signature:             secretKey.Sign(sigRoot[:]).Marshal(),
	}

	deposit := &ethpb.Deposit{
		Data: depositData,
	}
	return deposit, nil
}

// DeterministicDepositTrie returns a merkle trie of the requested size from the
// deterministic deposits.
func DeterministicDepositTrie(size int) (*trie.SparseMerkleTrie, [][32]byte, error) {
	if t == nil {
		return nil, [][32]byte{}, errors.New("trie cache is empty, generate deposits at an earlier point")
	}

	return DepositTrieSubset(t, size)
}

// DepositTrieSubset takes in a full tree and the desired size and returns a subset of the deposit trie.
func DepositTrieSubset(sparseTrie *trie.SparseMerkleTrie, size int) (*trie.SparseMerkleTrie, [][32]byte, error) {
	if sparseTrie == nil {
		return nil, [][32]byte{}, errors.New("trie is empty")
	}

	items := sparseTrie.Items()
	if size > len(items) {
		return nil, [][32]byte{}, errors.New("requested a larger tree than amount of deposits")
	}

	items = items[:size]
	depositTrie, err := trie.GenerateTrieFromItems(items, params.BeaconConfig().DepositContractTreeDepth)
	if err != nil {
		return nil, [][32]byte{}, errors.Wrapf(err, "could not generate trie of %d length", size)
	}

	roots := make([][32]byte, len(items))
	for i, dep := range items {
		roots[i] = bytesutil.ToBytes32(dep)
	}
	return depositTrie, roots, nil
}

// DeterministicEth1Data takes an array of deposits and returns the eth1Data made from the deposit trie.
func DeterministicEth1Data(size int) (*ethpb.Eth1Data, error) {
	depositTrie, _, err := DeterministicDepositTrie(size)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create trie")
	}
	root := depositTrie.HashTreeRoot()

	finHash := &gwatCommon.Hash{}
	finHash.SetBytes(root[:])
	candidates := gwatCommon.HashArray{*finHash}

	eth1Data := &ethpb.Eth1Data{
		Candidates:   candidates.ToBytes(),
		BlockHash:    root[:],
		DepositRoot:  root[:],
		DepositCount: uint64(size),
	}
	return eth1Data, nil
}

// GenesisBeaconState returns the genesis beacon state.
func GenesisBeaconState(ctx context.Context, deposits []*ethpb.Deposit, genesisTime uint64, eth1Data *ethpb.Eth1Data) (state.BeaconStateAltair, error) {
	st, err := emptyGenesisState()
	if err != nil {
		return nil, err
	}

	// Process initial deposits.
	st, err = helpers.UpdateGenesisEth1Data(st, deposits, eth1Data)
	if err != nil {
		return nil, err
	}

	st, err = processPreGenesisDeposits(ctx, st, deposits)
	if err != nil {
		return nil, errors.Wrap(err, "could not process validator deposits")
	}

	return buildGenesisBeaconState(genesisTime, st, st.Eth1Data())
}

func buildGenesisBeaconState(genesisTime uint64, preState state.BeaconStateAltair, eth1Data *ethpb.Eth1Data) (state.BeaconStateAltair, error) {
	if eth1Data == nil {
		return nil, errors.New("no eth1data provided for genesis state")
	}

	randaoMixes := make([][]byte, params.BeaconConfig().EpochsPerHistoricalVector)
	for i := 0; i < len(randaoMixes); i++ {
		h := make([]byte, 32)
		copy(h, eth1Data.BlockHash)
		randaoMixes[i] = h
	}

	zeroHash := params.BeaconConfig().ZeroHash[:]

	activeIndexRoots := make([][]byte, params.BeaconConfig().EpochsPerHistoricalVector)
	for i := 0; i < len(activeIndexRoots); i++ {
		activeIndexRoots[i] = zeroHash
	}

	blockRoots := make([][]byte, params.BeaconConfig().SlotsPerHistoricalRoot)
	for i := 0; i < len(blockRoots); i++ {
		blockRoots[i] = zeroHash
	}

	stateRoots := make([][]byte, params.BeaconConfig().SlotsPerHistoricalRoot)
	for i := 0; i < len(stateRoots); i++ {
		stateRoots[i] = zeroHash
	}

	slashings := make([]uint64, params.BeaconConfig().EpochsPerSlashingsVector)

	genesisValidatorsRoot, err := stateutil.ValidatorRegistryRoot(preState.Validators())
	if err != nil {
		return nil, errors.Wrapf(err, "could not hash tree root genesis validators %v", err)
	}

	prevEpochParticipation, err := preState.PreviousEpochParticipation()
	if err != nil {
		return nil, err
	}
	currEpochParticipation, err := preState.CurrentEpochParticipation()
	if err != nil {
		return nil, err
	}
	scores, err := preState.InactivityScores()
	if err != nil {
		return nil, err
	}
	st := &ethpb.BeaconStateAltair{
		// Misc fields.
		Slot:                  0,
		GenesisTime:           genesisTime,
		GenesisValidatorsRoot: genesisValidatorsRoot[:],

		Fork: &ethpb.Fork{
			PreviousVersion: params.BeaconConfig().GenesisForkVersion,
			CurrentVersion:  params.BeaconConfig().GenesisForkVersion,
			Epoch:           0,
		},

		// Validator registry fields.
		Validators:                 preState.Validators(),
		Balances:                   preState.Balances(),
		PreviousEpochParticipation: prevEpochParticipation,
		CurrentEpochParticipation:  currEpochParticipation,
		InactivityScores:           scores,

		// Randomness and committees.
		RandaoMixes: randaoMixes,

		// Finality.
		PreviousJustifiedCheckpoint: &ethpb.Checkpoint{
			Epoch: 0,
			Root:  params.BeaconConfig().ZeroHash[:],
		},
		CurrentJustifiedCheckpoint: &ethpb.Checkpoint{
			Epoch: 0,
			Root:  params.BeaconConfig().ZeroHash[:],
		},
		JustificationBits: []byte{0},
		FinalizedCheckpoint: &ethpb.Checkpoint{
			Epoch: 0,
			Root:  params.BeaconConfig().ZeroHash[:],
		},

		HistoricalRoots: [][]byte{},
		BlockRoots:      blockRoots,
		StateRoots:      stateRoots,
		Slashings:       slashings,

		// Eth1 data.
		Eth1Data:         eth1Data,
		Eth1DataVotes:    make([]*ethpb.Eth1Data, 0),
		BlockVoting:      make([]*ethpb.BlockVoting, 0),
		Eth1DepositIndex: preState.Eth1DepositIndex(),
	}

	bodyRoot, err := (&ethpb.BeaconBlockBodyAltair{
		RandaoReveal: make([]byte, 96),
		Eth1Data: &ethpb.Eth1Data{
			DepositRoot: make([]byte, fieldparams.RootLength),
			BlockHash:   make([]byte, 32),
			Candidates:  make([]byte, 0),
		},
		Graffiti: make([]byte, 32),
		SyncAggregate: &ethpb.SyncAggregate{
			SyncCommitteeBits:      make([]byte, len(bitfield.NewBitvector512())),
			SyncCommitteeSignature: make([]byte, 96),
		},
	}).HashTreeRoot()
	if err != nil {
		return nil, errors.Wrap(err, "could not hash tree root empty block body")
	}

	st.LatestBlockHeader = &ethpb.BeaconBlockHeader{
		ParentRoot: zeroHash,
		StateRoot:  zeroHash,
		BodyRoot:   bodyRoot[:],
	}

	var pubKeys [][]byte
	for i := uint64(0); i < params.BeaconConfig().SyncCommitteeSize; i++ {
		pubKeys = append(pubKeys, bytesutil.PadTo([]byte{}, params.BeaconConfig().BLSPubkeyLength))
	}
	st.CurrentSyncCommittee = &ethpb.SyncCommittee{
		Pubkeys:         pubKeys,
		AggregatePubkey: bytesutil.PadTo([]byte{}, params.BeaconConfig().BLSPubkeyLength),
	}
	st.NextSyncCommittee = &ethpb.SyncCommittee{
		Pubkeys:         bytesutil.SafeCopy2dBytes(pubKeys),
		AggregatePubkey: bytesutil.PadTo([]byte{}, params.BeaconConfig().BLSPubkeyLength),
	}

	return stateAltair.InitializeFromProto(st)
}

func emptyGenesisState() (state.BeaconStateAltair, error) {
	st := &ethpb.BeaconStateAltair{
		// Misc fields.
		Slot: 0,
		Fork: &ethpb.Fork{
			PreviousVersion: params.BeaconConfig().GenesisForkVersion,
			CurrentVersion:  params.BeaconConfig().AltairForkVersion,
			Epoch:           0,
		},
		// Validator registry fields.
		Validators:       []*ethpb.Validator{},
		Balances:         []uint64{},
		InactivityScores: []uint64{},

		JustificationBits:          []byte{0},
		HistoricalRoots:            [][]byte{},
		CurrentEpochParticipation:  []byte{},
		PreviousEpochParticipation: []byte{},

		// Eth1 data.
		Eth1Data:         &ethpb.Eth1Data{},
		Eth1DataVotes:    make([]*ethpb.Eth1Data, 0),
		BlockVoting:      make([]*ethpb.BlockVoting, 0),
		Eth1DepositIndex: 0,
		SpineData: &ethpb.SpineData{
			Spines:       nil,
			Prefix:       nil,
			Finalization: nil,
			CpFinalized:  nil,
			ParentSpines: nil,
		},
	}
	return stateAltair.InitializeFromProto(st)
}

// processPreGenesisDeposits processes a deposit for the beacon state Altair before chain start.
func processPreGenesisDeposits(
	ctx context.Context,
	beaconState state.BeaconStateAltair,
	deposits []*ethpb.Deposit,
) (state.BeaconStateAltair, error) {
	var err error
	beaconState, err = ProcessDeposits(ctx, beaconState, deposits)
	if err != nil {
		return nil, errors.Wrap(err, "could not process deposit")
	}
	beaconState, err = blocks.ActivateValidatorWithEffectiveBalance(beaconState, deposits)
	if err != nil {
		return nil, err
	}
	return beaconState, nil
}

// ProcessDeposits processes validator deposits for beacon state Altair.
func ProcessDeposits(
	ctx context.Context,
	beaconState state.BeaconStateAltair,
	deposits []*ethpb.Deposit,
) (state.BeaconStateAltair, error) {
	batchVerified, err := blocks.BatchVerifyDepositsSignatures(ctx, deposits)
	if err != nil {
		return nil, err
	}

	for _, deposit := range deposits {
		if deposit == nil || deposit.Data == nil {
			return nil, errors.New("got a nil deposit in block")
		}
		beaconState, err = ProcessDeposit(ctx, beaconState, deposit, batchVerified)
		if err != nil {
			return nil, errors.Wrapf(err, "could not process deposit from %#x", bytesutil.Trunc(deposit.Data.PublicKey))
		}
	}
	return beaconState, nil
}

// ProcessDeposit processes validator deposit for beacon state Altair.
func ProcessDeposit(ctx context.Context, beaconState state.BeaconStateAltair, deposit *ethpb.Deposit, verifySignature bool) (state.BeaconStateAltair, error) {
	beaconState, isNewValidator, err := blocks.ProcessDeposit(beaconState, deposit, verifySignature)
	if err != nil {
		return nil, err
	}
	if isNewValidator {
		if err := beaconState.AppendInactivityScore(0); err != nil {
			return nil, err
		}
		if err := beaconState.AppendPreviousParticipationBits(0); err != nil {
			return nil, err
		}
		if err := beaconState.AppendCurrentParticipationBits(0); err != nil {
			return nil, err
		}
	}

	return beaconState, nil
}

////////////////////////////////////

// setupDB instantiates and returns a Store instance.
func setupDB(t testing.TB) *Store {
	db, err := NewKVStore(context.Background(), t.TempDir(), &Config{})
	require.NoError(t, err, "Failed to instantiate DB")
	t.Cleanup(func() {
		require.NoError(t, db.Close(), "Failed to close database")
	})
	return db
}

type NewBeaconStateOption func(state *ethpb.BeaconState) error

// NewBeaconState creates a beacon state with minimum marshalable fields.
func NewBeaconState(options ...NewBeaconStateOption) (state.BeaconState, error) {
	seed := &ethpb.BeaconState{
		BlockRoots:                 filledByteSlice2D(uint64(params.MainnetConfig().SlotsPerHistoricalRoot), 32),
		StateRoots:                 filledByteSlice2D(uint64(params.MainnetConfig().SlotsPerHistoricalRoot), 32),
		Slashings:                  make([]uint64, params.MainnetConfig().EpochsPerSlashingsVector),
		RandaoMixes:                filledByteSlice2D(uint64(params.MainnetConfig().EpochsPerHistoricalVector), 32),
		Validators:                 make([]*ethpb.Validator, 0),
		CurrentJustifiedCheckpoint: &ethpb.Checkpoint{Root: make([]byte, fieldparams.RootLength)},
		Eth1Data: &ethpb.Eth1Data{
			DepositRoot: make([]byte, fieldparams.RootLength),
			BlockHash:   make([]byte, 32),
			Candidates:  make([]byte, 0),
		},
		Fork: &ethpb.Fork{
			PreviousVersion: make([]byte, 4),
			CurrentVersion:  make([]byte, 4),
		},
		Eth1DataVotes:               make([]*ethpb.Eth1Data, 0),
		BlockVoting:                 make([]*ethpb.BlockVoting, 0),
		HistoricalRoots:             make([][]byte, 0),
		JustificationBits:           bitfield.Bitvector4{0x0},
		FinalizedCheckpoint:         &ethpb.Checkpoint{Root: make([]byte, fieldparams.RootLength)},
		LatestBlockHeader:           HydrateBeaconHeader(&ethpb.BeaconBlockHeader{}),
		PreviousEpochAttestations:   make([]*ethpb.PendingAttestation, 0),
		CurrentEpochAttestations:    make([]*ethpb.PendingAttestation, 0),
		PreviousJustifiedCheckpoint: &ethpb.Checkpoint{Root: make([]byte, fieldparams.RootLength)},
		SpineData: &ethpb.SpineData{
			Spines:       []byte{},
			Prefix:       []byte{},
			Finalization: []byte{},
			CpFinalized:  []byte{}, //eth1Data.GetBlockHash(),
			ParentSpines: []*ethpb.SpinesSeq{},
		},
	}

	for _, opt := range options {
		err := opt(seed)
		if err != nil {
			return nil, err
		}
	}

	var st, err = v1.InitializeFromProtoUnsafe(seed)
	if err != nil {
		return nil, err
	}

	return st.Copy().(*v1.BeaconState), nil
}

// NewBeaconBlock creates a beacon block with minimum marshalable fields.
func NewBeaconBlock() *ethpb.SignedBeaconBlock {
	return &ethpb.SignedBeaconBlock{
		Block: &ethpb.BeaconBlock{
			ParentRoot: make([]byte, fieldparams.RootLength),
			StateRoot:  make([]byte, fieldparams.RootLength),
			Body: &ethpb.BeaconBlockBody{
				RandaoReveal: make([]byte, fieldparams.BLSSignatureLength),
				Eth1Data: &ethpb.Eth1Data{
					DepositRoot: make([]byte, fieldparams.RootLength),
					BlockHash:   make([]byte, fieldparams.RootLength),
					Candidates:  make([]byte, 0),
				},
				Graffiti:          make([]byte, fieldparams.RootLength),
				Attestations:      []*ethpb.Attestation{},
				AttesterSlashings: []*ethpb.AttesterSlashing{},
				Deposits:          []*ethpb.Deposit{},
				ProposerSlashings: []*ethpb.ProposerSlashing{},
				VoluntaryExits:    []*ethpb.VoluntaryExit{},
				Withdrawals: []*ethpb.Withdrawal{
					{
						PublicKey:      bytesutil.PadTo([]byte{0x77}, 48),
						ValidatorIndex: 0,
						Amount:         123456789,
						InitTxHash:     bytesutil.PadTo([]byte{0x77}, 32),
						Epoch:          5,
					},
				},
			},
		},
		Signature: make([]byte, fieldparams.BLSSignatureLength),
	}
}

// NewBeaconBlockAltair creates a beacon block with minimum marshalable fields.
func NewBeaconBlockAltair() *ethpb.SignedBeaconBlockAltair {
	return &ethpb.SignedBeaconBlockAltair{
		Block: &ethpb.BeaconBlockAltair{
			ParentRoot: make([]byte, fieldparams.RootLength),
			StateRoot:  make([]byte, fieldparams.RootLength),
			Body: &ethpb.BeaconBlockBodyAltair{
				RandaoReveal: make([]byte, 96),
				Eth1Data: &ethpb.Eth1Data{
					DepositRoot: make([]byte, fieldparams.RootLength),
					BlockHash:   make([]byte, 32),
					Candidates:  make([]byte, 0),
				},
				Graffiti:          make([]byte, 32),
				Attestations:      []*ethpb.Attestation{},
				AttesterSlashings: []*ethpb.AttesterSlashing{},
				Deposits:          []*ethpb.Deposit{},
				ProposerSlashings: []*ethpb.ProposerSlashing{},
				VoluntaryExits:    []*ethpb.VoluntaryExit{},
				SyncAggregate: &ethpb.SyncAggregate{
					SyncCommitteeBits:      make([]byte, len(bitfield.NewBitvector512())),
					SyncCommitteeSignature: make([]byte, 96),
				},
			},
		},
		Signature: make([]byte, 96),
	}
}

// SSZ will fill 2D byte slices with their respective values, so we must fill these in too for round
// trip testing.
func filledByteSlice2D(length, innerLen uint64) [][]byte {
	b := make([][]byte, length)
	for i := uint64(0); i < length; i++ {
		b[i] = make([]byte, innerLen)
	}
	return b
}

// HydrateBeaconHeader hydrates a beacon block header with correct field length sizes
// to comply with fssz marshaling and unmarshalling rules.
func HydrateBeaconHeader(h *ethpb.BeaconBlockHeader) *ethpb.BeaconBlockHeader {
	if h == nil {
		h = &ethpb.BeaconBlockHeader{}
	}
	if h.BodyRoot == nil {
		h.BodyRoot = make([]byte, fieldparams.RootLength)
	}
	if h.StateRoot == nil {
		h.StateRoot = make([]byte, fieldparams.RootLength)
	}
	if h.ParentRoot == nil {
		h.ParentRoot = make([]byte, fieldparams.RootLength)
	}
	return h
}

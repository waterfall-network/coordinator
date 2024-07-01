package transition

import (
	"context"

	"github.com/pkg/errors"
	b "gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/core/blocks"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/core/helpers"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/state"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/state/stateutil"
	v1 "gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/state/v1"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/config/params"
	ethpb "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1"
)

// GenesisBeaconState gets called when MinGenesisActiveValidatorCount count of
// full deposits were made to the deposit contract and the ChainStart log gets emitted.
//
// Spec pseudocode definition:
//
//	def initialize_beacon_state_from_eth1(eth1_block_hash: Bytes32,
//	                                    eth1_timestamp: uint64,
//	                                    deposits: Sequence[Deposit]) -> BeaconState:
//	  fork = Fork(
//	      previous_version=GENESIS_FORK_VERSION,
//	      current_version=GENESIS_FORK_VERSION,
//	      epoch=GENESIS_EPOCH,
//	  )
//	  state = BeaconState(
//	      genesis_time=eth1_timestamp + GENESIS_DELAY,
//	      fork=fork,
//	      eth1_data=Eth1Data(block_hash=eth1_block_hash, deposit_count=uint64(len(deposits))),
//	      latest_block_header=BeaconBlockHeader(body_root=hash_tree_root(BeaconBlockBody())),
//	      randao_mixes=[eth1_block_hash] * EPOCHS_PER_HISTORICAL_VECTOR,  # Seed RANDAO with Eth1 entropy
//	  )
//
//	  # Process deposits
//	  leaves = list(map(lambda deposit: deposit.data, deposits))
//	  for index, deposit in enumerate(deposits):
//	      deposit_data_list = List[DepositData, 2**DEPOSIT_CONTRACT_TREE_DEPTH](*leaves[:index + 1])
//	      state.eth1_data.deposit_root = hash_tree_root(deposit_data_list)
//	      process_deposit(state, deposit)
//
//	  # Process activations
//	  for index, validator in enumerate(state.validators):
//	      balance = state.balances[index]
//	      validator.effective_balance = min(balance - balance % EFFECTIVE_BALANCE_INCREMENT, MAX_EFFECTIVE_BALANCE)
//	      if validator.effective_balance == MAX_EFFECTIVE_BALANCE:
//	          validator.activation_eligibility_epoch = GENESIS_EPOCH
//	          validator.activation_epoch = GENESIS_EPOCH
//
//	  # Set genesis validators root for domain separation and chain versioning
//	  state.genesis_validators_root = hash_tree_root(state.validators)
//
//	  return state
//
// This method differs from the spec so as to process deposits beforehand instead of the end of the function.
func GenesisBeaconState(ctx context.Context, deposits []*ethpb.Deposit, genesisTime uint64, eth1Data *ethpb.Eth1Data) (state.BeaconState, error) {
	st, err := EmptyGenesisState()
	if err != nil {
		return nil, err
	}

	// Process initial deposits.
	st, err = helpers.UpdateGenesisEth1Data(st, deposits, eth1Data)
	if err != nil {
		return nil, err
	}

	st, err = b.ProcessPreGenesisDeposits(ctx, st, deposits)
	if err != nil {
		return nil, errors.Wrap(err, "could not process validator deposits")
	}

	return OptimizedGenesisBeaconState(genesisTime, st, st.Eth1Data())
}

// OptimizedGenesisBeaconState is used to create a state that has already processed deposits. This is to efficiently
// create a mainnet state at chainstart.
func OptimizedGenesisBeaconState(genesisTime uint64, preState state.BeaconState, eth1Data *ethpb.Eth1Data) (state.BeaconState, error) {
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

	st := &ethpb.BeaconState{
		// Misc fields.
		Slot:                  0,
		GenesisTime:           genesisTime,
		GenesisValidatorsRoot: genesisValidatorsRoot[:],

		Fork: &ethpb.Fork{
			PreviousVersion: params.BeaconConfig().GenesisForkVersion,
			CurrentVersion:  params.BeaconConfig().GenesisForkVersion,
			//PreviousVersion: params.PyrmontConfig().GenesisForkVersion,
			//CurrentVersion:  params.PyrmontConfig().GenesisForkVersion,
			Epoch: 0,
		},

		// Validator registry fields.
		Validators: preState.Validators(),
		Balances:   preState.Balances(),

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
		JustificationBits: preState.JustificationBits(),
		FinalizedCheckpoint: &ethpb.Checkpoint{
			Epoch: 0,
			Root:  params.BeaconConfig().ZeroHash[:],
		},

		//HistoricalRoots: [][]byte{},
		HistoricalRoots: make([][]byte, 0, 16777216),
		BlockRoots:      blockRoots,
		StateRoots:      stateRoots,
		Slashings:       slashings,
		//CurrentEpochAttestations: []*ethpb.PendingAttestation{},
		CurrentEpochAttestations: make([]*ethpb.PendingAttestation, 0, 4096),
		//PreviousEpochAttestations: []*ethpb.PendingAttestation{},
		PreviousEpochAttestations: make([]*ethpb.PendingAttestation, 0, 4096),

		// Eth1 data.
		Eth1Data: &ethpb.Eth1Data{
			DepositRoot:  eth1Data.GetDepositRoot(),
			DepositCount: 0,
			BlockHash:    eth1Data.GetBlockHash(),
			Candidates:   eth1Data.GetCandidates(),
		},
		Eth1DataVotes:    make([]*ethpb.Eth1Data, 0),
		BlockVoting:      make([]*ethpb.BlockVoting, 0),
		Eth1DepositIndex: 0,
		// Spine data.
		SpineData: &ethpb.SpineData{
			Spines:       []byte{},
			Prefix:       []byte{},
			Finalization: []byte{},
			CpFinalized:  eth1Data.GetBlockHash(),
			ParentSpines: []*ethpb.SpinesSeq{},
		},
	}

	bodyRoot, err := (&ethpb.BeaconBlockBody{
		RandaoReveal: make([]byte, 96),
		Eth1Data: &ethpb.Eth1Data{
			DepositRoot: make([]byte, 32),
			BlockHash:   make([]byte, 32),
			Candidates:  make([]byte, 0),
		},
		Withdrawals: make([]*ethpb.Withdrawal, 0),
		Graffiti:    make([]byte, 32),
	}).HashTreeRoot()
	if err != nil {
		return nil, errors.Wrap(err, "could not hash tree root empty block body")
	}

	st.LatestBlockHeader = &ethpb.BeaconBlockHeader{
		Slot:          0,
		ProposerIndex: 0,
		ParentRoot:    zeroHash,
		StateRoot:     zeroHash,
		BodyRoot:      bodyRoot[:],
	}

	return v1.InitializeFromProto(st)
}

// EmptyGenesisState returns an empty beacon state object.
func EmptyGenesisState() (state.BeaconState, error) {
	st := &ethpb.BeaconState{
		// Misc fields.
		Slot: 0,
		Fork: &ethpb.Fork{
			PreviousVersion: params.BeaconConfig().GenesisForkVersion,
			CurrentVersion:  params.BeaconConfig().GenesisForkVersion,
			Epoch:           0,
		},
		// Validator registry fields.
		Validators: []*ethpb.Validator{},
		Balances:   []uint64{},

		JustificationBits:         []byte{0},
		HistoricalRoots:           [][]byte{},
		CurrentEpochAttestations:  make([]*ethpb.PendingAttestation, 0),
		PreviousEpochAttestations: make([]*ethpb.PendingAttestation, 0),

		// Eth1 data.
		Eth1Data:         &ethpb.Eth1Data{},
		Eth1DataVotes:    make([]*ethpb.Eth1Data, 0),
		BlockVoting:      make([]*ethpb.BlockVoting, 0),
		Eth1DepositIndex: 0,
		// Spine data.
		SpineData: &ethpb.SpineData{
			Spines:       []byte{},
			Prefix:       []byte{},
			Finalization: []byte{},
			CpFinalized:  []byte{},
			ParentSpines: []*ethpb.SpinesSeq{},
		},
	}
	return v1.InitializeFromProto(st)
}

// IsValidGenesisState gets called whenever there's a deposit event,
// it checks whether there's enough effective balance to trigger and
// if the minimum genesis time arrived already.
//
// Spec pseudocode definition:
//
//	def is_valid_genesis_state(state: BeaconState) -> bool:
//	   if state.genesis_time < MIN_GENESIS_TIME:
//	       return False
//	   if len(get_active_validator_indices(state, GENESIS_EPOCH)) < MIN_GENESIS_ACTIVE_VALIDATOR_COUNT:
//	       return False
//	   return True
//
// This method has been modified from the spec to allow whole states not to be saved
// but instead only cache the relevant information.
func IsValidGenesisState(chainStartDepositCount, currentTime uint64) bool {
	if currentTime < params.BeaconConfig().MinGenesisTime {
		return false
	}
	if chainStartDepositCount < params.BeaconConfig().MinGenesisActiveValidatorCount {
		return false
	}
	return true
}

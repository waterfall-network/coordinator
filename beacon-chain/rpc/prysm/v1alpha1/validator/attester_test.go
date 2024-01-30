package validator

import (
	"context"
	"math/rand"
	"sync"
	"testing"
	"time"

	types "github.com/prysmaticlabs/eth2-types"
	mock "gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/blockchain/testing"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/cache"
	dbutil "gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/db/testing"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/forkchoice/protoarray"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/operations/attestations"
	mockp2p "gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/p2p/testing"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/state/stategen"
	v1 "gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/state/v1"
	mockSync "gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/sync/initial-sync/testing"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/config/params"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/crypto/bls"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/encoding/bytesutil"
	ethpb "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1/wrapper"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/assert"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/require"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/util"
	prysmTime "gitlab.waterfall.network/waterfall/protocol/coordinator/time"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/time/slots"
	"google.golang.org/protobuf/proto"
)

func TestProposeAttestation_OK(t *testing.T) {
	attesterServer := &Server{
		HeadFetcher:       &mock.ChainService{},
		P2P:               &mockp2p.MockBroadcaster{},
		AttestationCache:  cache.NewAttestationCache(),
		AttPool:           attestations.NewPool(),
		OperationNotifier: (&mock.ChainService{}).OperationNotifier(),
	}
	head := util.NewBeaconBlock()
	head.Block.Slot = 999
	head.Block.ParentRoot = bytesutil.PadTo([]byte{'a'}, 32)
	root, err := head.Block.HashTreeRoot()
	require.NoError(t, err)

	validators := make([]*ethpb.Validator, 64)
	for i := 0; i < len(validators); i++ {
		validators[i] = &ethpb.Validator{
			PublicKey:             make([]byte, 48),
			CreatorAddress:        make([]byte, 20),
			WithdrawalCredentials: make([]byte, 20),
			ExitEpoch:             params.BeaconConfig().FarFutureEpoch,
			EffectiveBalance:      params.BeaconConfig().MaxEffectiveBalance,
			ActivationHash:        make([]byte, 32),
			ExitHash:              make([]byte, 32),
			WithdrawalOps:         make([]*ethpb.WithdrawalOp, 0),
		}
	}

	state, err := util.NewBeaconState()
	require.NoError(t, err)
	require.NoError(t, state.SetSlot(params.BeaconConfig().SlotsPerEpoch+1))
	require.NoError(t, state.SetValidators(validators))

	sk, err := bls.RandKey()
	require.NoError(t, err)
	sig := sk.Sign([]byte("dummy_test_data"))
	req := &ethpb.Attestation{
		Signature: sig.Marshal(),
		Data: &ethpb.AttestationData{
			BeaconBlockRoot: root[:],
			Source:          &ethpb.Checkpoint{Root: make([]byte, 32)},
			Target:          &ethpb.Checkpoint{Root: make([]byte, 32)},
		},
	}
	_, err = attesterServer.ProposeAttestation(context.Background(), req)
	assert.NoError(t, err)
}

func TestProposeAttestation_IncorrectSignature(t *testing.T) {
	attesterServer := &Server{
		HeadFetcher:       &mock.ChainService{},
		P2P:               &mockp2p.MockBroadcaster{},
		AttestationCache:  cache.NewAttestationCache(),
		AttPool:           attestations.NewPool(),
		OperationNotifier: (&mock.ChainService{}).OperationNotifier(),
	}

	req := util.HydrateAttestation(&ethpb.Attestation{})
	wanted := "Incorrect attestation signature"
	_, err := attesterServer.ProposeAttestation(context.Background(), req)
	assert.ErrorContains(t, wanted, err)
}

func TestGetAttestationData_OK(t *testing.T) {
	block := util.NewBeaconBlock()
	block.Block.Slot = 3*params.BeaconConfig().SlotsPerEpoch + 1
	targetBlock := util.NewBeaconBlock()
	targetBlock.Block.Slot = 1 * params.BeaconConfig().SlotsPerEpoch
	justifiedBlock := util.NewBeaconBlock()
	justifiedBlock.Block.Slot = 2 * params.BeaconConfig().SlotsPerEpoch
	blockRoot, err := block.Block.HashTreeRoot()
	require.NoError(t, err, "Could not hash beacon block")
	justifiedRoot, err := justifiedBlock.Block.HashTreeRoot()
	require.NoError(t, err, "Could not get signing root for justified block")
	targetRoot, err := targetBlock.Block.HashTreeRoot()
	require.NoError(t, err, "Could not get signing root for target block")
	slot := 3*params.BeaconConfig().SlotsPerEpoch + 1
	beaconState, err := util.NewBeaconState()
	require.NoError(t, err)
	require.NoError(t, beaconState.SetSlot(slot))
	err = beaconState.SetCurrentJustifiedCheckpoint(&ethpb.Checkpoint{
		Epoch: 2,
		Root:  justifiedRoot[:],
	})
	require.NoError(t, err)

	blockRoots := beaconState.BlockRoots()
	blockRoots[1] = blockRoot[:]
	blockRoots[1*params.BeaconConfig().SlotsPerEpoch] = targetRoot[:]
	blockRoots[2*params.BeaconConfig().SlotsPerEpoch] = justifiedRoot[:]
	require.NoError(t, beaconState.SetBlockRoots(blockRoots))
	chainService := &mock.ChainService{
		Genesis: time.Now(),
	}
	offset := int64(slot.Mul(params.BeaconConfig().SecondsPerSlot))
	db := dbutil.SetupDB(t)
	err = db.SaveGenesisData(context.Background(), beaconState)
	require.NoError(t, err)
	attesterServer := &Server{
		P2P:              &mockp2p.MockBroadcaster{},
		SyncChecker:      &mockSync.Sync{IsSyncing: false},
		AttestationCache: cache.NewAttestationCache(),
		HeadFetcher: &mock.ChainService{
			State: beaconState, Root: blockRoot[:],
			ForkChoiceStore: protoarray.New(5, 5),
		},
		FinalizationFetcher: &mock.ChainService{
			CurrentJustifiedCheckPoint: beaconState.CurrentJustifiedCheckpoint(),
		},
		TimeFetcher: &mock.ChainService{
			Genesis: time.Now().Add(time.Duration(-1*offset) * time.Second),
		},
		StateNotifier: chainService.StateNotifier(),
		StateGen:      stategen.New(db),
	}

	req := &ethpb.AttestationDataRequest{
		CommitteeIndex: 0,
		Slot:           3*params.BeaconConfig().SlotsPerEpoch + 1,
	}
	res, err := attesterServer.GetAttestationData(context.Background(), req)
	require.NoError(t, err, "Could not get attestation info at slot")

	expectedInfo := &ethpb.AttestationData{
		Slot:            3*params.BeaconConfig().SlotsPerEpoch + 1,
		BeaconBlockRoot: make([]byte, 32),
		Source: &ethpb.Checkpoint{
			Epoch: 2,
			Root:  justifiedRoot[:],
		},
		Target: &ethpb.Checkpoint{
			Epoch: 3,
			Root:  make([]byte, 32),
		},
	}

	if !proto.Equal(res, expectedInfo) {
		t.Errorf("Expected attestation info to match, received %v, wanted %v", res, expectedInfo)
	}
}

func TestGetAttestationData_SyncNotReady(t *testing.T) {
	as := &Server{
		SyncChecker: &mockSync.Sync{IsSyncing: true},
	}
	_, err := as.GetAttestationData(context.Background(), &ethpb.AttestationDataRequest{})
	assert.ErrorContains(t, "Syncing to latest head", err)
}

func TestAttestationDataAtSlot_HandlesFarAwayJustifiedEpoch(t *testing.T) {
	// Scenario:
	//
	// State slot = 10000
	// Last justified slot = epoch start of 1500
	// HistoricalRootsLimit = 8192
	//
	// More background: https://github.com/prysmaticlabs/prysm/issues/2153
	// This test breaks if it doesnt use mainnet config

	// Ensure HistoricalRootsLimit matches scenario
	params.SetupTestConfigCleanup(t)
	cfg := params.MainnetConfig()
	cfg.HistoricalRootsLimit = 8192
	params.OverrideBeaconConfig(cfg)

	block := util.NewBeaconBlock()
	block.Block.Slot = 10000
	epochBoundaryBlock := util.NewBeaconBlock()
	var err error
	epochBoundaryBlock.Block.Slot, err = slots.EpochStart(slots.ToEpoch(10000))
	require.NoError(t, err)
	justifiedBlock := util.NewBeaconBlock()
	justifiedBlock.Block.Slot, err = slots.EpochStart(slots.ToEpoch(1500))
	require.NoError(t, err)
	justifiedBlock.Block.Slot -= 2 // Imagine two skip block
	blockRoot, err := block.Block.HashTreeRoot()
	require.NoError(t, err, "Could not hash beacon block")
	justifiedBlockRoot, err := justifiedBlock.Block.HashTreeRoot()
	require.NoError(t, err, "Could not hash justified block")
	epochBoundaryRoot, err := epochBoundaryBlock.Block.HashTreeRoot()
	require.NoError(t, err, "Could not hash justified block")
	slot := types.Slot(10000)

	beaconState, err := util.NewBeaconState()
	require.NoError(t, err)
	require.NoError(t, beaconState.SetSlot(slot))
	err = beaconState.SetCurrentJustifiedCheckpoint(&ethpb.Checkpoint{
		Epoch: slots.ToEpoch(1500),
		Root:  justifiedBlockRoot[:],
	})
	require.NoError(t, err)
	blockRoots := beaconState.BlockRoots()
	blockRoots[1] = blockRoot[:]
	blockRoots[1*params.BeaconConfig().SlotsPerEpoch] = epochBoundaryRoot[:]
	blockRoots[2*params.BeaconConfig().SlotsPerEpoch] = justifiedBlockRoot[:]
	require.NoError(t, beaconState.SetBlockRoots(blockRoots))
	chainService := &mock.ChainService{
		Genesis: time.Now(),
	}
	offset := int64(slot.Mul(params.BeaconConfig().SecondsPerSlot))
	db := dbutil.SetupDB(t)
	err = db.SaveGenesisData(context.Background(), beaconState)
	require.NoError(t, err)
	attesterServer := &Server{
		P2P:              &mockp2p.MockBroadcaster{},
		AttestationCache: cache.NewAttestationCache(),
		HeadFetcher:      &mock.ChainService{State: beaconState, Root: blockRoot[:], ForkChoiceStore: protoarray.New(5, 5)},
		FinalizationFetcher: &mock.ChainService{
			CurrentJustifiedCheckPoint: beaconState.CurrentJustifiedCheckpoint(),
		},
		SyncChecker:   &mockSync.Sync{IsSyncing: false},
		TimeFetcher:   &mock.ChainService{Genesis: time.Now().Add(time.Duration(-1*offset) * time.Second)},
		StateNotifier: chainService.StateNotifier(),
		StateGen:      stategen.New(db),
	}

	req := &ethpb.AttestationDataRequest{
		CommitteeIndex: 0,
		Slot:           10000,
	}
	res, err := attesterServer.GetAttestationData(context.Background(), req)
	require.NoError(t, err, "Could not get attestation info at slot")

	expectedInfo := &ethpb.AttestationData{
		Slot:            req.Slot,
		BeaconBlockRoot: make([]byte, 32),
		Source: &ethpb.Checkpoint{
			Epoch: 46,
			Root:  justifiedBlockRoot[:],
		},
		Target: &ethpb.Checkpoint{
			Epoch: 312,
			Root:  make([]byte, 32),
		},
	}

	if !proto.Equal(res, expectedInfo) {
		t.Errorf("Expected attestation info to match, received %v, wanted %v", res, expectedInfo)
	}
}

func TestAttestationDataSlot_handlesInProgressRequest(t *testing.T) {
	s := &ethpb.BeaconState{Slot: 100}
	state, err := v1.InitializeFromProto(s)
	require.NoError(t, err)
	ctx := context.Background()
	chainService := &mock.ChainService{
		Genesis: time.Now(),
	}
	slot := types.Slot(2)
	offset := int64(slot.Mul(params.BeaconConfig().SecondsPerSlot))
	server := &Server{
		HeadFetcher:      &mock.ChainService{State: state},
		AttestationCache: cache.NewAttestationCache(),
		SyncChecker:      &mockSync.Sync{IsSyncing: false},
		TimeFetcher:      &mock.ChainService{Genesis: time.Now().Add(time.Duration(-1*offset) * time.Second)},
		StateNotifier:    chainService.StateNotifier(),
	}

	req := &ethpb.AttestationDataRequest{
		CommitteeIndex: 1,
		Slot:           slot,
	}

	res := &ethpb.AttestationData{
		CommitteeIndex: 1,
		Target:         &ethpb.Checkpoint{Epoch: 55, Root: make([]byte, 32)},
	}

	require.NoError(t, server.AttestationCache.MarkInProgress(req))

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		response, err := server.GetAttestationData(ctx, req)
		require.NoError(t, err)
		if !proto.Equal(res, response) {
			t.Error("Expected  equal responses from cache")
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		assert.NoError(t, server.AttestationCache.Put(ctx, req, res))
		assert.NoError(t, server.AttestationCache.MarkNotInProgress(req))
	}()

	wg.Wait()
}

func TestServer_GetAttestationData_InvalidRequestSlot(t *testing.T) {
	ctx := context.Background()

	slot := 3*params.BeaconConfig().SlotsPerEpoch + 1
	offset := int64(slot.Mul(params.BeaconConfig().SecondsPerSlot))
	attesterServer := &Server{
		SyncChecker: &mockSync.Sync{IsSyncing: false},
		HeadFetcher: &mock.ChainService{},
		TimeFetcher: &mock.ChainService{Genesis: time.Now().Add(time.Duration(-1*offset) * time.Second)},
	}

	req := &ethpb.AttestationDataRequest{
		Slot: 1000000000000,
	}
	_, err := attesterServer.GetAttestationData(ctx, req)
	assert.ErrorContains(t, "invalid request", err)
}

func TestServer_GetAttestationData_HeadStateSlotGreaterThanRequestSlot(t *testing.T) {
	// There exists a rare scenario where the validator may request an attestation for a slot less
	// than the head state's slot. The Ethereum consensus spec constraints require the block root the
	// attestation is referencing be less than or equal to the attestation data slot.
	// See: https://github.com/prysmaticlabs/prysm/issues/5164
	ctx := context.Background()
	db := dbutil.SetupDB(t)

	slot := 3*params.BeaconConfig().SlotsPerEpoch + 1
	block := util.NewBeaconBlock()
	block.Block.Slot = slot
	block2 := util.NewBeaconBlock()
	block2.Block.Slot = slot - 1
	targetBlock := util.NewBeaconBlock()
	targetBlock.Block.Slot = 1 * params.BeaconConfig().SlotsPerEpoch
	justifiedBlock := util.NewBeaconBlock()
	justifiedBlock.Block.Slot = 2 * params.BeaconConfig().SlotsPerEpoch
	blockRoot, err := block.Block.HashTreeRoot()
	require.NoError(t, err, "Could not hash beacon block")
	blockRoot2, err := block2.HashTreeRoot()
	require.NoError(t, err)
	wsb, err := wrapper.WrappedSignedBeaconBlock(block2)
	require.NoError(t, err)
	require.NoError(t, db.SaveBlock(ctx, wsb))
	justifiedRoot, err := justifiedBlock.Block.HashTreeRoot()
	require.NoError(t, err, "Could not get signing root for justified block")
	targetRoot, err := targetBlock.Block.HashTreeRoot()
	require.NoError(t, err, "Could not get signing root for target block")

	beaconState, err := util.NewBeaconState()
	require.NoError(t, err)
	require.NoError(t, beaconState.SetSlot(slot))
	offset := int64(slot.Mul(params.BeaconConfig().SecondsPerSlot))
	require.NoError(t, beaconState.SetGenesisTime(uint64(time.Now().Unix()-offset)))
	err = beaconState.SetLatestBlockHeader(util.HydrateBeaconHeader(&ethpb.BeaconBlockHeader{
		ParentRoot: blockRoot2[:],
	}))
	require.NoError(t, err)
	err = beaconState.SetCurrentJustifiedCheckpoint(&ethpb.Checkpoint{
		Epoch: 2,
		Root:  justifiedRoot[:],
	})
	require.NoError(t, err)
	blockRoots := beaconState.BlockRoots()
	blockRoots[1] = blockRoot[:]
	blockRoots[1*params.BeaconConfig().SlotsPerEpoch] = targetRoot[:]
	blockRoots[2*params.BeaconConfig().SlotsPerEpoch] = justifiedRoot[:]
	blockRoots[3*params.BeaconConfig().SlotsPerEpoch] = blockRoot2[:]
	require.NoError(t, beaconState.SetBlockRoots(blockRoots))

	beaconstate := beaconState.Copy()
	require.NoError(t, beaconstate.SetSlot(beaconstate.Slot()-1))
	require.NoError(t, db.SaveState(ctx, beaconstate, blockRoot2))
	chainService := &mock.ChainService{
		Genesis: time.Now(),
	}
	offset = int64(slot.Mul(params.BeaconConfig().SecondsPerSlot))
	err = db.SaveGenesisData(context.Background(), beaconState)
	require.NoError(t, err)

	attesterServer := &Server{
		P2P:                 &mockp2p.MockBroadcaster{},
		SyncChecker:         &mockSync.Sync{IsSyncing: false},
		AttestationCache:    cache.NewAttestationCache(),
		HeadFetcher:         &mock.ChainService{State: beaconState, Root: blockRoot[:], ForkChoiceStore: protoarray.New(5, 5)},
		FinalizationFetcher: &mock.ChainService{CurrentJustifiedCheckPoint: beaconState.CurrentJustifiedCheckpoint()},
		TimeFetcher:         &mock.ChainService{Genesis: time.Now().Add(time.Duration(-1*offset) * time.Second)},
		StateNotifier:       chainService.StateNotifier(),
		StateGen:            stategen.New(db),
	}
	require.NoError(t, db.SaveState(ctx, beaconState, blockRoot))
	wsb, err = wrapper.WrappedSignedBeaconBlock(block)
	require.NoError(t, err)
	require.NoError(t, db.SaveBlock(ctx, wsb))
	require.NoError(t, db.SaveHeadBlockRoot(ctx, blockRoot))

	req := &ethpb.AttestationDataRequest{
		CommitteeIndex: 0,
		Slot:           slot - 1,
	}
	res, err := attesterServer.GetAttestationData(ctx, req)
	require.NoError(t, err, "Could not get attestation info at slot")

	expectedInfo := &ethpb.AttestationData{
		Slot:            slot - 1,
		BeaconBlockRoot: blockRoot2[:],
		Source: &ethpb.Checkpoint{
			Epoch: 2,
			Root:  justifiedRoot[:],
		},
		Target: &ethpb.Checkpoint{
			Epoch: 3,
			Root:  blockRoot2[:],
		},
	}

	if !proto.Equal(res, expectedInfo) {
		t.Errorf("Expected attestation info to match, received %v, wanted %v", res, expectedInfo)
	}
}

func TestGetAttestationData_SucceedsInFirstEpoch(t *testing.T) {
	slot := types.Slot(5)
	block := util.NewBeaconBlock()
	block.Block.Slot = slot
	targetBlock := util.NewBeaconBlock()
	targetBlock.Block.Slot = 0
	justifiedBlock := util.NewBeaconBlock()
	justifiedBlock.Block.Slot = 0
	blockRoot, err := block.Block.HashTreeRoot()
	require.NoError(t, err, "Could not hash beacon block")
	justifiedRoot, err := justifiedBlock.Block.HashTreeRoot()
	require.NoError(t, err, "Could not get signing root for justified block")
	targetRoot, err := targetBlock.Block.HashTreeRoot()
	require.NoError(t, err, "Could not get signing root for target block")

	beaconState, err := util.NewBeaconState()
	require.NoError(t, err)
	require.NoError(t, beaconState.SetSlot(slot))
	err = beaconState.SetCurrentJustifiedCheckpoint(&ethpb.Checkpoint{
		Epoch: 1,
		Root:  justifiedRoot[:],
	})
	require.NoError(t, err)
	blockRoots := beaconState.BlockRoots()
	blockRoots[1] = blockRoot[:]
	blockRoots[1*params.BeaconConfig().SlotsPerEpoch] = targetRoot[:]
	blockRoots[2*params.BeaconConfig().SlotsPerEpoch] = justifiedRoot[:]
	require.NoError(t, beaconState.SetBlockRoots(blockRoots))
	chainService := &mock.ChainService{
		Genesis: time.Now(),
	}
	offset := int64(slot.Mul(params.BeaconConfig().SecondsPerSlot))
	db := dbutil.SetupDB(t)
	err = db.SaveGenesisData(context.Background(), beaconState)
	require.NoError(t, err)
	attesterServer := &Server{
		P2P:              &mockp2p.MockBroadcaster{},
		SyncChecker:      &mockSync.Sync{IsSyncing: false},
		AttestationCache: cache.NewAttestationCache(),
		HeadFetcher: &mock.ChainService{
			State: beaconState, Root: blockRoot[:],
			ForkChoiceStore: protoarray.New(5, 5),
		},
		FinalizationFetcher: &mock.ChainService{
			CurrentJustifiedCheckPoint: beaconState.CurrentJustifiedCheckpoint(),
		},
		TimeFetcher:   &mock.ChainService{Genesis: prysmTime.Now().Add(time.Duration(-1*offset) * time.Second)},
		StateNotifier: chainService.StateNotifier(),
		StateGen:      stategen.New(db),
	}

	req := &ethpb.AttestationDataRequest{
		CommitteeIndex: 0,
		Slot:           5,
	}
	res, err := attesterServer.GetAttestationData(context.Background(), req)
	require.NoError(t, err, "Could not get attestation info at slot")

	expectedInfo := &ethpb.AttestationData{
		Slot:            slot,
		BeaconBlockRoot: make([]byte, 32),
		Source: &ethpb.Checkpoint{
			Epoch: 1,
			Root:  justifiedRoot[:],
		},
		Target: &ethpb.Checkpoint{
			Epoch: 0,
			Root:  make([]byte, 32),
		},
	}

	if !proto.Equal(res, expectedInfo) {
		t.Errorf("Expected attestation info to match, received %v, wanted %v", res, expectedInfo)
	}
}

func TestServer_SubscribeCommitteeSubnets_NoSlots(t *testing.T) {
	attesterServer := &Server{
		HeadFetcher:       &mock.ChainService{},
		P2P:               &mockp2p.MockBroadcaster{},
		AttestationCache:  cache.NewAttestationCache(),
		AttPool:           attestations.NewPool(),
		OperationNotifier: (&mock.ChainService{}).OperationNotifier(),
	}

	_, err := attesterServer.SubscribeCommitteeSubnets(context.Background(), &ethpb.CommitteeSubnetsSubscribeRequest{
		Slots:        nil,
		CommitteeIds: nil,
		IsAggregator: nil,
	})
	assert.ErrorContains(t, "no attester slots provided", err)
}

func TestServer_SubscribeCommitteeSubnets_DifferentLengthSlots(t *testing.T) {
	// fixed seed
	s := rand.NewSource(10)
	randGen := rand.New(s)

	attesterServer := &Server{
		HeadFetcher:       &mock.ChainService{},
		P2P:               &mockp2p.MockBroadcaster{},
		AttestationCache:  cache.NewAttestationCache(),
		AttPool:           attestations.NewPool(),
		OperationNotifier: (&mock.ChainService{}).OperationNotifier(),
	}

	var slots []types.Slot
	var comIdxs []types.CommitteeIndex
	var isAggregator []bool

	for i := types.Slot(100); i < 200; i++ {
		slots = append(slots, i)
		comIdxs = append(comIdxs, types.CommitteeIndex(randGen.Int63n(64)))
		boolVal := randGen.Uint64()%2 == 0
		isAggregator = append(isAggregator, boolVal)
	}

	slots = append(slots, 321)

	_, err := attesterServer.SubscribeCommitteeSubnets(context.Background(), &ethpb.CommitteeSubnetsSubscribeRequest{
		Slots:        slots,
		CommitteeIds: comIdxs,
		IsAggregator: isAggregator,
	})
	assert.ErrorContains(t, "request fields are not the same length", err)
}

func TestServer_SubscribeCommitteeSubnets_MultipleSlots(t *testing.T) {
	// fixed seed
	s := rand.NewSource(10)
	randGen := rand.New(s)

	validators := make([]*ethpb.Validator, 64)
	for i := 0; i < len(validators); i++ {
		validators[i] = &ethpb.Validator{
			ExitEpoch:        params.BeaconConfig().FarFutureEpoch,
			EffectiveBalance: params.BeaconConfig().MaxEffectiveBalance,
			ActivationHash:   make([]byte, 32),
			ExitHash:         make([]byte, 32),
			WithdrawalOps:    make([]*ethpb.WithdrawalOp, 0),
		}
	}

	state, err := util.NewBeaconState()
	require.NoError(t, err)
	require.NoError(t, state.SetValidators(validators))

	attesterServer := &Server{
		HeadFetcher:       &mock.ChainService{State: state},
		P2P:               &mockp2p.MockBroadcaster{},
		AttestationCache:  cache.NewAttestationCache(),
		AttPool:           attestations.NewPool(),
		OperationNotifier: (&mock.ChainService{}).OperationNotifier(),
	}

	var slots []types.Slot
	var comIdxs []types.CommitteeIndex
	var isAggregator []bool

	for i := types.Slot(100); i < 200; i++ {
		slots = append(slots, i)
		comIdxs = append(comIdxs, types.CommitteeIndex(randGen.Int63n(64)))
		boolVal := randGen.Uint64()%2 == 0
		isAggregator = append(isAggregator, boolVal)
	}

	_, err = attesterServer.SubscribeCommitteeSubnets(context.Background(), &ethpb.CommitteeSubnetsSubscribeRequest{
		Slots:        slots,
		CommitteeIds: comIdxs,
		IsAggregator: isAggregator,
	})
	require.NoError(t, err)
	for i := types.Slot(100); i < 200; i++ {
		subnets := cache.SubnetIDs.GetAttesterSubnetIDs(i)
		assert.Equal(t, 1, len(subnets))
		if isAggregator[i-100] {
			subnets = cache.SubnetIDs.GetAggregatorSubnetIDs(i)
			assert.Equal(t, 1, len(subnets))
		}
	}
}

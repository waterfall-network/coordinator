package validator

import (
	"context"
	"testing"
	"time"

	types "github.com/prysmaticlabs/eth2-types"
	mockChain "gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/blockchain/testing"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/core/feed"
	opfeed "gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/core/feed/operation"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/core/transition"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/operations/voluntaryexits"
	mockp2p "gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/p2p/testing"
	mockSync "gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/sync/initial-sync/testing"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/config/params"
	ethpb "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/assert"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/require"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/util"
)

func TestProposeExit_Notification(t *testing.T) {
	ctx := context.Background()

	deposits, _, err := util.DeterministicDepositsAndKeys(params.BeaconConfig().MinGenesisActiveValidatorCount)
	require.NoError(t, err)
	beaconState, err := transition.GenesisBeaconState(ctx, deposits, 0, &ethpb.Eth1Data{Candidates: make([]byte, 0), BlockHash: make([]byte, 32)})
	require.NoError(t, err)
	epoch := types.Epoch(2048)
	require.NoError(t, beaconState.SetSlot(params.BeaconConfig().SlotsPerEpoch.Mul(uint64(epoch))))
	block := util.NewBeaconBlock()
	genesisRoot, err := block.Block.HashTreeRoot()
	require.NoError(t, err, "Could not get signing root")

	// Set genesis time to be 100 epochs ago.
	offset := int64(params.BeaconConfig().SlotsPerEpoch.Mul(params.BeaconConfig().SecondsPerSlot))
	genesisTime := time.Now().Add(time.Duration(-100*offset) * time.Second)
	mockChainService := &mockChain.ChainService{State: beaconState, Root: genesisRoot[:], Genesis: genesisTime}
	server := &Server{
		HeadFetcher:       mockChainService,
		SyncChecker:       &mockSync.Sync{IsSyncing: false},
		TimeFetcher:       mockChainService,
		StateNotifier:     mockChainService.StateNotifier(),
		OperationNotifier: mockChainService.OperationNotifier(),
		ExitPool:          voluntaryexits.NewPool(),
		P2P:               mockp2p.NewTestP2P(t),
	}

	// Subscribe to operation notifications.
	opChannel := make(chan *feed.Event, 1024)
	opSub := server.OperationNotifier.OperationFeed().Subscribe(opChannel)
	defer opSub.Unsubscribe()

	// Send the request, expect a result on the state feed.
	validatorIndex := types.ValidatorIndex(0)
	req := &ethpb.VoluntaryExit{
		Epoch:          epoch,
		ValidatorIndex: validatorIndex,
		InitTxHash:     make([]byte, 32),
	}

	resp, err := server.ProposeExit(context.Background(), req)
	require.NoError(t, err)
	expectedRoot, err := req.HashTreeRoot()
	require.NoError(t, err)
	assert.DeepEqual(t, expectedRoot[:], resp.ExitRoot)

	// Ensure the state notification was broadcast.
	notificationFound := false
	for !notificationFound {
		select {
		case event := <-opChannel:
			if event.Type == opfeed.ExitReceived {
				notificationFound = true
				data, ok := event.Data.(*opfeed.ExitReceivedData)
				assert.Equal(t, true, ok, "Entity is of the wrong type")
				assert.NotNil(t, data)
			}
		case <-opSub.Err():
			t.Error("Subscription to state notifier failed")
			return
		}
	}
}

func TestProposeExit_NoPanic(t *testing.T) {
	ctx := context.Background()

	deposits, _, err := util.DeterministicDepositsAndKeys(params.BeaconConfig().MinGenesisActiveValidatorCount)
	require.NoError(t, err)
	beaconState, err := transition.GenesisBeaconState(ctx, deposits, 0, &ethpb.Eth1Data{BlockHash: make([]byte, 32), Candidates: make([]byte, 0)})
	require.NoError(t, err)
	epoch := types.Epoch(2048)
	require.NoError(t, beaconState.SetSlot(params.BeaconConfig().SlotsPerEpoch.Mul(uint64(epoch))))
	block := util.NewBeaconBlock()
	genesisRoot, err := block.Block.HashTreeRoot()
	require.NoError(t, err, "Could not get signing root")

	// Set genesis time to be 100 epochs ago.
	offset := int64(params.BeaconConfig().SlotsPerEpoch.Mul(params.BeaconConfig().SecondsPerSlot))
	genesisTime := time.Now().Add(time.Duration(-100*offset) * time.Second)
	mockChainService := &mockChain.ChainService{State: beaconState, Root: genesisRoot[:], Genesis: genesisTime}
	server := &Server{
		HeadFetcher:       mockChainService,
		SyncChecker:       &mockSync.Sync{IsSyncing: false},
		TimeFetcher:       mockChainService,
		StateNotifier:     mockChainService.StateNotifier(),
		OperationNotifier: mockChainService.OperationNotifier(),
		ExitPool:          voluntaryexits.NewPool(),
		P2P:               mockp2p.NewTestP2P(t),
	}

	// Subscribe to operation notifications.
	opChannel := make(chan *feed.Event, 1024)
	opSub := server.OperationNotifier.OperationFeed().Subscribe(opChannel)
	defer opSub.Unsubscribe()

	req := &ethpb.VoluntaryExit{}
	_, err = server.ProposeExit(context.Background(), req)
	require.ErrorContains(t, "Could not get tree hash of exit: bytes array does not have the correct length", err, "Expected error for no exit existing")

	// Send the request, expect a result on the state feed.
	validatorIndex := types.ValidatorIndex(0)
	req = &ethpb.VoluntaryExit{
		Epoch:          epoch,
		ValidatorIndex: validatorIndex,
		InitTxHash:     make([]byte, 32),
	}

	resp, err := server.ProposeExit(context.Background(), req)
	require.NoError(t, err)
	expectedRoot, err := req.HashTreeRoot()
	require.NoError(t, err)
	assert.DeepEqual(t, expectedRoot[:], resp.ExitRoot)
}

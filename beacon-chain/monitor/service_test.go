package monitor

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	types "github.com/prysmaticlabs/eth2-types"
	logTest "github.com/sirupsen/logrus/hooks/test"
	mock "gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/blockchain/testing"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/core/altair"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/core/feed"
	statefeed "gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/core/feed/state"
	testDB "gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/db/testing"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/state/stategen"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1/wrapper"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/require"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/util"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/time/slots"
)

func TestTrackedIndex(t *testing.T) {
	s := &Service{
		TrackedValidators: map[types.ValidatorIndex]bool{
			1: true,
			2: true,
		},
	}
	require.Equal(t, s.trackedIndex(types.ValidatorIndex(1)), true)
	require.Equal(t, s.trackedIndex(types.ValidatorIndex(3)), false)
}

func TestNewService(t *testing.T) {
	config := &ValidatorMonitorConfig{}
	tracked := []types.ValidatorIndex{}
	ctx := context.Background()
	_, err := NewService(ctx, config, tracked)
	require.NoError(t, err)
}

func TestStart(t *testing.T) {
	hook := logTest.NewGlobal()
	s := setupService(t)
	stateChannel := make(chan *feed.Event, 1)
	stateSub := s.config.StateNotifier.StateFeed().Subscribe(stateChannel)
	defer stateSub.Unsubscribe()

	wg := &sync.WaitGroup{}
	wg.Add(1)
	s.Start()

	go func() {
		select {
		case stateEvent := <-stateChannel:
			if stateEvent.Type == statefeed.Synced {
				_, ok := stateEvent.Data.(*statefeed.SyncedData)
				require.Equal(t, true, ok, "Event feed data is not type *statefeed.SyncedData")
			}
		case <-s.ctx.Done():
		}
		wg.Done()
	}()

	for sent := 0; sent == 0; {
		sent = s.config.StateNotifier.StateFeed().Send(&feed.Event{
			Type: statefeed.Synced,
			Data: &statefeed.SyncedData{
				StartTime: time.Now(),
			},
		})
	}

	// wait for Logrus
	time.Sleep(1000 * time.Millisecond)
	require.LogsContain(t, hook, "Synced to head epoch, starting reporting performance")
	require.LogsContain(t, hook, "\"Starting service\" ValidatorIndices=\"[1 2 7 12]\"")
	require.Equal(t, s.isLogging, true, "monitor is not running")
}

func TestInitializePerformanceStructures(t *testing.T) {
	hook := logTest.NewGlobal()
	ctx := context.Background()
	s := setupService(t)
	state, err := s.config.HeadFetcher.HeadState(ctx)
	require.NoError(t, err)
	epoch := slots.ToEpoch(state.Slot())
	s.initializePerformanceStructures(state, epoch)
	require.LogsDoNotContain(t, hook, "Could not fetch starting balance")
	latestPerformance := map[types.ValidatorIndex]ValidatorLatestPerformance{
		1: {
			balance: 3200000000000,
		},
		2: {
			balance: 3200000000000,
		},
		7: {
			balance: 3200000000000,
		},
		12: {
			balance: 3200000000000,
		},
	}
	aggregatedPerformance := map[types.ValidatorIndex]ValidatorAggregatedPerformance{
		1: {
			startBalance: 3200000000000,
		},
		2: {
			startBalance: 3200000000000,
		},
		7: {
			startBalance: 3200000000000,
		},
		12: {
			startBalance: 3200000000000,
		},
	}

	require.DeepEqual(t, s.latestPerformance, latestPerformance)
	require.DeepEqual(t, s.aggregatedPerformance, aggregatedPerformance)
}

func TestMonitorRoutine(t *testing.T) {
	ctx := context.Background()
	hook := logTest.NewGlobal()
	s := setupService(t)
	stateChannel := make(chan *feed.Event, 1)
	stateSub := s.config.StateNotifier.StateFeed().Subscribe(stateChannel)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		s.monitorRoutine(stateChannel, stateSub)
		wg.Done()
	}()

	genesis, keys := util.DeterministicGenesisStateAltair(t, 64)
	c, err := altair.NextSyncCommittee(ctx, genesis)
	require.NoError(t, err)
	require.NoError(t, genesis.SetCurrentSyncCommittee(c))

	genConfig := util.DefaultBlockGenConfig()
	block, err := util.GenerateFullBlockAltair(genesis, keys, genConfig, 1)
	require.NoError(t, err)
	root, err := block.GetBlock().HashTreeRoot()
	require.NoError(t, err)
	require.NoError(t, s.config.StateGen.SaveState(ctx, root, genesis))

	wrapped, err := wrapper.WrappedSignedBeaconBlock(block)
	require.NoError(t, err)

	stateChannel <- &feed.Event{
		Type: statefeed.BlockProcessed,
		Data: &statefeed.BlockProcessedData{
			Slot:        1,
			Verified:    true,
			SignedBlock: wrapped,
		},
	}

	// Wait for Logrus
	time.Sleep(1000 * time.Millisecond)
	wanted1 := fmt.Sprintf("\"Sync committee contribution included\" BalanceChange=3168000000000 ContribCount=4 ExpectedContribCount=4 NewBalance=3200000000000 ValidatorIndex=1 prefix=monitor")
	wanted2 := fmt.Sprintf("\"Sync committee contribution included\" BalanceChange=3168100000000 ContribCount=2 ExpectedContribCount=2 NewBalance=3200000000000 ValidatorIndex=12 prefix=monitor")
	require.LogsContain(t, hook, wanted1)
	require.LogsContain(t, hook, wanted2)

}

func TestWaitForSync(t *testing.T) {
	s := setupService(t)
	stateChannel := make(chan *feed.Event, 1)
	stateSub := s.config.StateNotifier.StateFeed().Subscribe(stateChannel)
	defer stateSub.Unsubscribe()

	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		err := s.waitForSync(stateChannel, stateSub)
		require.NoError(t, err)
		wg.Done()
	}()

	stateChannel <- &feed.Event{
		Type: statefeed.Synced,
		Data: &statefeed.SyncedData{
			StartTime: time.Now(),
		},
	}
}

func TestRun(t *testing.T) {
	hook := logTest.NewGlobal()
	s := setupService(t)
	stateChannel := make(chan *feed.Event, 1)
	stateSub := s.config.StateNotifier.StateFeed().Subscribe(stateChannel)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		s.run(stateChannel, stateSub)
		wg.Done()
	}()

	stateChannel <- &feed.Event{
		Type: statefeed.Synced,
		Data: &statefeed.SyncedData{
			StartTime: time.Now(),
		},
	}
	//wait for Logrus
	time.Sleep(1000 * time.Millisecond)
	require.LogsContain(t, hook, "Synced to head epoch, starting reporting performance")
}

func setupService(t *testing.T) *Service {
	beaconDB := testDB.SetupDB(t)
	state, _ := util.DeterministicGenesisStateAltair(t, 256)

	pubKeys := make([][]byte, 3)
	pubKeys[0] = state.Validators()[0].PublicKey
	pubKeys[1] = state.Validators()[1].PublicKey
	pubKeys[2] = state.Validators()[2].PublicKey

	currentSyncCommittee := util.ConvertToCommittee([][]byte{
		pubKeys[0], pubKeys[1], pubKeys[2], pubKeys[1], pubKeys[1],
	})
	require.NoError(t, state.SetCurrentSyncCommittee(currentSyncCommittee))

	chainService := &mock.ChainService{
		Genesis:        time.Now(),
		DB:             beaconDB,
		State:          state,
		Root:           []byte("hello-world"),
		ValidatorsRoot: [32]byte{},
	}

	trackedVals := map[types.ValidatorIndex]bool{
		1:  true,
		2:  true,
		7:  true,
		12: true,
	}
	latestPerformance := map[types.ValidatorIndex]ValidatorLatestPerformance{
		1: {
			balance: 32000000000,
		},
		2: {
			balance: 32000000000,
		},
		7: {
			balance: 31900000000,
		},
		12: {
			balance: 31900000000,
		},
	}
	aggregatedPerformance := map[types.ValidatorIndex]ValidatorAggregatedPerformance{
		1: {
			startEpoch:                     0,
			startBalance:                   31700000000,
			totalAttestedCount:             12,
			totalRequestedCount:            15,
			totalDistance:                  14,
			totalCorrectHead:               8,
			totalCorrectSource:             11,
			totalCorrectTarget:             12,
			totalProposedCount:             1,
			totalSyncComitteeContributions: 0,
			totalSyncComitteeAggregations:  0,
		},
		2:  {},
		7:  {},
		12: {},
	}
	trackedSyncCommitteeIndices := map[types.ValidatorIndex][]types.CommitteeIndex{
		1:  {0, 1, 2, 3},
		12: {4, 5},
	}
	return &Service{
		config: &ValidatorMonitorConfig{
			StateGen:            stategen.New(beaconDB),
			StateNotifier:       chainService.StateNotifier(),
			HeadFetcher:         chainService,
			AttestationNotifier: chainService.OperationNotifier(),
		},

		ctx:                         context.Background(),
		TrackedValidators:           trackedVals,
		latestPerformance:           latestPerformance,
		aggregatedPerformance:       aggregatedPerformance,
		trackedSyncCommitteeIndices: trackedSyncCommitteeIndices,
		lastSyncedEpoch:             0,
	}
}

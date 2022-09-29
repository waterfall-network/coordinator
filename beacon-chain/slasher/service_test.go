package slasher

import (
	"context"
	"io/ioutil"
	"testing"
	"time"

	types "github.com/prysmaticlabs/eth2-types"
	"github.com/sirupsen/logrus"
	logTest "github.com/sirupsen/logrus/hooks/test"
	"github.com/waterfall-foundation/coordinator/async/event"
	mock "github.com/waterfall-foundation/coordinator/beacon-chain/blockchain/testing"
	"github.com/waterfall-foundation/coordinator/beacon-chain/core/feed"
	statefeed "github.com/waterfall-foundation/coordinator/beacon-chain/core/feed/state"
	dbtest "github.com/waterfall-foundation/coordinator/beacon-chain/db/testing"
	mockslasher "github.com/waterfall-foundation/coordinator/beacon-chain/slasher/mock"
	mockSync "github.com/waterfall-foundation/coordinator/beacon-chain/sync/initial-sync/testing"
	"github.com/waterfall-foundation/coordinator/testing/require"
	"github.com/waterfall-foundation/coordinator/testing/util"
	"github.com/waterfall-foundation/coordinator/time/slots"
)

var _ = SlashingChecker(&Service{})
var _ = SlashingChecker(&mockslasher.MockSlashingChecker{})

func TestMain(m *testing.M) {
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetOutput(ioutil.Discard)

	m.Run()
}

func TestService_StartStop_ChainInitialized(t *testing.T) {
	slasherDB := dbtest.SetupSlasherDB(t)
	hook := logTest.NewGlobal()
	beaconState, err := util.NewBeaconState()
	require.NoError(t, err)
	currentSlot := types.Slot(4)
	require.NoError(t, beaconState.SetSlot(currentSlot))
	mockChain := &mock.ChainService{
		State: beaconState,
		Slot:  &currentSlot,
	}
	srv, err := New(context.Background(), &ServiceConfig{
		IndexedAttestationsFeed: new(event.Feed),
		BeaconBlockHeadersFeed:  new(event.Feed),
		StateNotifier:           &mock.MockStateNotifier{},
		Database:                slasherDB,
		HeadStateFetcher:        mockChain,
		SyncChecker:             &mockSync.Sync{IsSyncing: false},
	})
	require.NoError(t, err)
	go srv.Start()
	time.Sleep(time.Millisecond * 100)
	srv.serviceCfg.StateNotifier.StateFeed().Send(&feed.Event{
		Type: statefeed.Initialized,
		Data: &statefeed.InitializedData{StartTime: time.Now()},
	})
	time.Sleep(time.Millisecond * 100)
	srv.attsSlotTicker = &slots.SlotTicker{}
	srv.blocksSlotTicker = &slots.SlotTicker{}
	srv.pruningSlotTicker = &slots.SlotTicker{}
	require.NoError(t, srv.Stop())
	require.NoError(t, srv.Status())
	require.LogsContain(t, hook, "received chain initialization")
}

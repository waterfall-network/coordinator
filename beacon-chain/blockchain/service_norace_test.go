package blockchain

import (
	"context"
	"io/ioutil"
	"testing"

	"github.com/sirupsen/logrus"
	testDB "github.com/waterfall-foundation/coordinator/beacon-chain/db/testing"
	"github.com/waterfall-foundation/coordinator/proto/prysm/v1alpha1/wrapper"
	"github.com/waterfall-foundation/coordinator/testing/require"
	"github.com/waterfall-foundation/coordinator/testing/util"
)

func init() {
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetOutput(ioutil.Discard)
}

func TestChainService_SaveHead_DataRace(t *testing.T) {
	beaconDB := testDB.SetupDB(t)
	s := &Service{
		cfg: &config{BeaconDB: beaconDB},
	}
	b, err := wrapper.WrappedSignedBeaconBlock(util.NewBeaconBlock())
	st, _ := util.DeterministicGenesisState(t, 1)
	require.NoError(t, err)
	go func() {
		require.NoError(t, s.saveHead(context.Background(), [32]byte{}, b, st))
	}()
	require.NoError(t, s.saveHead(context.Background(), [32]byte{}, b, st))
}

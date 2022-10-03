package peers_test

import (
	"io/ioutil"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/waterfall-foundation/coordinator/cmd/beacon-chain/flags"
	"github.com/waterfall-foundation/coordinator/config/features"
)

func TestMain(m *testing.M) {
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetOutput(ioutil.Discard)

	resetCfg := features.InitWithReset(&features.Flags{
		EnablePeerScorer: true,
	})
	defer resetCfg()

	resetFlags := flags.Get()
	flags.Init(&flags.GlobalFlags{
		BlockBatchLimit:            64,
		BlockBatchLimitBurstFactor: 10,
	})
	defer func() {
		flags.Init(resetFlags)
	}()
	m.Run()
}

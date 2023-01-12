package sanity

import (
	"testing"

	"gitlab.waterfall.network/waterfall/protocol/coordinator/config/features"
)

func TestMain(m *testing.M) {
	resetCfg := features.InitWithReset(&features.Flags{EnableBalanceTrieComputation: true})
	defer resetCfg()
	m.Run()
}

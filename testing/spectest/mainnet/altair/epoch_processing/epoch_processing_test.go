package epoch_processing

import (
	"testing"

	"github.com/waterfall-foundation/coordinator/config/features"
)

func TestMain(m *testing.M) {
	resetCfg := features.InitWithReset(&features.Flags{EnableBalanceTrieComputation: true})
	defer resetCfg()
	m.Run()
}

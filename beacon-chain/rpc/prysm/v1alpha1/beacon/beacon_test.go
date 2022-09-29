package beacon

import (
	"testing"

	"github.com/waterfall-foundation/coordinator/cmd/beacon-chain/flags"
	"github.com/waterfall-foundation/coordinator/config/params"
)

func TestMain(m *testing.M) {
	// Use minimal config to reduce test setup time.
	prevConfig := params.BeaconConfig().Copy()
	defer params.OverrideBeaconConfig(prevConfig)
	params.OverrideBeaconConfig(params.MinimalSpecConfig())

	resetFlags := flags.Get()
	flags.Init(&flags.GlobalFlags{
		MinimumSyncPeers: 30,
	})
	defer func() {
		flags.Init(resetFlags)
	}()

	m.Run()
}

package kv

import (
	"gitlab.waterfall.network/waterfall/protocol/coordinator/config/params"
)

func init() {
	// Override network name so that hardcoded genesis files are not loaded.
	cfg := params.BeaconConfig()
	cfg.ConfigName = "test"
	params.OverrideBeaconConfig(cfg)
}

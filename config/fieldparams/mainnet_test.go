//go:build !minimal
// +build !minimal

package field_params_test

import (
	"testing"

	fieldparams "gitlab.waterfall.network/waterfall/protocol/coordinator/config/fieldparams"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/config/params"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/assert"
)

func TestFieldParametersValues(t *testing.T) {
	params.UseMainnetConfig()
	assert.Equal(t, "mainnet", fieldparams.Preset)
	testFieldParametersMatchConfig(t)
}

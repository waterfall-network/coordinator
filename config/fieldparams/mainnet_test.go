//go:build !minimal
// +build !minimal

package field_params_test

import (
	"testing"

	fieldparams "github.com/waterfall-foundation/coordinator/config/fieldparams"
	"github.com/waterfall-foundation/coordinator/config/params"
	"github.com/waterfall-foundation/coordinator/testing/assert"
)

func TestFieldParametersValues(t *testing.T) {
	params.UseMainnetConfig()
	assert.Equal(t, "mainnet", fieldparams.Preset)
	testFieldParametersMatchConfig(t)
}

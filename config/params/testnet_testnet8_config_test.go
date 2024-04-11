package params_test

import (
	"testing"

	"gitlab.waterfall.network/waterfall/protocol/coordinator/config/params"
)

func TestTestnet8ConfigMatchesUpstreamYaml(t *testing.T) {
	t.Skip()
	presetFPs := presetsFilePath(t, "mainnet")
	for _, fp := range presetFPs {
		params.LoadChainConfigFile(fp, nil)
	}
	configFP := testnetConfigFilePath(t, "testnet8")
	params.LoadChainConfigFile(configFP, nil)
	fields := fieldsFromYamls(t, append(presetFPs, configFP))
	assertYamlFieldsMatch(t, "testnet8", fields, params.BeaconConfig(), params.Testnet8Config())
}

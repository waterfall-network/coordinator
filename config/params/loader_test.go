package params_test

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/bazelbuild/rules_go/go/tools/bazel"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/config/params"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/io/file"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/assert"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/require"
)

var placeholderFields = []string{"UPDATE_TIMEOUT", "INTERVALS_PER_SLOT"}

func TestLoadConfigFile_OverwriteCorrectly(t *testing.T) {
	file, err := ioutil.TempFile("", "")
	require.NoError(t, err)
	// Set current config to minimal config
	params.OverrideBeaconConfig(params.MainnetConfig())

	// load empty config file, so that it defaults to mainnet values
	params.LoadChainConfigFile(file.Name(), nil)
	if params.BeaconConfig().MinGenesisTime != params.MainnetConfig().MinGenesisTime {
		t.Errorf("Expected MinGenesisTime to be set to mainnet value: %d found: %d",
			params.MainnetConfig().MinGenesisTime,
			params.BeaconConfig().MinGenesisTime)
	}

	if params.BeaconConfig().SlotsPerEpoch != params.MainnetConfig().SlotsPerEpoch {
		t.Errorf("Expected SlotsPerEpoch to be set to mainnet value: %d found: %d",
			params.MainnetConfig().SlotsPerEpoch,
			params.BeaconConfig().SlotsPerEpoch)
	}
	require.Equal(t, "devnet", params.BeaconConfig().ConfigName)
}

func Test_replaceHexStringWithYAMLFormat(t *testing.T) {
	testLines := []struct {
		line   string
		wanted string
	}{
		{
			line:   "ONE_BYTE: 0x41",
			wanted: "ONE_BYTE: 65\n",
		},
		{
			line:   "FOUR_BYTES: 0x41414141",
			wanted: "FOUR_BYTES: \n- 65\n- 65\n- 65\n- 65\n",
		},
		{
			line:   "THREE_BYTES: 0x414141",
			wanted: "THREE_BYTES: \n- 65\n- 65\n- 65\n- 0\n",
		},
		{
			line:   "EIGHT_BYTES: 0x4141414141414141",
			wanted: "EIGHT_BYTES: \n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n",
		},
		{
			line: "SIXTEEN_BYTES: 0x41414141414141414141414141414141",
			wanted: "SIXTEEN_BYTES: \n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n" +
				"- 65\n- 65\n- 65\n- 65\n",
		},
		{
			line: "TWENTY_BYTES: 0x4141414141414141414141414141414141414141",
			wanted: "TWENTY_BYTES: \n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n" +
				"- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n",
		},
		{
			line: "THIRTY_TWO_BYTES: 0x4141414141414141414141414141414141414141414141414141414141414141",
			wanted: "THIRTY_TWO_BYTES: \n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n" +
				"- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n" +
				"- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n",
		},
		{
			line: "FORTY_EIGHT_BYTES: 0x41414141414141414141414141414141414141414141414141414141414141414141" +
				"4141414141414141414141414141",
			wanted: "FORTY_EIGHT_BYTES: \n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n" +
				"- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n" +
				"- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n" +
				"- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n",
		},
		{
			line: "NINETY_SIX_BYTES: 0x414141414141414141414141414141414141414141414141414141414141414141414141" +
				"4141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141" +
				"41414141414141414141414141",
			wanted: "NINETY_SIX_BYTES: \n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n" +
				"- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n" +
				"- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n" +
				"- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n" +
				"- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n" +
				"- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n" +
				"- 65\n- 65\n- 65\n- 65\n- 65\n- 65\n",
		},
	}
	for _, line := range testLines {
		parts := params.ReplaceHexStringWithYAMLFormat(line.line)
		res := strings.Join(parts, "\n")

		if res != line.wanted {
			t.Errorf("expected conversion to be: %v got: %v", line.wanted, res)
		}
	}
}

func TestConfigParityYaml(t *testing.T) {
	params.SetupTestConfigCleanup(t)
	testDir := bazel.TestTmpDir()
	yamlDir := filepath.Join(testDir, "config.yaml")

	testCfg := params.E2ETestConfig()
	yamlObj := params.ConfigToYaml(testCfg)
	assert.NoError(t, file.WriteFile(yamlDir, yamlObj))

	params.LoadChainConfigFile(yamlDir, params.E2ETestConfig().Copy())
	assert.DeepEqual(t, params.BeaconConfig(), testCfg)
}

// configFilePath sets the proper config and returns the relevant
// config file path from eth2-spec-tests directory.
func configFilePath(t *testing.T, config string) string {
	filepath, err := bazel.Runfile("external/consensus_spec")
	require.NoError(t, err)
	configFilePath := path.Join(filepath, "configs", config+".yaml")
	return configFilePath
}

// presetsFilePath returns the relevant preset file paths from eth2-spec-tests
// directory. This method returns a preset file path for each hard fork or
// major network upgrade, in order.
func presetsFilePath(t *testing.T, config string) []string {
	filepath, err := bazel.Runfile("external/consensus_spec")
	require.NoError(t, err)
	return []string{
		path.Join(filepath, "presets", config, "phase0.yaml"),
		path.Join(filepath, "presets", config, "altair.yaml"),
	}
}

func fieldsFromYamls(t *testing.T, fps []string) []string {
	var keys []string
	for _, fp := range fps {
		yamlFile, err := ioutil.ReadFile(fp)
		require.NoError(t, err)
		m := make(map[string]interface{})
		require.NoError(t, yaml.Unmarshal(yamlFile, &m))

		for k := range m {
			keys = append(keys, k)
		}

		if len(keys) == 0 {
			t.Errorf("No fields loaded from yaml file %s", fp)
		}
	}

	return keys
}

func assertYamlFieldsMatch(t *testing.T, name string, fields []string, c1, c2 *params.BeaconChainConfig) {
	// Ensure all fields from the yaml file exist, were set, and correctly match the expected value.
	ft1 := reflect.TypeOf(*c1)
	for _, field := range fields {
		var found bool
		for i := 0; i < ft1.NumField(); i++ {
			v, ok := ft1.Field(i).Tag.Lookup("yaml")
			if ok && v == field {
				if isPlaceholderField(v) {
					// If you see this error, remove the field from placeholderFields.
					t.Errorf("beacon config has a placeholder field defined, remove %s from the placeholder fields variable", v)
					continue
				}
				found = true
				v1 := reflect.ValueOf(*c1).Field(i).Interface()
				v2 := reflect.ValueOf(*c2).Field(i).Interface()
				if reflect.ValueOf(v1).Kind() == reflect.Slice {
					assert.DeepEqual(t, v1, v2, "%s: %s", name, field)
				} else {
					assert.Equal(t, v1, v2, "%s: %s", name, field)
				}
				break
			}
		}
		if !found && !isPlaceholderField(field) { // Ignore placeholder fields
			t.Errorf("No struct tag found `yaml:%s`", field)
		}
	}
}

func isPlaceholderField(field string) bool {
	for _, f := range placeholderFields {
		if f == field {
			return true
		}
	}
	return false
}

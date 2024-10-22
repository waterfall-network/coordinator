package interop_test

import (
	"io/ioutil"
	"testing"

	"github.com/bazelbuild/rules_go/go/tools/bazel"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/runtime/interop"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/assert"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/require"
	"gitlab.waterfall.network/waterfall/protocol/gwat/common/hexutil"
	"gopkg.in/yaml.v2"
)

type TestCase struct {
	Privkey string `yaml:"privkey"`
}

type KeyTest struct {
	TestCases []*TestCase `yaml:"test_cases"`
}

func TestKeyGenerator(t *testing.T) {
	path, err := bazel.Runfile("keygen_test_vector.yaml")
	require.NoError(t, err)
	file, err := ioutil.ReadFile(path)
	require.NoError(t, err)
	testCases := &KeyTest{}
	require.NoError(t, yaml.Unmarshal(file, testCases))
	priv, _, err := interop.DeterministicallyGenerateKeys(0, 1000)
	require.NoError(t, err)
	// cross-check with the first 1000 keys generated from the python spec
	for i, key := range priv {
		hexKey := testCases.TestCases[i].Privkey
		nKey, err := hexutil.Decode("0x" + hexKey)
		if err != nil {
			t.Error(err)
			continue
		}
		assert.DeepEqual(t, key.Marshal(), nKey)
	}
}

package bls

import (
	"encoding/hex"
	"testing"

	"github.com/ghodss/yaml"
	"github.com/waterfall-foundation/coordinator/crypto/bls"
	"github.com/waterfall-foundation/coordinator/crypto/bls/common"
	"github.com/waterfall-foundation/coordinator/testing/bls/utils"
	"github.com/waterfall-foundation/coordinator/testing/require"
)

func TestAggregate(t *testing.T) {
	t.Run("blst", testAggregate)
}

func testAggregate(t *testing.T) {
	fNames, fContent := utils.RetrieveFiles("aggregate", t)
	for i, folder := range fNames {
		t.Run(folder, func(t *testing.T) {
			test := &AggregateTest{}
			require.NoError(t, yaml.Unmarshal(fContent[i], test))
			var sigs []common.Signature
			for _, s := range test.Input {
				sigBytes, err := hex.DecodeString(s[2:])
				require.NoError(t, err)
				sig, err := bls.SignatureFromBytes(sigBytes)
				require.NoError(t, err)
				sigs = append(sigs, sig)
			}
			if len(test.Input) == 0 {
				if test.Output != "" {
					t.Fatalf("Output Aggregate is not of zero length:Output %s", test.Output)
				}
				return
			}
			sig := bls.AggregateSignatures(sigs)
			outputBytes, err := hex.DecodeString(test.Output[2:])
			require.NoError(t, err)
			require.DeepEqual(t, outputBytes, sig.Marshal())
		})
	}
}

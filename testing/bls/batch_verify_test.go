package bls

import (
	"encoding/hex"
	"errors"
	"testing"

	"github.com/ghodss/yaml"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/crypto/bls"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/crypto/bls/common"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/encoding/bytesutil"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/bls/utils"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/require"
)

func TestBatchVerify(t *testing.T) {
	t.Run("blst", testBatchVerify)
}

func testBatchVerify(t *testing.T) {
	fNames, fContent := utils.RetrieveFiles("batch_verify", t)

	for i, file := range fNames {
		t.Run(file, func(t *testing.T) {
			test := &BatchVerifyTest{}
			require.NoError(t, yaml.Unmarshal(fContent[i], test))

			pubkeys := make([]common.PublicKey, len(test.Input.Pubkeys))
			messages := make([][32]byte, len(test.Input.Messages))
			signatures := make([][]byte, len(test.Input.Signatures))
			for j, raw := range test.Input.Pubkeys {
				pkBytes, err := hex.DecodeString(raw[2:])
				require.NoError(t, err)
				pk, err := bls.PublicKeyFromBytes(pkBytes)
				if err != nil {
					if test.Output == false && errors.Is(err, common.ErrInfinitePubKey) {
						return
					}
					t.Fatalf("cannot unmarshal pubkey: %v", err)
				}
				pubkeys[j] = pk
			}
			for j, raw := range test.Input.Messages {
				msgBytes, err := hex.DecodeString(raw[2:])
				require.NoError(t, err)
				messages[j] = bytesutil.ToBytes32(msgBytes)
			}
			for j, raw := range test.Input.Signatures {
				sigBytes, err := hex.DecodeString(raw[2:])
				require.NoError(t, err)
				signatures[j] = sigBytes
			}

			verified, err := bls.VerifyMultipleSignatures(signatures, messages, pubkeys)
			require.NoError(t, err)
			if verified != test.Output {
				t.Fatalf("Signature does not match the expected verification output. "+
					"Expected %#v but received %#v for test case %d", test.Output, verified, i)
			}
			t.Log("Success")
		})
	}
}

package bls

import (
	"bytes"
	"encoding/hex"
	"errors"
	"path"
	"testing"

	"github.com/ghodss/yaml"
	"github.com/waterfall-foundation/coordinator/crypto/bls"
	"github.com/waterfall-foundation/coordinator/crypto/bls/common"
	"github.com/waterfall-foundation/coordinator/testing/require"
	"github.com/waterfall-foundation/coordinator/testing/spectest/utils"
	"github.com/waterfall-foundation/coordinator/testing/util"
)

func TestSign(t *testing.T) {
	t.Run("blst", testSign)
}

func testSign(t *testing.T) {
	testFolders, testFolderPath := utils.TestFolders(t, "general", "phase0", "bls/sign/small")

	for i, folder := range testFolders {
		t.Run(folder.Name(), func(t *testing.T) {
			file, err := util.BazelFileBytes(path.Join(testFolderPath, folder.Name(), "data.yaml"))
			require.NoError(t, err)
			test := &SignMsgTest{}
			require.NoError(t, yaml.Unmarshal(file, test))
			pkBytes, err := hex.DecodeString(test.Input.Privkey[2:])
			require.NoError(t, err)
			sk, err := bls.SecretKeyFromBytes(pkBytes)
			if err != nil {
				if test.Output == "" &&
					(errors.Is(err, common.ErrZeroKey) || errors.Is(err, common.ErrSecretUnmarshal)) {
					return
				}
				t.Fatalf("cannot unmarshal secret key: %v", err)
			}
			msgBytes, err := hex.DecodeString(test.Input.Message[2:])
			require.NoError(t, err)
			sig := sk.Sign(msgBytes)

			if !sig.Verify(sk.PublicKey(), msgBytes) {
				t.Fatal("could not verify signature")
			}

			outputBytes, err := hex.DecodeString(test.Output[2:])
			require.NoError(t, err)

			if !bytes.Equal(outputBytes, sig.Marshal()) {
				t.Fatalf("Test Case %d: Signature does not match the expected output. "+
					"Expected %#x but received %#x", i, outputBytes, sig.Marshal())
			}
			t.Log("Success")
		})
	}
}

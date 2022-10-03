package wallet_test

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/waterfall-foundation/coordinator/config/params"
	"github.com/waterfall-foundation/coordinator/testing/assert"
	"github.com/waterfall-foundation/coordinator/testing/require"
	"github.com/waterfall-foundation/coordinator/validator/accounts/iface"
	"github.com/waterfall-foundation/coordinator/validator/accounts/wallet"
	remote_web3signer "github.com/waterfall-foundation/coordinator/validator/keymanager/remote-web3signer"
	"github.com/waterfall-foundation/gwat/common/hexutil"
)

func init() {
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetOutput(ioutil.Discard)
}

func Test_Exists_RandomFiles(t *testing.T) {
	path := filepath.Join(t.TempDir(), "wallet")

	exists, err := wallet.Exists(path)
	require.Equal(t, false, exists)
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(path+"/direct", params.BeaconIoConfig().ReadWriteExecutePermissions), "Failed to create directory")

	exists, err = wallet.Exists(path)
	require.NoError(t, err)
	require.Equal(t, true, exists)
}

func Test_IsValid_RandomFiles(t *testing.T) {
	path := filepath.Join(t.TempDir(), "wallet")
	valid, err := wallet.IsValid(path)
	require.NoError(t, err)
	require.Equal(t, false, valid)

	require.NoError(t, os.MkdirAll(path, params.BeaconIoConfig().ReadWriteExecutePermissions), "Failed to create directory")

	valid, err = wallet.IsValid(path)
	require.ErrorContains(t, "no wallet found", err)
	require.Equal(t, false, valid)

	walletDir := filepath.Join(path, "direct")
	require.NoError(t, os.MkdirAll(walletDir, params.BeaconIoConfig().ReadWriteExecutePermissions), "Failed to create directory")

	valid, err = wallet.IsValid(path)
	require.NoError(t, err)
	require.Equal(t, true, valid)
}

func TestWallet_InitializeKeymanager_web3Signer_HappyPath(t *testing.T) {
	w := wallet.NewWalletForWeb3Signer()
	ctx := context.Background()
	root, err := hexutil.Decode("0x270d43e74ce340de4bca2b1936beca0f4f5408d9e78aec4850920baf659d5b69")
	require.NoError(t, err)
	config := iface.InitKeymanagerConfig{
		ListenForChanges: false,
		Web3SignerConfig: &remote_web3signer.SetupConfig{
			BaseEndpoint:          "http://localhost:8545",
			GenesisValidatorsRoot: root,
			PublicKeysURL:         "http://localhost:8545/public_keys",
		},
	}
	km, err := w.InitializeKeymanager(ctx, config)
	require.NoError(t, err)
	assert.NotNil(t, km)
}

func TestWallet_InitializeKeymanager_web3Signer_nilConfig(t *testing.T) {
	w := wallet.NewWalletForWeb3Signer()
	ctx := context.Background()
	config := iface.InitKeymanagerConfig{
		ListenForChanges: false,
		Web3SignerConfig: nil,
	}
	km, err := w.InitializeKeymanager(ctx, config)
	assert.NotNil(t, err)
	assert.Equal(t, nil, km)
}

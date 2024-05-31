package accounts

import (
	"testing"

	"github.com/sirupsen/logrus/hooks/test"
	fieldparams "gitlab.waterfall.network/waterfall/protocol/coordinator/config/fieldparams"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/encoding/bytesutil"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/assert"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/require"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/validator/accounts/wallet"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/validator/keymanager"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/validator/keymanager/local"
	"google.golang.org/grpc/metadata"
)

func TestPrepareWallet_EmptyWalletReturnsError(t *testing.T) {
	local.ResetCaches()
	walletDir, _, passwordFilePath := setupWalletAndPasswordsDir(t)
	cliCtx := setupWalletCtx(t, &testWalletConfig{
		walletDir:           walletDir,
		keymanagerKind:      keymanager.Local,
		walletPasswordFile:  passwordFilePath,
		accountPasswordFile: passwordFilePath,
	})
	_, err := CreateWalletWithKeymanager(cliCtx.Context, &CreateWalletConfig{
		WalletCfg: &wallet.Config{
			WalletDir:      walletDir,
			KeymanagerKind: keymanager.Local,
			WalletPassword: password,
		},
	})
	require.NoError(t, err)
	_, _, err = prepareWallet(cliCtx)
	assert.ErrorContains(t, "wallet is empty", err)
}

func TestPrepareClients_AddsGRPCHeaders(t *testing.T) {
	local.ResetCaches()
	walletDir, _, passwordFilePath := setupWalletAndPasswordsDir(t)
	cliCtx := setupWalletCtx(t, &testWalletConfig{
		walletDir:           walletDir,
		keymanagerKind:      keymanager.Local,
		walletPasswordFile:  passwordFilePath,
		accountPasswordFile: passwordFilePath,
		grpcHeaders:         "Authorization=Basic some-token,Some-Other-Header=some-value",
	})
	_, err := CreateWalletWithKeymanager(cliCtx.Context, &CreateWalletConfig{
		WalletCfg: &wallet.Config{
			WalletDir:      walletDir,
			KeymanagerKind: keymanager.Local,
			WalletPassword: password,
		},
	})
	require.NoError(t, err)
	_, _, err = prepareClients(cliCtx)
	require.NoError(t, err)
	md, _ := metadata.FromOutgoingContext(cliCtx.Context)
	assert.Equal(t, "Basic some-token", md.Get("Authorization")[0])
	assert.Equal(t, "some-value", md.Get("Some-Other-Header")[0])
}

func TestDisplayExitInfo(t *testing.T) {
	logHook := test.NewGlobal()
	key := []byte("0x123456")
	displayExitInfo([][]byte{key}, []string{string(key)})
	assert.LogsContain(t, logHook, "beaconcha.in/validator/3078313233343536")
}

func TestDisplayExitInfo_NoKeys(t *testing.T) {
	logHook := test.NewGlobal()
	displayExitInfo([][]byte{}, []string{})
	assert.LogsContain(t, logHook, "No successful voluntary exits")
}

func TestPrepareAllKeys(t *testing.T) {
	key1 := bytesutil.ToBytes48([]byte("key1"))
	key2 := bytesutil.ToBytes48([]byte("key2"))
	raw, formatted := prepareAllKeys([][fieldparams.BLSPubkeyLength]byte{key1, key2})
	require.Equal(t, 2, len(raw))
	require.Equal(t, 2, len(formatted))
	assert.DeepEqual(t, bytesutil.ToBytes48([]byte{107, 101, 121, 49}), bytesutil.ToBytes48(raw[0]))
	assert.DeepEqual(t, bytesutil.ToBytes48([]byte{107, 101, 121, 50}), bytesutil.ToBytes48(raw[1]))
	assert.Equal(t, "0x6b6579310000", formatted[0])
	assert.Equal(t, "0x6b6579320000", formatted[1])
}

package node

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	logTest "github.com/sirupsen/logrus/hooks/test"
	"github.com/urfave/cli/v2"
	statefeed "gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/core/feed/state"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/powchain"
	mockPOW "gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/powchain/testing"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/cmd"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/require"
)

// Ensure BeaconNode implements interfaces.
var _ statefeed.Notifier = (*BeaconNode)(nil)

// TestClearDB tests clearing the database
func TestClearDB(t *testing.T) {
	hook := logTest.NewGlobal()
	srv, endpoint, err := mockPOW.SetupRPCServer()
	require.NoError(t, err)
	t.Cleanup(func() {
		srv.Stop()
	})

	tmp := filepath.Join(t.TempDir(), "datadirtest")

	app := cli.App{}
	set := flag.NewFlagSet("test", 0)
	set.String("datadir", tmp, "node data directory")
	set.Bool(cmd.ForceClearDB.Name, true, "force clear db")
	set.String("genesis-state", "testing/testdata/genesis.ssz", "")

	context := cli.NewContext(&app, set, nil)
	_, err = New(context, WithPowchainFlagOptions([]powchain.Option{
		powchain.WithHttpEndpoints([]string{endpoint}),
	}))
	require.NoError(t, err)
	require.LogsContain(t, hook, "Removing database")
	require.NoError(t, os.RemoveAll(tmp))
}

package node

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	logTest "github.com/sirupsen/logrus/hooks/test"
	"github.com/urfave/cli/v2"
	statefeed "github.com/waterfall-foundation/coordinator/beacon-chain/core/feed/state"
	"github.com/waterfall-foundation/coordinator/beacon-chain/powchain"
	mockPOW "github.com/waterfall-foundation/coordinator/beacon-chain/powchain/testing"
	"github.com/waterfall-foundation/coordinator/cmd"
	"github.com/waterfall-foundation/coordinator/testing/require"
)

// Ensure BeaconNode implements interfaces.
var _ statefeed.Notifier = (*BeaconNode)(nil)

// Test that beacon chain node can close.
func TestNodeClose_OK(t *testing.T) {
	hook := logTest.NewGlobal()

	tmp := fmt.Sprintf("%s/datadirtest2", t.TempDir())

	app := cli.App{}
	set := flag.NewFlagSet("test", 0)
	set.Bool("test-skip-pow", true, "skip pow dial")
	set.String("datadir", tmp, "node data directory")
	set.String("p2p-encoding", "ssz", "p2p encoding scheme")
	set.Bool("demo-config", true, "demo configuration")
	set.String("deposit-contract", "0x0000000000000000000000000000000000000000", "deposit contract address")

	context := cli.NewContext(&app, set, nil)

	node, err := New(context)
	require.NoError(t, err)

	node.Close()

	require.LogsContain(t, hook, "Stopping beacon node")
	require.NoError(t, os.RemoveAll(tmp))
}

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

	context := cli.NewContext(&app, set, nil)
	_, err = New(context, WithPowchainFlagOptions([]powchain.Option{
		powchain.WithHttpEndpoints([]string{endpoint}),
	}))
	require.NoError(t, err)

	require.LogsContain(t, hook, "Removing database")
	require.NoError(t, os.RemoveAll(tmp))
}

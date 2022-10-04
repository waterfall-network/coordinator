package kv

import (
	"context"
	"testing"

	"github.com/waterfall-foundation/coordinator/beacon-chain/state/genesis"
	"github.com/waterfall-foundation/coordinator/config/params"
	"github.com/waterfall-foundation/coordinator/proto/prysm/v1alpha1/wrapper"
	"github.com/waterfall-foundation/coordinator/testing/require"
	"github.com/waterfall-foundation/coordinator/testing/util"
)

func TestSaveOrigin(t *testing.T) {
	// Embedded Genesis works with Mainnet config
	params.SetupTestConfigCleanup(t)
	cfg := params.BeaconConfig()
	cfg.ConfigName = params.ConfigNames[params.Mainnet]
	params.OverrideBeaconConfig(cfg)

	ctx := context.Background()
	db := setupDB(t)

	st, err := genesis.State(params.Mainnet.String())
	require.NoError(t, err)

	sb, err := st.MarshalSSZ()
	require.NoError(t, err)
	require.NoError(t, db.LoadGenesis(ctx, sb))

	// this is necessary for mainnet, because LoadGenesis is short-circuited by the embedded state,
	// so the genesis root key is never written to the db.
	require.NoError(t, db.EnsureEmbeddedGenesis(ctx))

	cst, err := util.NewBeaconState()
	require.NoError(t, err)
	csb, err := cst.MarshalSSZ()
	require.NoError(t, err)
	cb := util.NewBeaconBlock()
	scb, err := wrapper.WrappedSignedBeaconBlock(cb)
	require.NoError(t, err)
	cbb, err := scb.MarshalSSZ()
	require.NoError(t, err)
	require.NoError(t, db.SaveOrigin(ctx, csb, cbb))
}
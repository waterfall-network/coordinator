package kv

import (
	"context"
	"testing"

	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/db/iface"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/config/params"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/assert"
)

func TestStore_SaveGenesisData(t *testing.T) {
	ctx := context.Background()
	db := setupDB(t)

	gs, err := NewBeaconState()
	assert.NoError(t, err)

	assert.NoError(t, db.SaveGenesisData(ctx, gs))

	testGenesisDataSaved(t, db)
}

func testGenesisDataSaved(t *testing.T, db iface.Database) {
	ctx := context.Background()

	gb, err := db.GenesisBlock(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, gb)

	gbHTR, err := gb.Block().HashTreeRoot()
	assert.NoError(t, err)

	gss, err := db.StateSummary(ctx, gbHTR)
	assert.NoError(t, err)
	assert.NotNil(t, gss)

	head, err := db.HeadBlock(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, head)

	headHTR, err := head.Block().HashTreeRoot()
	assert.NoError(t, err)
	assert.Equal(t, gbHTR, headHTR, "head block does not match genesis block")
}

func TestEnsureEmbeddedGenesis(t *testing.T) {
	// Embedded Genesis works with Mainnet config
	params.SetupTestConfigCleanup(t)
	cfg := params.BeaconConfig()
	cfg.ConfigName = params.ConfigNames[params.Mainnet]
	params.OverrideBeaconConfig(cfg)

	ctx := context.Background()
	db := setupDB(t)

	db.genesisSszPath = "beacon-chain/db/kv/testdata/mainnet.genesis.ssz"

	gs, err := NewBeaconState()
	assert.NoError(t, err)

	assert.NoError(t, db.SaveGenesisData(context.Background(), gs))

	gs, err = db.GenesisState(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, gs, "an embedded genesis state does not exist")

	assert.NoError(t, db.EnsureEmbeddedGenesis(ctx))

	gb, err := db.GenesisBlock(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, gb)

	testGenesisDataSaved(t, db)
}

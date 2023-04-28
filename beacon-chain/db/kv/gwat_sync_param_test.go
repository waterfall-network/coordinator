package kv

import (
	"bytes"
	"context"
	"testing"

	types "github.com/prysmaticlabs/eth2-types"
	ethpb "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1/wrapper"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/assert"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/require"
	gwatCommon "gitlab.waterfall.network/waterfall/protocol/gwat/common"
	gwatTypes "gitlab.waterfall.network/waterfall/protocol/gwat/core/types"
)

var gwatSyncParamTests = []struct {
	name             string
	newGwatSyncParam func() wrapper.GwatSyncParam
}{
	{
		name: "gwat sync param",
		newGwatSyncParam: func() wrapper.GwatSyncParam {
			return *NewGwatSyncParam()
		},
	},
}

func TestStore_GwatSyncParamCRUD(t *testing.T) {
	ctx := context.Background()

	for _, tt := range gwatSyncParamTests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupDB(t)

			gsp := tt.newGwatSyncParam()
			epoch := types.Epoch(gsp.Epoch())
			retrievedGsp, err := db.GwatSyncParam(ctx, epoch)
			require.NoError(t, err)
			var nilGsp *wrapper.GwatSyncParam
			assert.Equal(t, nilGsp, retrievedGsp, "Expected nil GwatSyncParam")
			require.NoError(t, db.SaveGwatSyncParam(ctx, gsp))
			retrievedGsp, err = db.GwatSyncParam(ctx, epoch)
			require.NoError(t, err)

			gsp_m, err := gsp.Bytes()
			require.NoError(t, err)
			retrievedGsp_m, err := retrievedGsp.Bytes()
			require.NoError(t, err)
			assert.Equal(t, true, bytes.Equal(gsp_m, retrievedGsp_m), "Wanted: %v, received: %v", gsp, retrievedGsp)
		})
	}
}

// NewGwatSyncParam creates a beacon block with minimum marshalable fields.
func NewGwatSyncParam() *wrapper.GwatSyncParam {
	baseSpine := gwatCommon.HexToHash("0x12380221e25ac8aedaa824fa4a456072dbe48f3421794edafcaed1f57f9aab59")
	param := &gwatTypes.FinalizationParams{
		Spines: gwatCommon.HashArray{
			gwatCommon.HexToHash("0x12380221e25ac8aedaa824fa4a456072dbe48f3421794edafcaed1f57f9aab59"),
			gwatCommon.HexToHash("0x4646e30459df69be4de0042179caa70d486a992ad10d30ed6a7d3737f054411f"),
			gwatCommon.HexToHash("0x8b67feb3b50025fedc5e62be44679b460d336eec59d753547272522ddab09db2"),
			gwatCommon.HexToHash("0xb48e9c41a7ee369d6437367b65ebd6e149e46611808e4531288eb681ee98ae12"),
		},
		BaseSpine: &baseSpine,
		Checkpoint: &gwatTypes.Checkpoint{
			Epoch:    123,
			FinEpoch: 128,
			Root:     gwatCommon.HexToHash("0x317b57e99bda9cba1cdeb257876dd1d31cadba4b8356f2f1925c2bce08d71377"),
			Spine:    gwatCommon.HexToHash("0xedefc29821526716c997dc29fdf9d3bbb549c3fd883fdd68ab2bcee35c212965"),
		},
		ValSyncData: []*gwatTypes.ValidatorSync{},
	}
	cp := &ethpb.Checkpoint{
		Epoch: 124,
		Root:  gwatCommon.HexToHash("0x317b57e99bda9cba1cdeb257876dd1d31cadba4b8356f2f1925c2bce08d71377").Bytes(),
	}
	return wrapper.NewGwatSyncParam(cp, param, 128)
}

package kv

import (
	"bytes"
	"context"
	"testing"

	"gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1/wrapper"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/assert"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/require"
	gwatCommon "gitlab.waterfall.network/waterfall/protocol/gwat/common"
)

var paramSpinesTests = []struct {
	name           string
	newSpinesParam func() wrapper.Spines
}{
	{
		name: "gwat sync param",
		newSpinesParam: func() wrapper.Spines {
			return NewSpinesParam()
		},
	},
}

func TestStore_GwatSyncParamCRUD(t *testing.T) {
	ctx := context.Background()

	for _, tt := range paramSpinesTests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupDB(t)

			spines := tt.newSpinesParam()
			key := spines.Key()
			retrievedSpines, err := db.ReadSpines(ctx, key)
			require.NoError(t, err)
			var nilGsp wrapper.Spines
			assert.Equal(t, nilGsp, retrievedSpines, "Expected nil ReadSpines")
			require.NoError(t, db.WriteSpines(ctx, spines))
			retrievedSpines, err = db.ReadSpines(ctx, key)
			require.NoError(t, err)

			assert.Equal(t, true, bytes.Equal(spines, retrievedSpines), "Wanted: %v, received: %v", spines, retrievedSpines)
		})
	}
}

func NewSpinesParam() wrapper.Spines {
	return gwatCommon.HashArray{
		gwatCommon.HexToHash("0x12380221e25ac8aedaa824fa4a456072dbe48f3421794edafcaed1f57f9aab59"),
		gwatCommon.HexToHash("0x4646e30459df69be4de0042179caa70d486a992ad10d30ed6a7d3737f054411f"),
		gwatCommon.HexToHash("0x8b67feb3b50025fedc5e62be44679b460d336eec59d753547272522ddab09db2"),
		gwatCommon.HexToHash("0xb48e9c41a7ee369d6437367b65ebd6e149e46611808e4531288eb681ee98ae12"),
		gwatCommon.HexToHash("0x317b57e99bda9cba1cdeb257876dd1d31cadba4b8356f2f1925c2bce08d71377"),
		gwatCommon.HexToHash("0xedefc29821526716c997dc29fdf9d3bbb549c3fd883fdd68ab2bcee35c212965"),
	}.ToBytes()
}

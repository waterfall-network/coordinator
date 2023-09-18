package kv

import (
	"context"
	"fmt"
	"testing"

	"gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1/wrapper"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/assert"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/require"
	gwatCommon "gitlab.waterfall.network/waterfall/protocol/gwat/common"
)

func TestStore_GwatSyncParamCRUD(t *testing.T) {
	ctx := context.Background()

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

	for _, tt := range paramSpinesTests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupDB(t)

			spines := tt.newSpinesParam()
			key := spines.Key()
			retrievedSpines, err := db.ReadSpines(ctx, key)
			require.NoError(t, err)
			var nilGsp wrapper.Spines
			assert.Equal(t, fmt.Sprintf("%v", nilGsp), fmt.Sprintf("%v", retrievedSpines), "Expected nil ReadSpines")
			wrKey, err := db.WriteSpines(ctx, spines)
			require.NoError(t, err)
			assert.Equal(t, key, wrKey, "Wanted: %#x, received: %#x", key, wrKey)
			// check in cache.
			cached, ok := db.spinesCache.Get(key)
			assert.Equal(t, true, ok, "Wanted: %v, received: %v", true, ok)
			assert.Equal(t, fmt.Sprintf("%#x", spines), fmt.Sprintf("%#x", cached.(wrapper.Spines)), "Wanted: %#x, received: %#x", spines, cached)

			//no cache
			db.spinesCache.Remove(key)
			retrievedSpines, err = db.ReadSpines(ctx, key)
			require.NoError(t, err)
			assert.Equal(t, fmt.Sprintf("%#x", spines), fmt.Sprintf("%#x", retrievedSpines), "Wanted: %#x, received: %#x", spines, retrievedSpines)

			// check cache is updated.
			cached, ok = db.spinesCache.Get(key)
			assert.Equal(t, true, ok, "Wanted: %v, received: %v", true, ok)
			assert.Equal(t, fmt.Sprintf("%#x", spines), fmt.Sprintf("%#x", cached.(wrapper.Spines)), "Wanted: %#x, received: %#x", spines, cached)

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

package powchain

import (
	"math/big"
	"testing"

	"github.com/waterfall-foundation/coordinator/beacon-chain/powchain/types"
	"github.com/waterfall-foundation/coordinator/testing/assert"
	"github.com/waterfall-foundation/coordinator/testing/require"
	"gitlab.waterfall.network/waterfall/protocol/gwat/common"
	gethTypes "gitlab.waterfall.network/waterfall/protocol/gwat/core/types"
)

func TestHashKeyFn_OK(t *testing.T) {
	hInfo := &types.HeaderInfo{
		Hash: common.HexToHash("0x0123456"),
	}

	key, err := hashKeyFn(hInfo)
	require.NoError(t, err)
	assert.Equal(t, hInfo.Hash.Hex(), key)
}

func TestHashKeyFn_InvalidObj(t *testing.T) {
	_, err := hashKeyFn("bad")
	assert.Equal(t, ErrNotAHeaderInfo, err)
}

func TestHeightKeyFn_OK(t *testing.T) {
	hInfo := &types.HeaderInfo{
		Number: big.NewInt(555),
	}

	key, err := heightKeyFn(hInfo)
	require.NoError(t, err)
	assert.Equal(t, hInfo.Number.String(), key)
}

func TestHeightKeyFn_InvalidObj(t *testing.T) {
	_, err := heightKeyFn("bad")
	assert.Equal(t, ErrNotAHeaderInfo, err)
}

func TestBlockCache_byHash(t *testing.T) {
	cache := newHeaderCache()
	nr_0 := uint64(55)
	header := &gethTypes.Header{
		ParentHashes: common.HashArray{common.HexToHash("0x12345")},
		Number:       &nr_0,
	}

	exists, _, err := cache.HeaderInfoByHash(header.Hash())
	require.NoError(t, err)
	assert.Equal(t, false, exists, "Expected block info not to exist in empty cache")

	err = cache.AddHeader(header)
	require.NoError(t, err)

	exists, fetchedInfo, err := cache.HeaderInfoByHash(header.Hash())
	require.NoError(t, err)
	assert.Equal(t, true, exists, "Expected headerInfo to exist")
	assert.Equal(t, 0, fetchedInfo.Number.Cmp(new(big.Int).SetUint64(header.Nr())), "Expected fetched info number to be equal")
	assert.Equal(t, header.Hash(), fetchedInfo.Hash, "Expected hash to be equal")

}

func TestBlockCache_byHeight(t *testing.T) {
	cache := newHeaderCache()
	nr_0 := uint64(55)
	header := &gethTypes.Header{
		ParentHashes: common.HashArray{common.HexToHash("0x12345")},
		Number:       &nr_0,
	}

	exists, _, err := cache.HeaderInfoByHeight(new(big.Int).SetUint64(header.Nr()))
	require.NoError(t, err)
	assert.Equal(t, false, exists, "Expected block info not to exist in empty cache")

	err = cache.AddHeader(header)
	require.NoError(t, err)

	exists, fetchedInfo, err := cache.HeaderInfoByHeight(new(big.Int).SetUint64(header.Nr()))
	require.NoError(t, err)
	assert.Equal(t, true, exists, "Expected headerInfo to exist")

	assert.Equal(t, 0, fetchedInfo.Number.Cmp(new(big.Int).SetUint64(header.Nr())), "Expected fetched info number to be equal")
	assert.Equal(t, header.Hash(), fetchedInfo.Hash, "Expected hash to be equal")

}

func TestBlockCache_maxSize(t *testing.T) {
	cache := newHeaderCache()

	for i := int64(0); i < int64(maxCacheSize+10); i++ {
		nr := uint64(i)
		header := &gethTypes.Header{
			Number: &nr,
			Height: nr,
		}
		err := cache.AddHeader(header)
		require.NoError(t, err)

	}

	assert.Equal(t, int(maxCacheSize), len(cache.hashCache.ListKeys()))
	assert.Equal(t, int(maxCacheSize), len(cache.heightCache.ListKeys()))
}

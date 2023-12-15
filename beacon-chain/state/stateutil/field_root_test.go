package stateutil

import (
	"fmt"
	"testing"

	"gitlab.waterfall.network/waterfall/protocol/coordinator/encoding/bytesutil"
	ethpb "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/assert"
	gwatCommon "gitlab.waterfall.network/waterfall/protocol/gwat/common"
)

func TestArraysTreeRoot_OnlyPowerOf2(t *testing.T) {
	_, err := arraysRoot([][]byte{}, 1)
	assert.NoError(t, err)
	_, err = arraysRoot([][]byte{}, 4)
	assert.NoError(t, err)
	_, err = arraysRoot([][]byte{}, 8)
	assert.NoError(t, err)
	_, err = arraysRoot([][]byte{}, 10)
	assert.ErrorContains(t, "hash layer is a non power of 2", err)
}

func TestArraysTreeRoot_ZeroLength(t *testing.T) {
	_, err := arraysRoot([][]byte{}, 0)
	assert.ErrorContains(t, "zero leaves provided", err)
}

func TestEth1DataRootWithHasher(t *testing.T) {
	lastFinHash := gwatCommon.HexToHash("0x5084316e3b55e27ea074588f3f1000ceff6e1c67d35e9e4eb14d6e5a7980426e")
	finHash := gwatCommon.HexToHash("0xffffffffffffffffffffffffffffffffffffffffffffffff0101010101010101")
	candidates := gwatCommon.HashArray{finHash}
	eth1Data := &ethpb.Eth1Data{
		DepositRoot:  bytesutil.PadTo([]byte("DepositRoot"), 32),
		DepositCount: 3,
		BlockHash:    lastFinHash.Bytes(),
		Candidates:   candidates.ToBytes(),
	}
	root, err := Eth1DataRootWithHasher(eth1Data)
	fmt.Printf("root=%x \n", root)

	assert.Equal(t, "5d5aed27b8e979737abbdc8fb09ead4cc8c41123f24433d2102faf7e213b461c", fmt.Sprintf("%x", root))
	assert.NoError(t, err)
}

func TestEth1DataRootWithHasher2(t *testing.T) {
	lastFinHash := gwatCommon.HexToHash("0x5084316e3b55e27ea074588f3f1000ceff6e1c67d35e9e4eb14d6e5a7980426e")
	finHash := gwatCommon.HexToHash("0xffffffffffffffffffffffffffffffffffffffffffffffff0101010101010101")
	candidates := gwatCommon.HashArray{
		finHash, finHash, finHash, finHash, finHash, finHash, finHash, finHash,
		finHash, finHash, finHash, finHash, finHash, finHash, finHash, finHash,
		finHash, finHash, finHash, finHash, finHash, finHash, finHash, finHash,
		finHash, finHash, finHash, finHash, finHash, finHash, finHash, finHash,
	}
	eth1Data := &ethpb.Eth1Data{
		DepositRoot:  bytesutil.PadTo([]byte("DepositRoot"), 32),
		DepositCount: 3,
		Candidates:   candidates.ToBytes(),
		BlockHash:    lastFinHash.Bytes(),
	}
	root, err := Eth1DataRootWithHasher(eth1Data)
	fmt.Printf("root=%x \n", root)

	assert.Equal(t, "16e6be592d535ea1ff39865aaab388393b31f4a7b8cb8ce4ada9920c16bbedd3", fmt.Sprintf("%x", root))
	assert.NoError(t, err)
}

func TestSpineDataRootWithHasher(t *testing.T) {
	spines := gwatCommon.HashArray{
		gwatCommon.HexToHash("0x1111111111111111111111111111111111111111111111111111111111111111"),
		gwatCommon.HexToHash("0x2222222222222222222222222222222222222222222222222222222222222222"),
	}
	prefix := gwatCommon.HashArray{
		gwatCommon.HexToHash("0x3333333333333333333333333333333333333333333333333333333333333333"),
		gwatCommon.HexToHash("0x4444444444444444444444444444444444444444444444444444444444444444"),
	}
	finalization := gwatCommon.HashArray{
		gwatCommon.HexToHash("0x5555555555555555555555555555555555555555555555555555555555555555"),
		gwatCommon.HexToHash("0x6666666666666666666666666666666666666666666666666666666666666666"),
	}
	cpFinalized := gwatCommon.HashArray{
		gwatCommon.HexToHash("0x5555555555555555555555555555555555555555555555555555555555555555"),
		gwatCommon.HexToHash("0x6666666666666666666666666666666666666666666666666666666666666666"),
	}

	parentSpines := []*ethpb.SpinesSeq{
		{Spines: spines.ToBytes()},
		{Spines: gwatCommon.HashArray{
			gwatCommon.HexToHash("0x1111111111111111111111111111111111111111111111111111111111111111"),
			gwatCommon.HexToHash("0x2222222222222222222222222222222222222222222222222222222222222222"),
			gwatCommon.HexToHash("0x7777777777777777777777777777777777777777777777777777777777777777"),
			gwatCommon.HexToHash("0xffffffffffffffffffffffffffffffffffffffffffffffff0101010101010101"),
		}.ToBytes(),
		},
	}

	spineData := &ethpb.SpineData{
		Spines:       spines.ToBytes(),
		Prefix:       prefix.ToBytes(),
		Finalization: finalization.ToBytes(),
		CpFinalized:  cpFinalized.ToBytes(),
		ParentSpines: parentSpines,
	}
	root, err := SpineDataRootWithHasher(spineData)
	fmt.Printf("root=%x \n", root)

	assert.Equal(t, "4bb30e34869e8dc9a662ba8bb91e2dfd6a2e21322e7cac9829ffe5818ae3424f", fmt.Sprintf("%x", root))
	assert.NoError(t, err)
}

func TestSpineDataRootWithHasher2(t *testing.T) {
	spines := gwatCommon.HashArray{
		gwatCommon.HexToHash("0x1111111111111111111111111111111111111111111111111111111111111111"),
		gwatCommon.HexToHash("0x2222222222222222222222222222222222222222222222222222222222222222"),
	}
	prefix := gwatCommon.HashArray{
		gwatCommon.HexToHash("0x3333333333333333333333333333333333333333333333333333333333333333"),
		gwatCommon.HexToHash("0x4444444444444444444444444444444444444444444444444444444444444444"),
	}
	finalization := gwatCommon.HashArray{
		gwatCommon.HexToHash("0x5555555555555555555555555555555555555555555555555555555555555555"),
		gwatCommon.HexToHash("0x6666666666666666666666666666666666666666666666666666666666666666"),
	}
	cpFinalized := gwatCommon.HashArray{
		gwatCommon.HexToHash("0x5555555555555555555555555555555555555555555555555555555555555555"),
		gwatCommon.HexToHash("0x6666666666666666666666666666666666666666666666666666666666666666"),
	}

	parentSpines := []*ethpb.SpinesSeq{
		{Spines: spines.ToBytes()},
		{Spines: gwatCommon.HashArray{
			gwatCommon.HexToHash("0x1111111111111111111111111111111111111111111111111111111111111111"),
			gwatCommon.HexToHash("0x2222222222222222222222222222222222222222222222222222222222222222"),
			gwatCommon.HexToHash("0x7777777777777777777777777777777777777777777777777777777777777777"),
			gwatCommon.HexToHash("0xffffffffffffffffffffffffffffffffffffffffffffffff0101010101010101"),
		}.ToBytes(),
		},
	}

	spineData := &ethpb.SpineData{
		Spines:       spines.ToBytes(),
		Prefix:       prefix.ToBytes(),
		Finalization: finalization.ToBytes(),
		CpFinalized:  cpFinalized.ToBytes(),
		ParentSpines: parentSpines,
	}
	root, err := SpineDataRootWithHasher(spineData)
	fmt.Printf("root=%x \n", root)

	assert.Equal(t, "4bb30e34869e8dc9a662ba8bb91e2dfd6a2e21322e7cac9829ffe5818ae3424f", fmt.Sprintf("%x", root))
	assert.NoError(t, err)
}

func TestBytesRoot_Hashes(t *testing.T) {
	spines := gwatCommon.HashArray{
		gwatCommon.HexToHash("0x1111111111111111111111111111111111111111111111111111111111111111"),
		gwatCommon.HexToHash("0x2222222222222222222222222222222222222222222222222222222222222222"),
		gwatCommon.HexToHash("0x3333333333333333333333333333333333333333333333333333333333333333"),
		gwatCommon.HexToHash("0x4444444444444444444444444444444444444444444444444444444444444444"),
		gwatCommon.HexToHash("0x5555555555555555555555555555555555555555555555555555555555555555"),
		gwatCommon.HexToHash("0x6666666666666666666666666666666666666666666666666666666666666666"),
	}

	root, err := BytesRoot(spines.ToBytes())
	fmt.Printf("root=%x \n", root)

	assert.Equal(t, "22fa0bfbc35e5e9339b82dc1d5e59502f8dc0849e534401d21f5fe2607b93f4a", fmt.Sprintf("%x", root))
	assert.NoError(t, err)
}

func TestBytesRoot_RandLen(t *testing.T) {
	bts := []byte{0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11}

	root, err := BytesRoot(bts)
	fmt.Printf("root=%x \n", root)

	assert.Equal(t, "92baee50d189868780bb856f3cdfbe9c06a8f36396f9d527cd7f7fa2230ec44a", fmt.Sprintf("%x", root))
	assert.NoError(t, err)
}

func TestBytesRoot_EmptyData(t *testing.T) {
	var bts []byte
	//bts = []byte{}

	root, err := BytesRoot(bts)
	fmt.Printf("root=%x \n", root)

	assert.Equal(t, "f5a5fd42d16a20302798ef6ed309979b43003d2320d9f0e8ea9831a92759fb4b", fmt.Sprintf("%x", root))
	assert.NoError(t, err)
}

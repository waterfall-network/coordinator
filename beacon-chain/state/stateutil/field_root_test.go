package stateutil

import (
	"fmt"
	"testing"

	"gitlab.waterfall.network/waterfall/protocol/coordinator/crypto/hash"
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
	hasher := hash.CustomSHA256Hasher()
	lastFinHash := gwatCommon.HexToHash("0x5084316e3b55e27ea074588f3f1000ceff6e1c67d35e9e4eb14d6e5a7980426e")
	finHash := gwatCommon.HexToHash("0xffffffffffffffffffffffffffffffffffffffffffffffff0101010101010101")
	candidates := gwatCommon.HashArray{finHash}
	eth1Data := &ethpb.Eth1Data{
		DepositRoot:  bytesutil.PadTo([]byte("DepositRoot"), 32),
		DepositCount: 3,
		BlockHash:    lastFinHash.Bytes(),
		Candidates:   candidates.ToBytes(),
	}
	root, err := Eth1DataRootWithHasher(hasher, eth1Data)
	fmt.Printf("root=%x \n", root)

	assert.Equal(t, "684026722548071805f95f9559255b9687addfa5227501fd6b8b74eae7ac2454", fmt.Sprintf("%x", root))
	assert.NoError(t, err)
}

func TestEth1DataRootWithHasher2(t *testing.T) {
	hasher := hash.CustomSHA256Hasher()
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
	root, err := Eth1DataRootWithHasher(hasher, eth1Data)
	fmt.Printf("root=%x \n", root)

	assert.Equal(t, "2c3bdd383d83fd7b6f11074a55eaedccdf45205045de027b4a1c67246b716267", fmt.Sprintf("%x", root))
	assert.NoError(t, err)
}

func TestSpineDataRootWithHasher(t *testing.T) {
	hasher := hash.CustomSHA256Hasher()
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

	parentSpines := []*ethpb.SpinesSeq{
		&ethpb.SpinesSeq{Spines: spines.ToBytes()},
		&ethpb.SpinesSeq{Spines: gwatCommon.HashArray{
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
		ParentSpines: parentSpines,
	}
	root, err := SpineDataRootWithHasher(hasher, spineData)
	fmt.Printf("root=%x \n", root)

	assert.Equal(t, "5d450e6b013c3f2b0902caffbd70220440e59663a987e093cfa94f30259c4781", fmt.Sprintf("%x", root))
	assert.NoError(t, err)
}

func TestSpineDataRootWithHasher2(t *testing.T) {
	hasher := hash.CustomSHA256Hasher()
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

	parentSpines := []*ethpb.SpinesSeq{
		&ethpb.SpinesSeq{Spines: spines.ToBytes()},
		&ethpb.SpinesSeq{Spines: gwatCommon.HashArray{
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
		ParentSpines: parentSpines,
	}
	root, err := SpineDataRootWithHasher(hasher, spineData)
	fmt.Printf("root=%x \n", root)

	assert.Equal(t, "5d450e6b013c3f2b0902caffbd70220440e59663a987e093cfa94f30259c4781", fmt.Sprintf("%x", root))
	assert.NoError(t, err)
}

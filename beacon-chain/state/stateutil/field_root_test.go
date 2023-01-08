package stateutil

import (
	"fmt"
	"testing"

	"github.com/waterfall-foundation/coordinator/crypto/hash"
	"github.com/waterfall-foundation/coordinator/encoding/bytesutil"
	ethpb "github.com/waterfall-foundation/coordinator/proto/prysm/v1alpha1"
	"github.com/waterfall-foundation/coordinator/testing/assert"
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

package stateutil

import (
	"github.com/pkg/errors"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/encoding/ssz"
	ethpb "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1"
)

// SpineDataRoot computes the HashTreeRoot Merkleization of
// a BeaconBlockHeader struct according to the eth2
// Simple Serialize specification.
func SpineDataRoot(hasher ssz.HashFn, spineData *ethpb.SpineData) ([32]byte, error) {
	if spineData == nil {
		return [32]byte{}, errors.New("nil spine data")
	}
	return SpineDataRootWithHasher(hasher, spineData)
}

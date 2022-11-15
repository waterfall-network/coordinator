package stateutil

import (
	"github.com/pkg/errors"
	"github.com/waterfall-foundation/coordinator/encoding/ssz"
	ethpb "github.com/waterfall-foundation/coordinator/proto/prysm/v1alpha1"
)

// BlockVotingRoot computes the HashTreeRoot Merkleization of
// a BeaconBlockHeader struct according to the BlockVoting
// Simple Serialize specification.
func BlockVotingRoot(hasher ssz.HashFn, blockVoting *ethpb.BlockVoting) ([32]byte, error) {
	if blockVoting == nil {
		return [32]byte{}, errors.New("nil blockVoting data")
	}
	return BlockVotingDataRootWithHasher(hasher, blockVoting)
}

// blockVotingsRoot computes the HashTreeRoot Merkleization of
// a list of BlockVoting structs according to the eth2
// Simple Serialize specification.
func blockVotingRoot(blockVotings []*ethpb.BlockVoting) ([32]byte, error) {
	return BlockVotingsRoot(blockVotings)
}

package stateutil

import (
	ethpb "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1"
)

// blockVotingsRoot computes the HashTreeRoot Merkleization of
// a list of BlockVoting structs according to the eth2
// Simple Serialize specification.
func blockVotingRoot(blockVotings []*ethpb.BlockVoting) ([32]byte, error) {
	return BlockVotingsRoot(blockVotings)
}

package store

import (
	"sync"

	types "github.com/prysmaticlabs/eth2-types"
	ethpb "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1"
)

// Store is defined in the fork choice consensus spec for tracking current time and various versions of checkpoints.
//
// Spec code:
// class Store(object):
//
//	time: uint64
//	genesis_time: uint64
//	justified_checkpoint: Checkpoint
//	finalized_checkpoint: Checkpoint
//	best_justified_checkpoint: Checkpoint
//	proposerBoostRoot: Root
type Store struct {
	lastEpoch            types.Epoch
	justifiedCheckpt     *ethpb.Checkpoint
	finalizedCheckpt     *ethpb.Checkpoint
	bestJustifiedCheckpt *ethpb.Checkpoint
	sync.RWMutex
	// These are not part of the consensus spec, but we do use them to return gRPC API requests.
	prevFinalizedCheckpt *ethpb.Checkpoint
	prevJustifiedCheckpt *ethpb.Checkpoint
}

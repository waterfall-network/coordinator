package store

import (
	ethpb "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1"
)

func New(justifiedCheckpt *ethpb.Checkpoint, finalizedCheckpt *ethpb.Checkpoint) *Store {
	return &Store{
		justifiedCheckpt:     justifiedCheckpt,
		prevJustifiedCheckpt: justifiedCheckpt,
		bestJustifiedCheckpt: justifiedCheckpt,
		finalizedCheckpt:     finalizedCheckpt,
		prevFinalizedCheckpt: finalizedCheckpt,
	}
}

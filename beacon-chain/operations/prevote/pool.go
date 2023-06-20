package prevote

import (
	ethpb "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1"
)

type Pool interface {
	HasPrevote(att *ethpb.PreVote) (bool, error)
	SavePrevote(att *ethpb.PreVote) error
}

func NewPool() *PrevoteCache {
	return NewPrevoteCache()
}

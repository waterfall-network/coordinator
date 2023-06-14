package prevote

import (
	ethpb "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1"
)

type Pool interface {
	HasAggregatedPrevote(att *ethpb.PreVote) (bool, error)
	SaveUnaggregatedPrevote(att *ethpb.PreVote) error
}

func NewPool() *PrevoteCaches {
	return NewPrevoteCaches()
}

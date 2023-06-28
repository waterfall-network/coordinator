package prevote

import (
	types "github.com/prysmaticlabs/eth2-types"
	ethpb "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1"
)

type Pool interface {
	HasPrevote(att *ethpb.PreVote) (bool, error)
	SavePrevote(att *ethpb.PreVote) error
	GetPrevoteBySlot(slot types.Slot) ([]*ethpb.PreVote, error)
}

func NewPool() *PrevoteCache {
	return NewPrevoteCache()
}

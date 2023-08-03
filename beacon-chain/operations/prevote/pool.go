package prevote

import (
	"context"
	"time"

	types "github.com/prysmaticlabs/eth2-types"
	ethpb "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1"
)

type Pool interface {
	HasPrevote(att *ethpb.PreVote) (bool, error)
	SavePrevote(att *ethpb.PreVote) error
	GetPrevoteBySlot(ctx context.Context, slot types.Slot) []*ethpb.PreVote
	PurgeOutdatedPrevote(t time.Time) error
}

func NewPool() *PrevoteCache {
	return NewPrevoteCache()
}

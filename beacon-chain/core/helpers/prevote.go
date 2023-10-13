package helpers

import (
	ethpb "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1"
)

func ComputeSubnetForPrevote(activeValCount uint64, prevote *ethpb.PreVote) uint64 {
	return ComputeSubnetFromCommitteeAndSlot(activeValCount, prevote.Data.Index, prevote.Data.Slot)
}

func IsAggregatedPrevote(prevote *ethpb.PreVote) bool {
	return prevote.AggregationBits.Count() > 1
}

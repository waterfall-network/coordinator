package helpers

import (
	types "github.com/prysmaticlabs/eth2-types"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/config/params"
	ethpb "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1"
)

// ComputeSubnetForPrevote returns the subnet for which the provided prevote will be broadcasted to.
func ComputeSubnetForPrevote(activeValCount uint64, prevote *ethpb.PreVote) uint64 {
	return ComputeSubnetPrevotingBySlot(activeValCount, prevote.Data.Slot)
}

func IsAggregatedPrevote(prevote *ethpb.PreVote) bool {
	return prevote.AggregationBits.Count() > 1
}

// ComputeSubnetPrevotingBySlot returns the subnet number for the passed slot.
// Prevoting subnets numbers belongs to next after attestations subnets range:
// from Config.AttestationSubnetCount to 2*Config.AttestationSubnetCount.
// One subnet for all committees per slot only.
func ComputeSubnetPrevotingBySlot(activeValCount uint64, prevotingSlot types.Slot) uint64 {
	// ignore committee for prevoting
	var comIdx types.CommitteeIndex = 0
	computedSubnet := ComputeSubnetFromCommitteeAndSlot(activeValCount, comIdx, prevotingSlot)
	// shift number to the next range
	computedSubnet += params.BeaconNetworkConfig().AttestationSubnetCount
	return computedSubnet
}

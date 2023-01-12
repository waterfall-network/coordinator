package epoch_processing

import (
	"testing"

	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/spectest/shared/phase0/epoch_processing"
)

func TestMinimal_Phase0_EpochProcessing_RewardsAndPenalties(t *testing.T) {
	epoch_processing.RunRewardsAndPenaltiesTests(t, "minimal")
}

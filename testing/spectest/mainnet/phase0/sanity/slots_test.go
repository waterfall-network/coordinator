package sanity

import (
	"testing"

	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/spectest/shared/phase0/sanity"
)

func TestMainnet_Phase0_Sanity_Slots(t *testing.T) {
	t.Skip() // Generate test data with pyton tool
	sanity.RunSlotProcessingTests(t, "mainnet")
}

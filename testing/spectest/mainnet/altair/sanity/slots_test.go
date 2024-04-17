package sanity

import (
	"testing"

	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/spectest/shared/altair/sanity"
)

func TestMainnet_Altair_Sanity_Slots(t *testing.T) {
	t.Skip() // Generate test data with pyton tool
	sanity.RunSlotProcessingTests(t, "mainnet")
}

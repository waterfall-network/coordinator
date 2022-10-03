package sanity

import (
	"testing"

	"github.com/waterfall-foundation/coordinator/testing/spectest/shared/phase0/sanity"
)

func TestMainnet_Phase0_Sanity_Slots(t *testing.T) {
	sanity.RunSlotProcessingTests(t, "mainnet")
}

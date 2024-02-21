package shuffle

import (
	"testing"

	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/spectest/shared/phase0/shuffling/core/shuffle"
)

func TestMainnet_Phase0_Shuffling_Core_Shuffle(t *testing.T) {
	t.Skip() // Generate test data with pyton tool
	shuffle.RunShuffleTests(t, "mainnet")
}

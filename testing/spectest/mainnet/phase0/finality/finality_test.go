package finality

import (
	"testing"

	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/spectest/shared/phase0/finality"
)

func TestMainnet_Phase0_Finality(t *testing.T) {
	t.Skip() // Generate test data with pyton tool
	finality.RunFinalityTest(t, "mainnet")
}

package finality

import (
	"testing"

	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/spectest/shared/altair/finality"
)

func TestMainnet_Altair_Finality(t *testing.T) {
	t.Skip() // Generate test data with pyton tool
	finality.RunFinalityTest(t, "test")
}

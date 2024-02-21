package operations

import (
	"testing"

	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/spectest/shared/altair/operations"
)

func TestMainnet_Altair_Operations_VoluntaryExit(t *testing.T) {
	t.Skip() // Generate test data with pyton tool
	operations.RunVoluntaryExitTest(t, "mainnet")
}

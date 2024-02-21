package epoch_processing

import (
	"testing"

	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/spectest/shared/altair/epoch_processing"
)

func TestMainnet_Altair_EpochProcessing_InactivityUpdates(t *testing.T) {
	t.Skip() // Generate test data with pyton tool
	epoch_processing.RunInactivityUpdatesTest(t, "mainnet")
}

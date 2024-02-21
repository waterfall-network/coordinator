package epoch_processing

import (
	"testing"

	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/spectest/shared/phase0/epoch_processing"
)

func TestMainnet_Phase0_EpochProcessing_Slashings(t *testing.T) {
	t.Skip() // Generate test data with pyton tool
	epoch_processing.RunSlashingsTests(t, "mainnet")
}

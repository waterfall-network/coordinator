package epoch_processing

import (
	"testing"

	"github.com/waterfall-foundation/coordinator/testing/spectest/shared/phase0/epoch_processing"
)

func TestMainnet_Phase0_EpochProcessing_HistoricalRootsUpdate(t *testing.T) {
	epoch_processing.RunHistoricalRootsUpdateTests(t, "mainnet")
}

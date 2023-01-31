package epoch_processing

import (
	"testing"

	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/spectest/shared/bellatrix/epoch_processing"
)

func TestMainnet_Bellatrix_EpochProcessing_RandaoMixesReset(t *testing.T) {
	epoch_processing.RunRandaoMixesResetTests(t, "mainnet")
}

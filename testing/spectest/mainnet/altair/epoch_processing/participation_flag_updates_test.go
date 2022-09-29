package epoch_processing

import (
	"testing"

	"github.com/waterfall-foundation/coordinator/testing/spectest/shared/altair/epoch_processing"
)

func TestMainnet_Altair_EpochProcessing_ParticipationFlag(t *testing.T) {
	epoch_processing.RunParticipationFlagUpdatesTests(t, "mainnet")
}

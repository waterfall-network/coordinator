package epoch_processing

import (
	"testing"

	"github.com/waterfall-foundation/coordinator/testing/spectest/shared/altair/epoch_processing"
)

func TestMinimal_Altair_EpochProcessing_SlashingsReset(t *testing.T) {
	epoch_processing.RunSlashingsResetTests(t, "minimal")
}

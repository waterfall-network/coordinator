package epoch_processing

import (
	"testing"

	"github.com/waterfall-foundation/coordinator/testing/spectest/shared/phase0/epoch_processing"
)

func TestMinimal_Phase0_EpochProcessing_JustificationAndFinalization(t *testing.T) {
	epoch_processing.RunJustificationAndFinalizationTests(t, "minimal")
}

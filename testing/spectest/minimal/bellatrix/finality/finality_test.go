package finality

import (
	"testing"

	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/spectest/shared/bellatrix/finality"
)

func TestMinimal_Bellatrix_Finality(t *testing.T) {
	finality.RunFinalityTest(t, "minimal")
}

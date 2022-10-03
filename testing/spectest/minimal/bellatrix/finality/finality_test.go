package finality

import (
	"testing"

	"github.com/waterfall-foundation/coordinator/testing/spectest/shared/bellatrix/finality"
)

func TestMinimal_Bellatrix_Finality(t *testing.T) {
	finality.RunFinalityTest(t, "minimal")
}

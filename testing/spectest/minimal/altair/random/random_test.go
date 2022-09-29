package random

import (
	"testing"

	"github.com/waterfall-foundation/coordinator/testing/spectest/shared/altair/sanity"
)

func TestMinimal_Altair_Random(t *testing.T) {
	sanity.RunBlockProcessingTest(t, "minimal", "random/random/pyspec_tests")
}

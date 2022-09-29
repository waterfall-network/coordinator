package operations

import (
	"testing"

	"github.com/waterfall-foundation/coordinator/testing/spectest/shared/bellatrix/operations"
)

func TestMinimal_Bellatrix_Operations_ProposerSlashing(t *testing.T) {
	operations.RunProposerSlashingTest(t, "minimal")
}

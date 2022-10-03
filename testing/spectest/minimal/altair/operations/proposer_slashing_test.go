package operations

import (
	"testing"

	"github.com/waterfall-foundation/coordinator/testing/spectest/shared/altair/operations"
)

func TestMinimal_Altair_Operations_ProposerSlashing(t *testing.T) {
	operations.RunProposerSlashingTest(t, "minimal")
}

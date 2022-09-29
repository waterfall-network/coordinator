package operations

import (
	"testing"

	"github.com/waterfall-foundation/coordinator/testing/spectest/shared/altair/operations"
)

func TestMinimal_Altair_Operations_AttesterSlashing(t *testing.T) {
	operations.RunAttesterSlashingTest(t, "minimal")
}

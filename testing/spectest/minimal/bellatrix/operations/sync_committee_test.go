package operations

import (
	"testing"

	"github.com/waterfall-foundation/coordinator/testing/spectest/shared/bellatrix/operations"
)

func TestMinimal_Bellatrix_Operations_SyncCommittee(t *testing.T) {
	operations.RunProposerSlashingTest(t, "minimal")
}

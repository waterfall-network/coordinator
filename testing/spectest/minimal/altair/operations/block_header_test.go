package operations

import (
	"testing"

	"github.com/waterfall-foundation/coordinator/testing/spectest/shared/altair/operations"
)

func TestMinimal_Altair_Operations_BlockHeader(t *testing.T) {
	operations.RunBlockHeaderTest(t, "minimal")
}

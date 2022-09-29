package operations

import (
	"testing"

	"github.com/waterfall-foundation/coordinator/testing/spectest/shared/bellatrix/operations"
)

func TestMinimal_Bellatrix_Operations_BlockHeader(t *testing.T) {
	operations.RunBlockHeaderTest(t, "minimal")
}

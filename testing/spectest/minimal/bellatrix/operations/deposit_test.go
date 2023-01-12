package operations

import (
	"testing"

	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/spectest/shared/bellatrix/operations"
)

func TestMinimal_Bellatrix_Operations_Deposit(t *testing.T) {
	operations.RunDepositTest(t, "minimal")
}

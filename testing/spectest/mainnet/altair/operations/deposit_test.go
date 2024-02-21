package operations

import (
	"testing"

	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/spectest/shared/altair/operations"
)

func TestMainnet_Altair_Operations_Deposit(t *testing.T) {
	t.Skip() // Generate test data with pyton tool
	operations.RunDepositTest(t, "mainnet")
}

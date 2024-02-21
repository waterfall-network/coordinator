package operations

import (
	"testing"

	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/spectest/shared/phase0/operations"
)

func TestMainnet_Phase0_Operations_BlockHeader(t *testing.T) {
	t.Skip() // Generate test data with pyton tool
	operations.RunBlockHeaderTest(t, "mainnet")
}

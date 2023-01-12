package operations

import (
	"testing"

	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/spectest/shared/bellatrix/operations"
)

func TestMainnet_Bellatrix_Operations_Attestation(t *testing.T) {
	operations.RunAttestationTest(t, "mainnet")
}

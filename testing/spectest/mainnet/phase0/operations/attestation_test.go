package operations

import (
	"testing"

	"github.com/waterfall-foundation/coordinator/testing/spectest/shared/phase0/operations"
)

func TestMainnet_Phase0_Operations_Attestation(t *testing.T) {
	operations.RunAttestationTest(t, "mainnet")
}

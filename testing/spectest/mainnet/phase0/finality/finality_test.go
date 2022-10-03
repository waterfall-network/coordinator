package finality

import (
	"testing"

	"github.com/waterfall-foundation/coordinator/testing/spectest/shared/phase0/finality"
)

func TestMainnet_Phase0_Finality(t *testing.T) {
	finality.RunFinalityTest(t, "mainnet")
}

package finality

import (
	"testing"

	"github.com/waterfall-foundation/coordinator/testing/spectest/shared/altair/finality"
)

func TestMainnet_Altair_Finality(t *testing.T) {
	finality.RunFinalityTest(t, "mainnet")
}

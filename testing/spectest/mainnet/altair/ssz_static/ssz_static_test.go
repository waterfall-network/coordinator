package ssz_static

import (
	"testing"

	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/spectest/shared/altair/ssz_static"
)

func TestMainnet_Altair_SSZStatic(t *testing.T) {
	t.Skip() // Generate test data with pyton tool
	ssz_static.RunSSZStaticTests(t, "mainnet")
}

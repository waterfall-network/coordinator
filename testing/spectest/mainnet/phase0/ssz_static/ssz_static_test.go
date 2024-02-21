package ssz_static

import (
	"testing"

	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/spectest/shared/phase0/ssz_static"
)

func TestMainnet_Phase0_SSZStatic(t *testing.T) {
	t.Skip() // Generate test data with pyton tool
	ssz_static.RunSSZStaticTests(t, "mainnet")
}

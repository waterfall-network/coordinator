package ssz_static

import (
	"testing"

	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/spectest/shared/phase0/ssz_static"
)

func TestMainnet_Phase0_SSZStatic(t *testing.T) {
	ssz_static.RunSSZStaticTests(t, "test")
}

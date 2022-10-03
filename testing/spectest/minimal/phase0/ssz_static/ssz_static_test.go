package ssz_static

import (
	"testing"

	"github.com/waterfall-foundation/coordinator/testing/spectest/shared/phase0/ssz_static"
)

func TestMinimal_Phase0_SSZStatic(t *testing.T) {
	ssz_static.RunSSZStaticTests(t, "minimal")
}

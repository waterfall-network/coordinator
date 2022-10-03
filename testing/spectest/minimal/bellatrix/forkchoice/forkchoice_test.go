package forkchoice

import (
	"testing"

	"github.com/waterfall-foundation/coordinator/config/features"
	"github.com/waterfall-foundation/coordinator/runtime/version"
	"github.com/waterfall-foundation/coordinator/testing/spectest/shared/common/forkchoice"
)

func TestMinimal_Bellatrix_Forkchoice(t *testing.T) {
	forkchoice.Run(t, "minimal", version.Bellatrix)
}

func TestMinimal_Bellatrix_Forkchoice_DoublyLinkTree(t *testing.T) {
	resetCfg := features.InitWithReset(&features.Flags{
		EnableForkChoiceDoublyLinkedTree: true,
	})
	defer resetCfg()
	forkchoice.Run(t, "minimal", version.Bellatrix)
}

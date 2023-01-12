package forkchoice

import (
	"testing"

	"gitlab.waterfall.network/waterfall/protocol/coordinator/config/features"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/runtime/version"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/spectest/shared/common/forkchoice"
)

func TestMinimal_Altair_Forkchoice(t *testing.T) {
	forkchoice.Run(t, "minimal", version.Phase0)
}

func TestMinimal_Altair_Forkchoice_DoublyLinkTre(t *testing.T) {
	resetCfg := features.InitWithReset(&features.Flags{
		EnableForkChoiceDoublyLinkedTree: true,
	})
	defer resetCfg()
	forkchoice.Run(t, "minimal", version.Phase0)
}

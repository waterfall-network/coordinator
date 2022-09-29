package forkchoice

import (
	"testing"

	"github.com/waterfall-foundation/coordinator/config/features"
	"github.com/waterfall-foundation/coordinator/runtime/version"
	"github.com/waterfall-foundation/coordinator/testing/spectest/shared/common/forkchoice"
)

func TestMainnet_Altair_Forkchoice(t *testing.T) {
	forkchoice.Run(t, "mainnet", version.Phase0)
}

func TestMainnet_Altair_Forkchoice_DoublyLinkTree(t *testing.T) {
	resetCfg := features.InitWithReset(&features.Flags{
		EnableForkChoiceDoublyLinkedTree: true,
	})
	defer resetCfg()
	forkchoice.Run(t, "mainnet", version.Phase0)
}

//go:build ((linux && amd64) || (linux && arm64) || (darwin && amd64) || (darwin && arm64) || (windows && amd64)) && !blst_disabled
// +build linux,amd64 linux,arm64 darwin,amd64 darwin,arm64 windows,amd64
// +build !blst_disabled

package blst

import (
	"fmt"
	"runtime"

	blst "github.com/supranational/blst/bindings/go"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/cache/nonblocking"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/crypto/bls/common"
)

func init() {
	// Reserve 1 core for general application work
	maxProcs := runtime.GOMAXPROCS(0) - 1
	if maxProcs <= 0 {
		maxProcs = 1
	}
	blst.SetMaxProcs(maxProcs)
	onEvict := func(_ [48]byte, _ common.PublicKey) {}
	keysCache, err := nonblocking.NewLRU(maxKeys, onEvict)
	if err != nil {
		panic(fmt.Sprintf("Could not initiate public keys cache: %v", err))
	}
	pubkeyCache = keysCache
}

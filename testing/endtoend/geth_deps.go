package endtoend

// This file contains the dependencies required for github.com/ethereum/go-ethereum/cmd/geth.
// Having these dependencies listed here helps go mod understand that these dependencies are
// necessary for end to end tests since we build go-ethereum binary for this test.
import (
	_ "gitlab.waterfall.network/waterfall/protocol/gwat/accounts"          // Required for go-ethereum e2e.
	_ "gitlab.waterfall.network/waterfall/protocol/gwat/accounts/keystore" // Required for go-ethereum e2e.
	_ "gitlab.waterfall.network/waterfall/protocol/gwat/cmd/utils"         // Required for go-ethereum e2e.
	_ "gitlab.waterfall.network/waterfall/protocol/gwat/common"            // Required for go-ethereum e2e.
	_ "gitlab.waterfall.network/waterfall/protocol/gwat/console"           // Required for go-ethereum e2e.
	_ "gitlab.waterfall.network/waterfall/protocol/gwat/eth"               // Required for go-ethereum e2e.
	_ "gitlab.waterfall.network/waterfall/protocol/gwat/eth/downloader"    // Required for go-ethereum e2e.
	_ "gitlab.waterfall.network/waterfall/protocol/gwat/ethclient"         // Required for go-ethereum e2e.
	//_ "gitlab.waterfall.network/waterfall/protocol/gwat/les"               // Required for go-ethereum e2e.
	_ "gitlab.waterfall.network/waterfall/protocol/gwat/log"     // Required for go-ethereum e2e.
	_ "gitlab.waterfall.network/waterfall/protocol/gwat/metrics" // Required for go-ethereum e2e.
	_ "gitlab.waterfall.network/waterfall/protocol/gwat/node"    // Required for go-ethereum e2e.
)

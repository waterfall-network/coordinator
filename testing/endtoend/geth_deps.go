package endtoend

// This file contains the dependencies required for github.com/ethereum/go-ethereum/cmd/geth.
// Having these dependencies listed here helps go mod understand that these dependencies are
// necessary for end to end tests since we build go-ethereum binary for this test.
import (
	_ "github.com/waterfall-foundation/gwat/accounts"          // Required for go-ethereum e2e.
	_ "github.com/waterfall-foundation/gwat/accounts/keystore" // Required for go-ethereum e2e.
	_ "github.com/waterfall-foundation/gwat/cmd/utils"         // Required for go-ethereum e2e.
	_ "github.com/waterfall-foundation/gwat/common"            // Required for go-ethereum e2e.
	_ "github.com/waterfall-foundation/gwat/console"           // Required for go-ethereum e2e.
	_ "github.com/waterfall-foundation/gwat/eth"               // Required for go-ethereum e2e.
	_ "github.com/waterfall-foundation/gwat/eth/downloader"    // Required for go-ethereum e2e.
	_ "github.com/waterfall-foundation/gwat/ethclient"         // Required for go-ethereum e2e.
	_ "github.com/waterfall-foundation/gwat/les"               // Required for go-ethereum e2e.
	_ "github.com/waterfall-foundation/gwat/log"               // Required for go-ethereum e2e.
	_ "github.com/waterfall-foundation/gwat/metrics"           // Required for go-ethereum e2e.
	_ "github.com/waterfall-foundation/gwat/node"              // Required for go-ethereum e2e.
)

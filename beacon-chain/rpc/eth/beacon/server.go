// Package beacon defines a gRPC beacon service implementation,
// following the official API standards https://ethereum.github.io/beacon-apis/#/.
// This package includes the beacon and config endpoints.
package beacon

import (
	"github.com/waterfall-foundation/coordinator/beacon-chain/blockchain"
	blockfeed "github.com/waterfall-foundation/coordinator/beacon-chain/core/feed/block"
	"github.com/waterfall-foundation/coordinator/beacon-chain/core/feed/operation"
	"github.com/waterfall-foundation/coordinator/beacon-chain/db"
	"github.com/waterfall-foundation/coordinator/beacon-chain/operations/attestations"
	"github.com/waterfall-foundation/coordinator/beacon-chain/operations/slashings"
	"github.com/waterfall-foundation/coordinator/beacon-chain/operations/voluntaryexits"
	"github.com/waterfall-foundation/coordinator/beacon-chain/p2p"
	v1alpha1validator "github.com/waterfall-foundation/coordinator/beacon-chain/rpc/prysm/v1alpha1/validator"
	"github.com/waterfall-foundation/coordinator/beacon-chain/rpc/statefetcher"
	"github.com/waterfall-foundation/coordinator/beacon-chain/state/stategen"
	"github.com/waterfall-foundation/coordinator/beacon-chain/sync"
)

// Server defines a server implementation of the gRPC Beacon Chain service,
// providing RPC endpoints to access data relevant to the Ethereum Beacon Chain.
type Server struct {
	BeaconDB                db.ReadOnlyDatabase
	ChainInfoFetcher        blockchain.ChainInfoFetcher
	GenesisTimeFetcher      blockchain.TimeFetcher
	BlockReceiver           blockchain.BlockReceiver
	BlockNotifier           blockfeed.Notifier
	OperationNotifier       operation.Notifier
	Broadcaster             p2p.Broadcaster
	AttestationsPool        attestations.Pool
	SlashingsPool           slashings.PoolManager
	VoluntaryExitsPool      voluntaryexits.PoolManager
	StateGenService         stategen.StateManager
	StateFetcher            statefetcher.Fetcher
	HeadFetcher             blockchain.HeadFetcher
	V1Alpha1ValidatorServer *v1alpha1validator.Server
	SyncChecker             sync.Checker
	CanonicalHistory        *stategen.CanonicalHistory
}

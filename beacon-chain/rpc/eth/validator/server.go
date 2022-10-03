package validator

import (
	"github.com/waterfall-foundation/coordinator/beacon-chain/blockchain"
	"github.com/waterfall-foundation/coordinator/beacon-chain/operations/attestations"
	"github.com/waterfall-foundation/coordinator/beacon-chain/operations/synccommittee"
	"github.com/waterfall-foundation/coordinator/beacon-chain/p2p"
	v1alpha1validator "github.com/waterfall-foundation/coordinator/beacon-chain/rpc/prysm/v1alpha1/validator"
	"github.com/waterfall-foundation/coordinator/beacon-chain/rpc/statefetcher"
	"github.com/waterfall-foundation/coordinator/beacon-chain/sync"
)

// Server defines a server implementation of the gRPC Validator service,
// providing RPC endpoints intended for validator clients.
type Server struct {
	HeadFetcher       blockchain.HeadFetcher
	TimeFetcher       blockchain.TimeFetcher
	SyncChecker       sync.Checker
	AttestationsPool  attestations.Pool
	PeerManager       p2p.PeerManager
	Broadcaster       p2p.Broadcaster
	StateFetcher      statefetcher.Fetcher
	SyncCommitteePool synccommittee.Pool
	V1Alpha1Server    *v1alpha1validator.Server
}

package validator

import (
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/blockchain"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/operations/attestations"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/operations/synccommittee"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/p2p"
	v1alpha1validator "gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/rpc/prysm/v1alpha1/validator"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/rpc/statefetcher"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/sync"
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

// Package validator defines a gRPC validator service implementation, providing
// critical endpoints for validator clients to submit blocks/attestations to the
// beacon node, receive assignments, and more.
package validator

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/blockchain"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/cache"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/cache/depositcache"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/core/feed"
	blockfeed "gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/core/feed/block"
	opfeed "gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/core/feed/operation"
	statefeed "gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/core/feed/state"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/core/helpers"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/core/signing"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/db"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/operations/attestations"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/operations/slashings"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/operations/synccommittee"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/operations/voluntaryexits"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/p2p"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/powchain"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/state/stategen"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/sync"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/config/params"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/encoding/bytesutil"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/network/forks"
	ethpb "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1"
	gwatCommon "gitlab.waterfall.network/waterfall/protocol/gwat/common"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Server defines a server implementation of the gRPC Validator service,
// providing RPC endpoints for obtaining validator assignments per epoch, the slots
// and committees in which particular validators need to perform their responsibilities,
// and more.
type Server struct {
	Ctx                    context.Context
	AttestationCache       *cache.AttestationCache
	ProposerSlotIndexCache *cache.ProposerPayloadIDsCache
	HeadFetcher            blockchain.HeadFetcher
	ForkFetcher            blockchain.ForkFetcher
	FinalizationFetcher    blockchain.FinalizationFetcher
	TimeFetcher            blockchain.TimeFetcher
	BlockFetcher           powchain.POWBlockFetcher
	DepositFetcher         depositcache.DepositFetcher
	ChainStartFetcher      powchain.ChainStartFetcher
	Eth1InfoFetcher        powchain.ChainInfoFetcher
	SyncChecker            sync.Checker
	StateNotifier          statefeed.Notifier
	BlockNotifier          blockfeed.Notifier
	P2P                    p2p.Broadcaster
	AttPool                attestations.Pool
	SlashingsPool          slashings.PoolManager
	ExitPool               voluntaryexits.PoolManager
	SyncCommitteePool      synccommittee.Pool
	BlockReceiver          blockchain.BlockReceiver
	MockEth1Votes          bool
	Eth1BlockFetcher       powchain.POWBlockFetcher
	PendingDepositsFetcher depositcache.PendingDepositsFetcher
	OperationNotifier      opfeed.Notifier
	StateGen               stategen.StateManager
	ReplayerBuilder        stategen.ReplayerBuilder
	BeaconDB               db.HeadAccessDatabase
	ExecutionEngineCaller  powchain.EngineCaller
}

// WaitForActivation checks if a validator public key exists in the active validator registry of the current
// beacon state, if not, then it creates a stream which listens for canonical states which contain
// the validator with the public key as an active validator record.
func (vs *Server) WaitForActivation(req *ethpb.ValidatorActivationRequest, stream ethpb.BeaconNodeValidator_WaitForActivationServer) error {
	activeValidatorExists, validatorStatuses, err := vs.activationStatus(stream.Context(), req.PublicKeys)
	if err != nil {
		return status.Errorf(codes.Internal, "Could not fetch validator status: %v", err)
	}
	res := &ethpb.ValidatorActivationResponse{
		Statuses: validatorStatuses,
	}
	if activeValidatorExists {
		return stream.Send(res)
	}
	if err := stream.Send(res); err != nil {
		return status.Errorf(codes.Internal, "Could not send response over stream: %v", err)
	}

	for {
		select {
		// Pinging every slot for activation.
		case <-time.After(time.Duration(params.BeaconConfig().SecondsPerSlot) * time.Second):
			activeValidatorExists, validatorStatuses, err := vs.activationStatus(stream.Context(), req.PublicKeys)
			if err != nil {
				return status.Errorf(codes.Internal, "Could not fetch validator status: %v", err)
			}
			res := &ethpb.ValidatorActivationResponse{
				Statuses: validatorStatuses,
			}
			if activeValidatorExists {
				return stream.Send(res)
			}
			if err := stream.Send(res); err != nil {
				return status.Errorf(codes.Internal, "Could not send response over stream: %v", err)
			}
		case <-stream.Context().Done():
			return status.Error(codes.Canceled, "Stream context canceled")
		case <-vs.Ctx.Done():
			return status.Error(codes.Canceled, "RPC context canceled")
		}
	}
}

// ValidatorIndex is called by a validator to get its index location in the beacon state.
func (vs *Server) ValidatorIndex(ctx context.Context, req *ethpb.ValidatorIndexRequest) (*ethpb.ValidatorIndexResponse, error) {
	st, err := vs.HeadFetcher.HeadState(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not determine head state: %v", err)
	}
	index, ok := st.ValidatorIndexByPubkey(bytesutil.ToBytes48(req.PublicKey))
	if !ok {
		return nil, status.Errorf(codes.Internal, "Could not find validator index for public key %#x not found", req.PublicKey)
	}

	return &ethpb.ValidatorIndexResponse{Index: index}, nil
}

// DomainData fetches the current domain version information from the beacon state.
func (vs *Server) DomainData(_ context.Context, request *ethpb.DomainRequest) (*ethpb.DomainResponse, error) {
	fork, err := forks.Fork(request.Epoch)
	if err != nil {
		return nil, err
	}
	headGenesisValidatorsRoot := vs.HeadFetcher.HeadGenesisValidatorsRoot()
	dv, err := signing.Domain(fork, request.Epoch, bytesutil.ToBytes4(request.Domain), headGenesisValidatorsRoot[:])
	if err != nil {
		return nil, err
	}
	return &ethpb.DomainResponse{
		SignatureDomain: dv,
	}, nil
}

// CanonicalHead of the current beacon chain. This method is requested on-demand
// by a validator when it is their time to propose or attest.
func (vs *Server) CanonicalHead(ctx context.Context, _ *emptypb.Empty) (*ethpb.SignedBeaconBlock, error) {
	headBlk, err := vs.HeadFetcher.HeadBlock(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not get head block: %v", err)
	}
	b, err := headBlk.PbPhase0Block()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not get head block: %v", err)
	}
	return b, nil
}

// WaitForChainStart queries the logs of the Deposit Contract in order to verify the beacon chain
// has started its runtime and validators begin their responsibilities. If it has not, it then
// subscribes to an event stream triggered by the powchain service whenever the ChainStart log does
// occur in the Deposit Contract on ETH 1.0.
func (vs *Server) WaitForChainStart(_ *emptypb.Empty, stream ethpb.BeaconNodeValidator_WaitForChainStartServer) error {
	head, err := vs.HeadFetcher.HeadState(stream.Context())
	if err != nil {
		return status.Errorf(codes.Internal, "Could not retrieve head state: %v", err)
	}
	if head != nil && !head.IsNil() {
		res := &ethpb.ChainStartResponse{
			Started:               true,
			GenesisTime:           head.GenesisTime(),
			GenesisValidatorsRoot: head.GenesisValidatorsRoot(),
		}
		return stream.Send(res)
	}

	stateChannel := make(chan *feed.Event, 1)
	stateSub := vs.StateNotifier.StateFeed().Subscribe(stateChannel)
	defer stateSub.Unsubscribe()
	for {
		select {
		case event := <-stateChannel:
			if event.Type == statefeed.Initialized {
				data, ok := event.Data.(*statefeed.InitializedData)
				if !ok {
					return errors.New("event data is not type *statefeed.InitializedData")
				}
				log.WithField("starttime", data.StartTime).Debug("Received chain started event")
				log.Debug("Sending genesis time notification to connected validator clients")
				res := &ethpb.ChainStartResponse{
					Started:               true,
					GenesisTime:           uint64(data.StartTime.Unix()),
					GenesisValidatorsRoot: data.GenesisValidatorsRoot,
				}
				return stream.Send(res)
			}
		case <-stateSub.Err():
			return status.Error(codes.Aborted, "Subscriber closed, exiting goroutine")
		case <-vs.Ctx.Done():
			return status.Error(codes.Canceled, "Context canceled")
		}
	}
}

// getOptimisticSpine retrieves optimistic spines from gwat or cached
// and extend by finalized chain to use with forkchoice
// to calculate the parent block by optimistic spine.
func (vs *Server) getOptimisticSpine(ctx context.Context) ([]gwatCommon.HashArray, error) {
	currHead, err := vs.HeadFetcher.HeadState(ctx)
	if err != nil {
		log.WithError(err).Error("Get optimistic spites failed: could not retrieve head state")
		return nil, status.Errorf(codes.Internal, "Could not retrieve head state: %v", err)
	}

	jCpRoot := bytesutil.ToBytes32(currHead.CurrentJustifiedCheckpoint().Root)
	if currHead.CurrentJustifiedCheckpoint().Epoch == 0 {
		jCpRoot, err = vs.BeaconDB.GenesisBlockRoot(ctx)
		if err != nil {
			log.WithError(err).Error("Get optimistic spites failed: retrieving of genesis root")
			return nil, status.Errorf(codes.Internal, "Could not retrieve of genesis root: %v", err)
		}
	}

	cpSt, err := vs.StateGen.StateByRoot(ctx, jCpRoot)
	if err != nil {
		log.WithError(err).WithFields(logrus.Fields{
			"jCpRoot": fmt.Sprintf("%#x", jCpRoot),
		}).Error("Get optimistic spites failed: Could not retrieve of cp state")
		return nil, status.Errorf(codes.Internal, "Could not retrieve of cp state: %v", err)
	}

	//request optimistic spine
	baseSpine := helpers.GetTerminalFinalizedSpine(cpSt)

	optSpines, err := vs.HeadFetcher.GetOptimisticSpines(ctx, baseSpine)
	if err != nil {
		errWrap := fmt.Errorf("could not get gwat optimistic spines: %v", err)
		log.WithError(errWrap).WithFields(logrus.Fields{
			"head.Slot": currHead.Slot(),
			"jcp.Slot":  cpSt.Slot(),
			"baseSpine": fmt.Sprintf("%#x", baseSpine),
		}).Error("Get optimistic spites failed: Could not retrieve of gwat optimistic spines")
		return nil, errWrap
	}

	log.WithFields(logrus.Fields{
		"head.Slot": currHead.Slot(),
		"jcp.Slot":  cpSt.Slot(),
		"baseSpine": fmt.Sprintf("%#x", baseSpine),
		//"jcp.Finalization": fmt.Sprintf("%#x", cpSt.SpineData().Finalization),
		//"jcp.CpFinalized":  fmt.Sprintf("%#x", cpSt.SpineData().CpFinalized),
		//"optSpines":        optSpines,
	}).Error("Get optimistic spites: data retrieved")

	//prepend current optimistic finalization to optimistic spine to calc parent
	cpf := gwatCommon.HashArrayFromBytes(cpSt.SpineData().CpFinalized)
	fin := gwatCommon.HashArrayFromBytes(cpSt.SpineData().Finalization)
	extOptSpines := make([]gwatCommon.HashArray, len(cpf)+len(fin)+len(optSpines))
	ofs := 0
	for i, offset, a := 0, ofs, cpf; i < len(a); i++ {
		ofs++
		extOptSpines[offset+i] = gwatCommon.HashArray{a[i]}
	}
	for i, offset, a := 0, ofs, fin; i < len(a); i++ {
		ofs++
		extOptSpines[offset+i] = gwatCommon.HashArray{a[i]}
	}
	for i, offset, a := 0, ofs, optSpines; i < len(a); i++ {
		ofs++
		extOptSpines[offset+i] = a[i].Copy()
	}
	return optSpines, nil
}

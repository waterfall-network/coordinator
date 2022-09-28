package powchain

import (
	"context"
	"fmt"
	"go.opencensus.io/trace"
	"math/big"

	"github.com/pkg/errors"
	types "github.com/prysmaticlabs/eth2-types"
	gwatCommon "github.com/waterfall-foundation/gwat/common"
	gwatTypes "github.com/waterfall-foundation/gwat/core/types"
	"github.com/waterfall-foundation/gwat/rpc"
)

const (
	//ExecutionDagSyncMethod request string for JSON-RPC of dag api.
	ExecutionDagSyncMethod = "dag_sync"
	//ExecutionDagGetCandidatesMethod request string for JSON-RPC of dag api.
	ExecutionDagGetCandidatesMethod = "dag_getCandidates"
	//ExecutionDagFinalizeMethod request string for JSON-RPC of dag api.
	ExecutionDagFinalizeMethod = "dag_finalize"
	//ExecutionDagHeadSyncReadyMethod request string for JSON-RPC of dag api.
	ExecutionDagHeadSyncReadyMethod = "dag_headSyncReady"
	//ExecutionDagHeadSyncMethod request string for JSON-RPC of dag api.
	ExecutionDagHeadSyncMethod = "dag_headSync"
)

// ExecutionDagSync executing following procedures:
// - finalisation
// - get candidates
// - block creation
// by calling dag_sync via JSON-RPC.
func (s *Service) ExecutionDagSync(ctx context.Context, syncParams *gwatTypes.ConsensusInfo) (gwatCommon.HashArray, error) {
	ctx, span := trace.StartSpan(ctx, "powchain.dag-api-client.ExecutionDagSync")
	defer span.End()
	result := &gwatTypes.ConsensusResult{}

	if s.rpcClient == nil {
		return result.Candidates, fmt.Errorf("Rpc Client not init")
	}

	err := s.rpcClient.CallContext(
		ctx,
		result,
		ExecutionDagSyncMethod,
		syncParams,
	)
	if result.Error != nil {
		err = errors.New(*result.Error)
	}
	if result.Candidates == nil {
		result.Candidates = gwatCommon.HashArray{}
	}
	return result.Candidates, handleDagRPCError(err)
}

// ExecutionDagFinalize executing finalisation procedure
// by calling dag_finalize via JSON-RPC.
func (s *Service) ExecutionDagFinalize(ctx context.Context, syncParams *gwatTypes.ConsensusInfo) (*map[string]string, error) {
	ctx, span := trace.StartSpan(ctx, "powchain.dag-api-client.ExecutionDagFinalize")
	defer span.End()
	result := &gwatTypes.FinalizationResult{}

	if s.rpcClient == nil {
		return result.Info, fmt.Errorf("Rpc Client not init")
	}

	err := s.rpcClient.CallContext(
		ctx,
		result,
		ExecutionDagFinalizeMethod,
		syncParams,
	)
	if result.Error != nil {
		err = errors.New(*result.Error)
	}
	return result.Info, handleDagRPCError(err)
}

// ExecutionDagGetCandidates executing consensus procedure
// by calling dag_getCandidates via JSON-RPC.
func (s *Service) ExecutionDagGetCandidates(ctx context.Context, slot types.Slot) (gwatCommon.HashArray, error) {
	ctx, span := trace.StartSpan(ctx, "powchain.dag-api-client.ExecutionGetCandidates")
	defer span.End()
	result := &gwatTypes.CandidatesResult{}

	if s.rpcClient == nil {
		return result.Candidates, fmt.Errorf("Rpc Client not init")
	}

	err := s.rpcClient.CallContext(
		ctx,
		result,
		ExecutionDagGetCandidatesMethod,
		slot,
	)
	if result.Error != nil {
		err = errors.New(*result.Error)
	}
	if result.Candidates == nil {
		result.Candidates = gwatCommon.HashArray{}
	}
	return result.Candidates, handleDagRPCError(err)
}

// ExecutionDagHeadSyncReady executing head sync ready procedure
// by calling dag_headSyncReady via JSON-RPC.
func (s *Service) ExecutionDagHeadSyncReady(ctx context.Context, params *gwatTypes.ConsensusInfo) (bool, error) {
	ctx, span := trace.StartSpan(ctx, "powchain.dag-api-client.ExecutionDagHeadSyncReady")
	defer span.End()
	result := false

	if s.rpcClient == nil {
		return result, fmt.Errorf("Rpc Client not init")
	}
	err := s.rpcClient.CallContext(
		ctx,
		&result,
		ExecutionDagHeadSyncReadyMethod,
		params,
	)
	return result, handleDagRPCError(err)
}

// ExecutionDagHeadSync executing head sync procedure
// by calling dag_headSync via JSON-RPC.
func (s *Service) ExecutionDagHeadSync(ctx context.Context, params []gwatTypes.ConsensusInfo) (bool, error) {
	ctx, span := trace.StartSpan(ctx, "powchain.dag-api-client.ExecutionDagHeadSync")
	defer span.End()
	result := false

	if s.rpcClient == nil {
		return result, fmt.Errorf("Rpc Client not init")
	}
	err := s.rpcClient.CallContext(
		ctx,
		&result,
		ExecutionDagHeadSyncMethod,
		params,
	)

	if err != nil {
		log.WithError(err).Error("ExecutionDagHeadSync")
	}

	return result, handleDagRPCError(err)
}

// GetHeaderByHash retrieves gwat block header by hash.
func (s *Service) GetHeaderByHash(ctx context.Context, hash gwatCommon.Hash) (*gwatTypes.Header, error) {
	ctx, span := trace.StartSpan(ctx, "powchain.dag-api-client.GetHeaderByHash")
	defer span.End()
	if s.rpcClient == nil {
		return nil, fmt.Errorf("Rpc Client not init")
	}
	header, err := s.eth1DataFetcher.HeaderByHash(ctx, hash)
	return header, handleDagRPCError(err)
}

// GetHeaderByNumber retrieves gwat block header by finalization number.
func (s *Service) GetHeaderByNumber(ctx context.Context, nr *big.Int) (*gwatTypes.Header, error) {
	ctx, span := trace.StartSpan(ctx, "powchain.dag-api-client.GetHeaderByNumber")
	defer span.End()
	if s.rpcClient == nil {
		return nil, fmt.Errorf("Rpc Client not init")
	}
	header, err := s.eth1DataFetcher.HeaderByNumber(ctx, nr)
	return header, handleDagRPCError(err)
}

// handleDagRPCError errors received from the RPC server according to the specification.
func handleDagRPCError(err error) error {
	if err == nil {
		return nil
	}
	if isTimeout(err) {
		return errors.Wrapf(ErrDagHTTPTimeout, "%s", err)
	}
	e, ok := err.(rpc.Error)
	if !ok {
		return errors.Wrap(err, "got an unexpected error")
	}
	switch e.ErrorCode() {
	case -32700:
		return ErrParse
	case -32600:
		return ErrInvalidRequest
	case -32601:
		return ErrMethodNotFound
	case -32602:
		return ErrInvalidParams
	case -32603:
		return ErrInternal
	case -32001:
		return ErrUnknownPayload
	case -32000:
		// Only -32000 status codes are data errors in the RPC specification.
		errWithData, ok := err.(rpc.DataError)
		if !ok {
			return errors.Wrap(err, "got an unexpected error")
		}
		return errors.Wrapf(ErrServer, "%v", errWithData.ErrorData())
	default:
		return err
	}
}

// ErrDagHTTPTimeout returns true if the error is a http.Client timeout error.
var ErrDagHTTPTimeout = errors.New("timeout from http.DagClient")

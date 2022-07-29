package powchain

import (
	"context"
	"github.com/waterfall-foundation/gwat/dag/finalizer"
	"go.opencensus.io/trace"

	"github.com/pkg/errors"
	"github.com/waterfall-foundation/gwat/dag"
	"github.com/waterfall-foundation/gwat/rpc"
)

const (
	//ExecutionDagSyncMethod request string for JSON-RPC of dag api.
	ExecutionDagSyncMethod = "dag_sync"
	//ExecutionDagGetCandidatesMethod request string for JSON-RPC of dag api.
	ExecutionDagGetCandidatesMethod = "dag_getCandidates"
	//ExecutionDagFinalizeMethod request string for JSON-RPC of dag api.
	ExecutionDagFinalizeMethod = "dag_finalize"
)

// ExecutionDagSync executing following procedures:
// - finalisation
// - get candidates
// - block creation
// by calling dag_sync via JSON-RPC.
func (s *Service) ExecutionDagSync(ctx context.Context, syncParams *dag.ConsensusInfo) (finalizer.NrHashMap, error) {
	ctx, span := trace.StartSpan(ctx, "powchain.dag-api-client.ExecutionDagSync")
	defer span.End()
	result := &dag.ConsensusResult{}
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
		result.Candidates = &finalizer.NrHashMap{}
	}
	if result.Candidates.HasGap() {
		err = finalizer.ErrChainGap
		result.Candidates = &finalizer.NrHashMap{}
	}
	return *result.Candidates, handleDagRPCError(err)
}

// ExecutionDagFinalize executing finalisation procedure
// by calling dag_finalize via JSON-RPC.
func (s *Service) ExecutionDagFinalize(ctx context.Context, syncParams *dag.ConsensusInfo) (*map[string]string, error) {
	ctx, span := trace.StartSpan(ctx, "powchain.dag-api-client.ExecutionDagFinalize")
	defer span.End()
	result := &dag.FinalizationResult{}
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
// by calling dag_sync via JSON-RPC.
func (s *Service) ExecutionDagGetCandidates(ctx context.Context) (finalizer.NrHashMap, error) {
	ctx, span := trace.StartSpan(ctx, "powchain.dag-api-client.ExecutionGetCandidates")
	defer span.End()
	result := &dag.CandidatesResult{}
	err := s.rpcClient.CallContext(
		ctx,
		result,
		ExecutionDagGetCandidatesMethod,
	)
	if result.Error != nil {
		err = errors.New(*result.Error)
	}
	if result.Candidates == nil {
		result.Candidates = &finalizer.NrHashMap{}
	}
	if result.Candidates.HasGap() {
		err = finalizer.ErrChainGap
		result.Candidates = &finalizer.NrHashMap{}
	}
	return *result.Candidates, handleDagRPCError(err)
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

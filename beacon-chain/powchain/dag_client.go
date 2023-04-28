package powchain

import (
	"context"
	"fmt"
	"go.opencensus.io/trace"
	"math/big"
	"time"

	"github.com/pkg/errors"
	types "github.com/prysmaticlabs/eth2-types"
	"github.com/sirupsen/logrus"
	gwatCommon "gitlab.waterfall.network/waterfall/protocol/gwat/common"
	"gitlab.waterfall.network/waterfall/protocol/gwat/common/hexutil"
	gwatTypes "gitlab.waterfall.network/waterfall/protocol/gwat/core/types"
	"gitlab.waterfall.network/waterfall/protocol/gwat/rpc"
)

const (
	//ExecutionDagGetCandidatesMethod request string for JSON-RPC of dag api.
	ExecutionDagGetCandidatesMethod = "dag_getCandidates"
	//ExecutionDagFinalizeMethod request string for JSON-RPC of dag api.
	ExecutionDagFinalizeMethod = "dag_finalize"
	//ExecutionDagCoordinatedStateMethod request string for JSON-RPC of dag api.
	ExecutionDagCoordinatedStateMethod = "dag_coordinatedState"
	//ExecutionDagSyncSlotInfoMethod request string for JSON-RPC of dag api.
	ExecutionDagSyncSlotInfoMethod = "dag_syncSlotInfo"
	//ExecutionDagValidateSpinesMethod request string for JSON-RPC of dag api.
	ExecutionDagValidateSpinesMethod = "dag_validateSpines"
	//ExecutionDepositCountMethod request string for JSON-RPC of validator api.
	ExecutionDepositCountMethod = "wat_validator_DepositCount"
)

// ExecutionDagFinalize executing finalisation procedure
// by calling dag_finalize via JSON-RPC.
func (s *Service) ExecutionDagFinalize(ctx context.Context, params *gwatTypes.FinalizationParams) (*gwatTypes.FinalizationResult, error) {
	ctx, span := trace.StartSpan(ctx, "powchain.dag-api-client.ExecutionDagFinalize")
	defer span.End()
	result := &gwatTypes.FinalizationResult{}

	start := time.Now()

	if s.rpcClient == nil {
		return nil, fmt.Errorf("Rpc Client not init")
	}

	err := s.rpcClient.CallContext(
		ctx,
		result,
		ExecutionDagFinalizeMethod,
		params,
	)

	if err != nil {
		log.WithError(err).WithFields(logrus.Fields{
			"BaseSpine": params.BaseSpine.Hex(),
			"Spines":    params.Spines,
		}).Error("Dag Finalize")
	}

	if result.Error != nil {
		err = errors.New(*result.Error)
	}

	if result.Error != nil {
		err = errors.New(*result.Error)
	}

	log.WithField("elapsed", time.Since(start)).WithField(
		"api", ExecutionDagFinalizeMethod,
	).Info("Request finish")

	return result, handleDagRPCError(err)
}

// ExecutionDagCoordinatedState executing procedure to retrieve gwat coordinated state
// by calling dag_finalize via JSON-RPC.
func (s *Service) ExecutionDagCoordinatedState(ctx context.Context) (*gwatTypes.FinalizationResult, error) {
	ctx, span := trace.StartSpan(ctx, "powchain.dag-api-client.DagCoordinatedState")
	defer span.End()
	result := &gwatTypes.FinalizationResult{}

	if s.rpcClient == nil {
		return nil, fmt.Errorf("Rpc Client not init")
	}

	err := s.rpcClient.CallContext(
		ctx,
		result,
		ExecutionDagCoordinatedStateMethod,
	)

	if err != nil {
		log.WithError(err).Error("Dag Coordinated State")
	}

	if result.Error != nil {
		err = errors.New(*result.Error)
	}

	if result.Error != nil {
		err = errors.New(*result.Error)
	}

	return result, handleDagRPCError(err)
}

// ExecutionDagGetCandidates executing consensus procedure
// by calling dag_getCandidates via JSON-RPC.
func (s *Service) ExecutionDagGetCandidates(ctx context.Context, slot types.Slot) (gwatCommon.HashArray, error) {
	ctx, span := trace.StartSpan(ctx, "powchain.dag-api-client.ExecutionGetCandidates")
	defer span.End()
	result := &gwatTypes.CandidatesResult{}

	start := time.Now()

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

	log.WithField("elapsed", time.Since(start)).WithField(
		"api", ExecutionDagGetCandidatesMethod,
	).Info("Request finish")

	return result.Candidates, handleDagRPCError(err)
}

// ExecutionDagSyncSlotInfo executing sync slot info procedure
// by calling dag_syncSlotInfo via JSON-RPC.
func (s *Service) ExecutionDagSyncSlotInfo(ctx context.Context, params *gwatTypes.SlotInfo) (bool, error) {
	ctx, span := trace.StartSpan(ctx, "powchain.dag-api-client.ExecutionDagSyncSlotInfo")
	defer span.End()
	var result bool

	if s.rpcClient == nil {
		return result, fmt.Errorf("Rpc Client not init")
	}
	err := s.rpcClient.CallContext(
		ctx,
		&result,
		ExecutionDagSyncSlotInfoMethod,
		params,
	)

	if err != nil {
		log.WithError(err).Error("ExecutionDagSyncSlotInfo")
	}

	return result, handleDagRPCError(err)
}

// ExecutionDagValidateSpines executing spines validation
// by calling dag_validateSpines via JSON-RPC.
func (s *Service) ExecutionDagValidateSpines(ctx context.Context, params gwatCommon.HashArray) (bool, error) {
	ctx, span := trace.StartSpan(ctx, "powchain.dag-api-client.ExecutionDagValidateSpines")
	defer span.End()
	var result bool

	start := time.Now()

	if s.rpcClient == nil {
		return result, fmt.Errorf("Rpc Client not init")
	}
	err := s.rpcClient.CallContext(
		ctx,
		&result,
		ExecutionDagValidateSpinesMethod,
		params,
	)

	if err != nil {
		log.WithError(err).Error("ExecutionDagValidateSpines")
	}

	log.WithField("elapsed", time.Since(start)).WithField(
		"api", ExecutionDagValidateSpinesMethod,
	).Info("Request finish")

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

// GetDepositCount retrieves current gwat deposit count
func (s *Service) GetDepositCount(ctx context.Context) (uint64, error) {
	ctx, span := trace.StartSpan(ctx, "powchain.dag-api-client.GetDepositCount")
	defer span.End()
	//var result uint64
	var result string

	if s.rpcClient == nil {
		return 0, fmt.Errorf("Rpc Client not init")
	}
	err := s.rpcClient.CallContext(
		ctx,
		&result,
		ExecutionDepositCountMethod,
		nil,
	)
	if err != nil {
		log.WithError(err).Error("GetDepositCount")
	}

	count, err := hexutil.DecodeUint64(result)
	log.WithError(err).WithField(
		"result", result,
	).WithField(
		"uint", count,
	).Info("Get deposit count")

	return count, handleDagRPCError(err)
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

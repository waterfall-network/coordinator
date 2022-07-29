package powchain

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	mocks "github.com/prysmaticlabs/prysm/beacon-chain/powchain/testing"
	"github.com/prysmaticlabs/prysm/testing/require"
	"github.com/waterfall-foundation/gwat/common"
	"github.com/waterfall-foundation/gwat/dag"
	"github.com/waterfall-foundation/gwat/dag/finalizer"
	"github.com/waterfall-foundation/gwat/rpc"
)

var (
	_ = EngineCaller(&Service{})
	_ = EngineCaller(&mocks.EngineClient{})
)

func TestDagClient_IPC(t *testing.T) {
	server := newTestDagIPCServer(t)
	defer server.Stop()
	rpcClient := rpc.DialInProc(server)
	defer rpcClient.Close()
	srv := &Service{}
	srv.rpcClient = rpcClient
	ctx := context.Background()
	fix := dagFixtures()

	t.Run(ExecutionDagGetCandidatesMethod, func(t *testing.T) {
		want, ok := fix["ExecutionCandidates"].(*dag.CandidatesResult)
		require.Equal(t, true, ok)
		resp, err := srv.ExecutionDagGetCandidates(ctx)
		require.NoError(t, err)
		require.DeepEqual(t, *want.Candidates, resp)
	})
	t.Run(ExecutionDagFinalizeMethod, func(t *testing.T) {
		want, ok := fix["ExecutionFinalize"].(*dag.FinalizationResult)
		require.Equal(t, true, ok)

		hash_1 := common.HexToHash("0xa659fcd4ed3f3ad9cd43ab36eb29080a4655328fe16f045962afab1d66a5da09")
		arg := &dag.ConsensusInfo{
			Epoch:      15,
			Slot:       10,
			Creators:   []common.Address{common.HexToAddress("0x0000000000000000000000000000000000000000")},
			Finalizing: finalizer.NrHashMap{1: &hash_1},
		}
		resp, err := srv.ExecutionDagFinalize(ctx, arg)
		require.ErrorContains(t, *want.Error, err)
		require.DeepEqual(t, want.Info, resp)
	})
	t.Run(ExecutionDagSyncMethod, func(t *testing.T) {
		want, ok := fix["ExecutionSync"].(*dag.ConsensusResult)
		require.Equal(t, true, ok)

		hash_1 := common.HexToHash("0xa659fcd4ed3f3ad9cd43ab36eb29080a4655328fe16f045962afab1d66a5da09")
		arg := &dag.ConsensusInfo{
			Epoch:      15,
			Slot:       10,
			Creators:   []common.Address{common.HexToAddress("0x0000000000000000000000000000000000000000")},
			Finalizing: finalizer.NrHashMap{1: &hash_1},
		}
		resp, err := srv.ExecutionDagSync(ctx, arg)
		require.NoError(t, err)
		require.DeepEqual(t, *want.Candidates, resp)
	})
}

func TestDagClient_HTTP(t *testing.T) {
	ctx := context.Background()
	fix := dagFixtures()

	t.Run(ExecutionDagGetCandidatesMethod, func(t *testing.T) {
		want, ok := fix["ExecutionCandidates"].(*dag.CandidatesResult)
		require.Equal(t, true, ok)
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			defer func() {
				require.NoError(t, r.Body.Close())
			}()
			_, err := ioutil.ReadAll(r.Body)
			require.NoError(t, err)
			resp := map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      1,
				"result":  want,
			}
			err = json.NewEncoder(w).Encode(resp)
			require.NoError(t, err)
		}))
		defer srv.Close()

		rpcClient, err := rpc.DialHTTP(srv.URL)
		require.NoError(t, err)
		defer rpcClient.Close()

		service := &Service{}
		service.rpcClient = rpcClient

		// We call the RPC method via HTTP and expect a proper result.
		resp, err := service.ExecutionDagGetCandidates(ctx)
		require.NoError(t, err)
		require.DeepEqual(t, *want.Candidates, resp)
	})
	t.Run(ExecutionDagFinalizeMethod, func(t *testing.T) {
		hash_1 := common.HexToHash("0xa659fcd4ed3f3ad9cd43ab36eb29080a4655328fe16f045962afab1d66a5da09")
		arg := &dag.ConsensusInfo{
			Epoch:      15,
			Slot:       10,
			Creators:   []common.Address{common.HexToAddress("0x0000000000000000000000000000000000000000")},
			Finalizing: finalizer.NrHashMap{1: &hash_1},
		}
		want, ok := fix["ExecutionFinalize"].(*dag.FinalizationResult)
		require.Equal(t, true, ok)
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			defer func() {
				require.NoError(t, r.Body.Close())
			}()
			enc, err := ioutil.ReadAll(r.Body)
			require.NoError(t, err)
			jsonRequestString := string(enc)
			// We expect the JSON string RPC request contains the right arguments.
			sArgs, _ := arg.MarshalJSON()
			//t.Logf("=========== %v", jsonRequestString)
			//t.Logf("=========== %v", fmt.Sprintf("%s", sArgs))
			require.Equal(t, true, strings.Contains(
				jsonRequestString, fmt.Sprintf("%v", fmt.Sprintf("%s", sArgs)),
			))
			resp := map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      1,
				"result":  want,
			}

			err = json.NewEncoder(w).Encode(resp)
			require.NoError(t, err)
		}))
		defer srv.Close()

		rpcClient, err := rpc.DialHTTP(srv.URL)
		require.NoError(t, err)
		defer rpcClient.Close()

		service := &Service{}
		service.rpcClient = rpcClient

		// We call the RPC method via HTTP and expect a proper result.
		resp, err := service.ExecutionDagFinalize(ctx, arg)
		require.ErrorContains(t, *want.Error, err)
		require.DeepEqual(t, want.Info, resp)
	})
	t.Run(ExecutionDagSyncMethod, func(t *testing.T) {
		hash_1 := common.HexToHash("0xa659fcd4ed3f3ad9cd43ab36eb29080a4655328fe16f045962afab1d66a5da09")
		arg := &dag.ConsensusInfo{
			Epoch:      15,
			Slot:       10,
			Creators:   []common.Address{common.HexToAddress("0x0000000000000000000000000000000000000000")},
			Finalizing: finalizer.NrHashMap{1: &hash_1},
		}
		want, ok := fix["ExecutionSync"].(*dag.ConsensusResult)
		require.Equal(t, true, ok)
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			defer func() {
				require.NoError(t, r.Body.Close())
			}()
			enc, err := ioutil.ReadAll(r.Body)
			require.NoError(t, err)
			jsonRequestString := string(enc)
			// We expect the JSON string RPC request contains the right arguments.
			sArgs, _ := arg.MarshalJSON()
			//t.Logf("=========== %v", jsonRequestString)
			//t.Logf("=========== %v", fmt.Sprintf("%s", sArgs))
			require.Equal(t, true, strings.Contains(
				jsonRequestString, fmt.Sprintf("%v", fmt.Sprintf("%s", sArgs)),
			))
			resp := map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      1,
				"result":  want,
			}

			err = json.NewEncoder(w).Encode(resp)
			require.NoError(t, err)
		}))
		defer srv.Close()

		rpcClient, err := rpc.DialHTTP(srv.URL)
		require.NoError(t, err)
		defer rpcClient.Close()

		service := &Service{}
		service.rpcClient = rpcClient

		// We call the RPC method via HTTP and expect a proper result.
		resp, err := service.ExecutionDagSync(ctx, arg)
		require.NoError(t, err)
		require.DeepEqual(t, *want.Candidates, resp)
	})
}

func newTestDagIPCServer(t *testing.T) *rpc.Server {
	server := rpc.NewServer()
	err := server.RegisterName("dag", new(testDagEngineService))
	require.NoError(t, err)
	return server
}

func dagFixtures() map[string]interface{} {
	hash_1 := common.HexToHash("0xa659fcd4ed3f3ad9cd43ab36eb29080a4655328fe16f045962afab1d66a5da09")
	executionCandidates := &dag.CandidatesResult{
		Error:      nil,
		Candidates: &finalizer.NrHashMap{1: &hash_1},
	}

	finErr := "test error"
	executionFinalize := &dag.FinalizationResult{
		Error: &finErr,
		Info:  nil,
	}

	executionSync := &dag.ConsensusResult{
		Error:      nil,
		Candidates: &finalizer.NrHashMap{1: &hash_1},
	}
	return map[string]interface{}{
		"ExecutionFinalize":   executionFinalize,
		"ExecutionCandidates": executionCandidates,
		"ExecutionSync":       executionSync,
	}
}

type testDagEngineService struct{}

func (*testDagEngineService) NoArgsRets() {}

func (*testDagEngineService) Sync(
	_ context.Context, _ *dag.ConsensusInfo,
) *dag.ConsensusResult {
	fix := dagFixtures()
	item, ok := fix["ExecutionSync"].(*dag.ConsensusResult)
	if !ok {
		panic("not found")
	}
	return item
}

func (*testDagEngineService) GetCandidates(
	_ context.Context,
) *dag.CandidatesResult {
	fix := dagFixtures()
	item, ok := fix["ExecutionCandidates"].(*dag.CandidatesResult)
	if !ok {
		panic("not found")
	}
	return item
}

func (*testDagEngineService) Finalize(
	_ context.Context, _ *dag.ConsensusInfo,
) *dag.FinalizationResult {
	fix := dagFixtures()
	item, ok := fix["ExecutionFinalize"].(*dag.FinalizationResult)
	if !ok {
		panic("not found")
	}
	return item
}

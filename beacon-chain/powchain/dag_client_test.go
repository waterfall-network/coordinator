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

	mocks "github.com/waterfall-foundation/coordinator/beacon-chain/powchain/testing"
	"github.com/waterfall-foundation/coordinator/testing/require"
	"github.com/waterfall-foundation/gwat/common"
	gwatCommon "github.com/waterfall-foundation/gwat/common"
	gwatTypes "github.com/waterfall-foundation/gwat/core/types"
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
		want, ok := fix["ExecutionCandidates"].(*gwatTypes.CandidatesResult)
		require.Equal(t, true, ok)
		resp, err := srv.ExecutionDagGetCandidates(ctx, 1000)
		require.NoError(t, err)
		require.DeepEqual(t, want.Candidates, resp)
	})
	t.Run(ExecutionDagFinalizeMethod, func(t *testing.T) {
		want, ok := fix["ExecutionFinalize"].(*gwatTypes.FinalizationResult)
		require.Equal(t, true, ok)

		hash_1 := common.HexToHash("0xa659fcd4ed3f3ad9cd43ab36eb29080a4655328fe16f045962afab1d66a5da09")
		arg := &gwatTypes.ConsensusInfo{
			Slot:       10,
			Creators:   []common.Address{common.HexToAddress("0x0000000000000000000000000000000000000000")},
			Finalizing: gwatCommon.HashArray{hash_1},
		}
		resp, err := srv.ExecutionDagFinalize(ctx, arg)
		require.ErrorContains(t, *want.Error, err)
		require.DeepEqual(t, want.Info, resp)
	})
	t.Run(ExecutionDagSyncMethod, func(t *testing.T) {
		want, ok := fix["ExecutionSync"].(*gwatTypes.ConsensusResult)
		require.Equal(t, true, ok)

		hash_1 := common.HexToHash("0xa659fcd4ed3f3ad9cd43ab36eb29080a4655328fe16f045962afab1d66a5da09")
		arg := &gwatTypes.ConsensusInfo{
			Slot:       10,
			Creators:   []common.Address{common.HexToAddress("0x0000000000000000000000000000000000000000")},
			Finalizing: gwatCommon.HashArray{hash_1},
		}
		resp, err := srv.ExecutionDagSync(ctx, arg)
		require.NoError(t, err)
		require.DeepEqual(t, want.Candidates, resp)
	})
	// head sync
	t.Run(ExecutionDagHeadSyncReadyMethod, func(t *testing.T) {
		want, ok := fix["ExecutionHeadSyncReady"].(bool)
		require.Equal(t, true, ok)

		hash_1 := common.HexToHash("0xa659fcd4ed3f3ad9cd43ab36eb29080a4655328fe16f045962afab1d66a5da09")
		arg := &gwatTypes.ConsensusInfo{
			Slot:       10,
			Creators:   []common.Address{common.HexToAddress("0x0000000000000000000000000000000000000000")},
			Finalizing: gwatCommon.HashArray{hash_1},
		}
		resp, err := srv.ExecutionDagHeadSyncReady(ctx, arg)
		require.NoError(t, err)
		require.DeepEqual(t, want, resp)
	})
	t.Run(ExecutionDagHeadSyncMethod, func(t *testing.T) {
		want, ok := fix["ExecutionHeadSync"].(bool)
		require.Equal(t, true, ok)

		hash_1 := common.HexToHash("0xa659fcd4ed3f3ad9cd43ab36eb29080a4655328fe16f045962afab1d66a5da09")
		arg := []gwatTypes.ConsensusInfo{
			gwatTypes.ConsensusInfo{
				Slot:       10,
				Creators:   []common.Address{common.HexToAddress("0x0000000000000000000000000000000000000000")},
				Finalizing: gwatCommon.HashArray{hash_1},
			},
		}
		resp, err := srv.ExecutionDagHeadSync(ctx, arg)
		require.NoError(t, err)
		require.DeepEqual(t, want, resp)
	})
}

func TestDagClient_HTTP(t *testing.T) {
	ctx := context.Background()
	fix := dagFixtures()

	t.Run(ExecutionDagGetCandidatesMethod, func(t *testing.T) {
		want, ok := fix["ExecutionCandidates"].(*gwatTypes.CandidatesResult)
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
		resp, err := service.ExecutionDagGetCandidates(ctx, 1000)
		require.NoError(t, err)
		require.DeepEqual(t, want.Candidates, resp)
	})
	t.Run(ExecutionDagFinalizeMethod, func(t *testing.T) {
		hash_1 := common.HexToHash("0xa659fcd4ed3f3ad9cd43ab36eb29080a4655328fe16f045962afab1d66a5da09")
		arg := &gwatTypes.ConsensusInfo{
			Slot:       10,
			Creators:   []common.Address{common.HexToAddress("0x0000000000000000000000000000000000000000")},
			Finalizing: gwatCommon.HashArray{hash_1},
		}
		want, ok := fix["ExecutionFinalize"].(*gwatTypes.FinalizationResult)
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
		arg := &gwatTypes.ConsensusInfo{
			Slot:       10,
			Creators:   []common.Address{common.HexToAddress("0x0000000000000000000000000000000000000000")},
			Finalizing: gwatCommon.HashArray{hash_1},
		}
		want, ok := fix["ExecutionSync"].(*gwatTypes.ConsensusResult)
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
		require.DeepEqual(t, want.Candidates, resp)
	})
	// head sync
	t.Run(ExecutionDagHeadSyncReadyMethod, func(t *testing.T) {
		hash_1 := common.HexToHash("0xa659fcd4ed3f3ad9cd43ab36eb29080a4655328fe16f045962afab1d66a5da09")
		arg := &gwatTypes.ConsensusInfo{
			Slot:       10,
			Creators:   []common.Address{common.HexToAddress("0x0000000000000000000000000000000000000000")},
			Finalizing: gwatCommon.HashArray{hash_1},
		}
		want, ok := fix["ExecutionHeadSyncReady"].(bool)
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
		resp, _ := service.ExecutionDagHeadSyncReady(ctx, arg)
		require.DeepEqual(t, want, resp)
	})

	t.Run(ExecutionDagHeadSyncMethod, func(t *testing.T) {
		hash_1 := common.HexToHash("0xa659fcd4ed3f3ad9cd43ab36eb29080a4655328fe16f045962afab1d66a5da09")
		arg := []gwatTypes.ConsensusInfo{
			gwatTypes.ConsensusInfo{
				Slot:       10,
				Creators:   []common.Address{common.HexToAddress("0x0000000000000000000000000000000000000000")},
				Finalizing: gwatCommon.HashArray{hash_1},
			},
		}
		want, ok := fix["ExecutionHeadSync"].(bool)
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
			sArgs, _ := json.Marshal(arg)
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
		resp, _ := service.ExecutionDagHeadSync(ctx, arg)
		require.DeepEqual(t, want, resp)
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
	executionCandidates := &gwatTypes.CandidatesResult{
		Error:      nil,
		Candidates: gwatCommon.HashArray{hash_1},
	}

	finErr := "test error"
	executionFinalize := &gwatTypes.FinalizationResult{
		Error: &finErr,
		Info:  nil,
	}

	executionSync := &gwatTypes.ConsensusResult{
		Error:      nil,
		Candidates: gwatCommon.HashArray{hash_1},
	}
	return map[string]interface{}{
		"ExecutionFinalize":      executionFinalize,
		"ExecutionCandidates":    executionCandidates,
		"ExecutionSync":          executionSync,
		"ExecutionHeadSyncReady": true,
		"ExecutionHeadSync":      true,
	}
}

type testDagEngineService struct{}

func (*testDagEngineService) NoArgsRets() {}

func (*testDagEngineService) Sync(
	_ context.Context, _ *gwatTypes.ConsensusInfo,
) *gwatTypes.ConsensusResult {
	fix := dagFixtures()
	item, ok := fix["ExecutionSync"].(*gwatTypes.ConsensusResult)
	if !ok {
		panic("not found")
	}
	return item
}

func (*testDagEngineService) GetCandidates(
	_ context.Context, slot uint64,
) *gwatTypes.CandidatesResult {
	fix := dagFixtures()
	item, ok := fix["ExecutionCandidates"].(*gwatTypes.CandidatesResult)
	if !ok {
		panic("not found")
	}
	return item
}

func (*testDagEngineService) Finalize(
	_ context.Context, _ *gwatTypes.ConsensusInfo,
) *gwatTypes.FinalizationResult {
	fix := dagFixtures()
	item, ok := fix["ExecutionFinalize"].(*gwatTypes.FinalizationResult)
	if !ok {
		panic("not found")
	}
	return item
}

func (*testDagEngineService) HeadSyncReady(
	_ context.Context, _ *gwatTypes.ConsensusInfo,
) bool {
	fix := dagFixtures()
	item, ok := fix["ExecutionHeadSyncReady"].(bool)
	if !ok {
		panic("not found")
	}
	return item
}

func (*testDagEngineService) HeadSync(
	_ context.Context, _ []gwatTypes.ConsensusInfo,
) bool {
	fix := dagFixtures()
	item, ok := fix["ExecutionHeadSync"].(bool)
	if !ok {
		panic("not found")
	}
	return item
}

package powchain

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	logTest "github.com/sirupsen/logrus/hooks/test"
	mockChain "gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/blockchain/testing"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/core/feed"
	mocks "gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/powchain/testing"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/config/params"
	pb "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/engine/v1"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/require"
	"gitlab.waterfall.network/waterfall/protocol/gwat/common"
	"gitlab.waterfall.network/waterfall/protocol/gwat/rpc"
	"google.golang.org/protobuf/proto"
)

func Test_checkTransitionConfiguration(t *testing.T) {
	params.SetupTestConfigCleanup(t)
	cfg := params.BeaconConfig().Copy()
	cfg.BellatrixForkEpoch = 0
	params.OverrideBeaconConfig(cfg)
	hook := logTest.NewGlobal()

	t.Run("context canceled", func(t *testing.T) {
		ctx := context.Background()
		m := &mocks.EngineClient{}
		m.Err = errors.New("something went wrong")

		srv := setupTransitionConfigTest(t)
		srv.cfg.stateNotifier = &mockChain.MockStateNotifier{}
		checkTransitionPollingInterval = time.Millisecond
		ctx, cancel := context.WithCancel(ctx)
		go srv.checkTransitionConfiguration(ctx, make(chan *feed.Event, 1))
		<-time.After(100 * time.Millisecond)
		cancel()
		require.LogsContain(t, hook, "Could not check configuration values")
	})
}

func TestService_handleExchangeConfigurationError(t *testing.T) {
	hook := logTest.NewGlobal()
	t.Run("clears existing service error", func(t *testing.T) {
		srv := setupTransitionConfigTest(t)
		srv.isRunning = true
		srv.runError = ErrConfigMismatch
		srv.handleExchangeConfigurationError(nil)
		require.Equal(t, true, srv.Status() == nil)
	})
	t.Run("does not clear existing service error if wrong kind", func(t *testing.T) {
		srv := setupTransitionConfigTest(t)
		srv.isRunning = true
		err := errors.New("something else went wrong")
		srv.runError = err
		srv.handleExchangeConfigurationError(nil)
		require.ErrorIs(t, err, srv.Status())
	})
	t.Run("sets service error on config mismatch", func(t *testing.T) {
		srv := setupTransitionConfigTest(t)
		srv.isRunning = true
		srv.handleExchangeConfigurationError(ErrConfigMismatch)
		require.Equal(t, ErrConfigMismatch, srv.Status())
		require.LogsContain(t, hook, configMismatchLog)
	})
	t.Run("does not set service error if unrelated problem", func(t *testing.T) {
		srv := setupTransitionConfigTest(t)
		srv.isRunning = true
		srv.handleExchangeConfigurationError(errors.New("foo"))
		require.Equal(t, true, srv.Status() == nil)
		require.LogsContain(t, hook, "Could not check configuration values")
	})
}

func setupTransitionConfigTest(t testing.TB) *Service {
	fix := fixtures()
	request, ok := fix["TransitionConfiguration"].(*pb.TransitionConfiguration)
	require.Equal(t, true, ok)
	resp, ok := proto.Clone(request).(*pb.TransitionConfiguration)
	require.Equal(t, true, ok)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		defer func() {
			require.NoError(t, r.Body.Close())
		}()

		// Change the terminal block hash.
		h := common.BytesToHash([]byte("foo"))
		resp.TerminalBlockHash = h[:]
		respJSON := map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      1,
			"result":  resp,
		}
		require.NoError(t, json.NewEncoder(w).Encode(respJSON))
	}))
	defer srv.Close()

	rpcClient, err := rpc.DialHTTP(srv.URL)
	require.NoError(t, err)
	defer rpcClient.Close()

	service := &Service{
		cfg: &config{},
	}
	service.rpcClient = rpcClient
	return service
}

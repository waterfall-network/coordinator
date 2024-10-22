package rpc

import (
	"context"
	"errors"
	"io/ioutil"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	logTest "github.com/sirupsen/logrus/hooks/test"
	mock "gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/blockchain/testing"
	mockPOW "gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/powchain/testing"
	mockSync "gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/sync/initial-sync/testing"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/assert"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/require"
)

func init() {
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetOutput(ioutil.Discard)
}

func TestLifecycle_OK(t *testing.T) {
	hook := logTest.NewGlobal()
	chainService := &mock.ChainService{
		Genesis: time.Now(),
	}
	rpcService := NewService(context.Background(), &Config{
		Port:                "7348",
		SyncService:         &mockSync.Sync{IsSyncing: false},
		BlockReceiver:       chainService,
		AttestationReceiver: chainService,
		HeadFetcher:         chainService,
		GenesisTimeFetcher:  chainService,
		POWChainService:     &mockPOW.POWChain{},
		StateNotifier:       chainService.StateNotifier(),
	})

	rpcService.Start()

	require.LogsContain(t, hook, "listening on port")
	assert.NoError(t, rpcService.Stop())
}

func TestStatus_CredentialError(t *testing.T) {
	credentialErr := errors.New("credentialError")
	s := &Service{
		cfg:             &Config{SyncService: &mockSync.Sync{IsSyncing: false}},
		credentialError: credentialErr,
	}

	assert.ErrorContains(t, s.credentialError.Error(), s.Status())
}

func TestRPC_InsecureEndpoint(t *testing.T) {
	hook := logTest.NewGlobal()
	chainService := &mock.ChainService{Genesis: time.Now()}
	rpcService := NewService(context.Background(), &Config{
		Port:                "7777",
		SyncService:         &mockSync.Sync{IsSyncing: false},
		BlockReceiver:       chainService,
		GenesisTimeFetcher:  chainService,
		AttestationReceiver: chainService,
		HeadFetcher:         chainService,
		POWChainService:     &mockPOW.POWChain{},
		StateNotifier:       chainService.StateNotifier(),
	})

	rpcService.Start()

	require.LogsContain(t, hook, "listening on port")
	require.LogsContain(t, hook, "You are using an insecure gRPC server")
	assert.NoError(t, rpcService.Stop())
}

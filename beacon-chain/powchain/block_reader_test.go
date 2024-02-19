package powchain

import (
	"context"
	"math/big"
	"testing"
	"time"

	dbutil "gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/db/testing"
	mockPOW "gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/powchain/testing"
	eth "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/assert"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/require"
	"gitlab.waterfall.network/waterfall/protocol/gwat/common"
	"gitlab.waterfall.network/waterfall/protocol/gwat/common/hexutil"
	gethTypes "gitlab.waterfall.network/waterfall/protocol/gwat/core/types"
	"gitlab.waterfall.network/waterfall/protocol/gwat/trie"
)

func setDefaultMocks(service *Service) *Service {
	service.eth1DataFetcher = &goodFetcher{}
	service.httpLogger = &goodLogger{}
	service.cfg.stateNotifier = &goodNotifier{}
	return service
}

func TestLatestMainchainInfo_OK(t *testing.T) {
	testAcc, err := mockPOW.Setup()
	require.NoError(t, err, "Unable to set up simulated backend")

	beaconDB := dbutil.SetupDB(t)
	server, endpoint, err := mockPOW.SetupRPCServer()
	require.NoError(t, err)
	t.Cleanup(func() {
		server.Stop()
	})
	web3Service, err := NewService(context.Background(),
		WithHttpEndpoints([]string{endpoint}),
		WithDatabase(beaconDB),
	)
	require.NoError(t, err, "Unable to setup web3 ETH1.0 chain service")

	web3Service = setDefaultMocks(web3Service)
	web3Service.rpcClient = &mockPOW.RPCClient{Backend: testAcc.Backend}
	web3Service.eth1DataFetcher = &goodFetcher{backend: testAcc.Backend}

	web3Service.latestEth1Data = &eth.LatestETH1Data{
		BlockHeight:        2,
		BlockTime:          20,
		BlockHash:          []byte{167, 119, 205, 247, 51, 60, 237, 42, 59, 241, 126, 94, 156, 163, 77, 216, 159, 17, 186, 52, 114, 231, 34, 81, 58, 154, 100, 205, 141, 81, 204, 145},
		LastRequestedBlock: 0,
		CpHash:             nil,
		CpNr:               0,
	}

	testAcc.Backend.Commit()

	exitRoutine := make(chan bool)

	go func() {
		web3Service.run(web3Service.ctx.Done())
		<-exitRoutine
	}()

	header, err := web3Service.eth1DataFetcher.HeaderByNumber(web3Service.ctx, nil)
	require.NoError(t, err)

	tickerChan := make(chan time.Time, 1)
	web3Service.headTicker = &time.Ticker{C: tickerChan}
	tickerChan <- time.Now()
	web3Service.cancel()
	exitRoutine <- true

	hash := header.Hash().Bytes()
	log.Infof("BYTES %+v", hash)

	assert.Equal(t, web3Service.latestEth1Data.BlockHeight, header.Nr())
	assert.Equal(t, hexutil.Encode(web3Service.latestEth1Data.BlockHash), header.Hash().Hex())
	assert.Equal(t, web3Service.latestEth1Data.BlockTime, header.Time)
}

func TestBlockHashByHeight_ReturnsHash(t *testing.T) {
	beaconDB := dbutil.SetupDB(t)
	server, endpoint, err := mockPOW.SetupRPCServer()
	require.NoError(t, err)
	t.Cleanup(func() {
		server.Stop()
	})
	web3Service, err := NewService(context.Background(),
		WithHttpEndpoints([]string{endpoint}),
		WithDatabase(beaconDB),
	)
	require.NoError(t, err, "unable to setup web3 ETH1.0 chain service")

	web3Service = setDefaultMocks(web3Service)
	ctx := context.Background()

	nr_15 := uint64(15)
	header := &gethTypes.Header{
		Number: &nr_15,
		Time:   150,
	}

	wanted := header.Hash()

	hash, err := web3Service.BlockHashByHeight(ctx, big.NewInt(0))
	require.NoError(t, err, "Could not get block hash with given height")
	require.DeepEqual(t, wanted.Bytes(), hash.Bytes(), "Block hash did not equal expected hash")

	exists, _, err := web3Service.headerCache.HeaderInfoByHash(wanted)
	require.NoError(t, err)
	require.Equal(t, true, exists, "Expected block info to be cached")
}

func TestBlockHashByHeight_ReturnsError_WhenNoEth1Client(t *testing.T) {
	beaconDB := dbutil.SetupDB(t)
	server, endpoint, err := mockPOW.SetupRPCServer()
	require.NoError(t, err)
	t.Cleanup(func() {
		server.Stop()
	})
	web3Service, err := NewService(context.Background(),
		WithHttpEndpoints([]string{endpoint}),
		WithDatabase(beaconDB),
	)
	require.NoError(t, err, "unable to setup web3 ETH1.0 chain service")

	web3Service = setDefaultMocks(web3Service)
	web3Service.eth1DataFetcher = nil
	ctx := context.Background()

	_, err = web3Service.BlockHashByHeight(ctx, big.NewInt(0))
	require.ErrorContains(t, "nil eth1DataFetcher", err)
}

func TestBlockExists_ValidHash(t *testing.T) {
	beaconDB := dbutil.SetupDB(t)
	server, endpoint, err := mockPOW.SetupRPCServer()
	require.NoError(t, err)
	t.Cleanup(func() {
		server.Stop()
	})
	web3Service, err := NewService(context.Background(),
		WithHttpEndpoints([]string{endpoint}),
		WithDatabase(beaconDB),
	)
	require.NoError(t, err, "unable to setup web3 ETH1.0 chain service")

	web3Service = setDefaultMocks(web3Service)

	nr_0 := uint64(0)
	block := gethTypes.NewBlock(
		&gethTypes.Header{
			Number: &nr_0,
		},
		[]*gethTypes.Transaction{},
		[]*gethTypes.Receipt{},
		new(trie.Trie),
	)

	exists, height, err := web3Service.BlockExists(context.Background(), block.Hash())
	require.NoError(t, err, "Could not get block hash with given height")
	require.Equal(t, true, exists)
	require.Equal(t, 0, height.Cmp(new(big.Int).SetUint64(block.Nr())))

	exists, _, err = web3Service.headerCache.HeaderInfoByHeight(height)
	require.NoError(t, err)
	require.Equal(t, true, exists, "Expected block to be cached")

}

func TestBlockExists_InvalidHash(t *testing.T) {
	beaconDB := dbutil.SetupDB(t)
	server, endpoint, err := mockPOW.SetupRPCServer()
	require.NoError(t, err)
	t.Cleanup(func() {
		server.Stop()
	})
	web3Service, err := NewService(context.Background(),
		WithHttpEndpoints([]string{endpoint}),
		WithDatabase(beaconDB),
	)
	require.NoError(t, err, "unable to setup web3 ETH1.0 chain service")

	web3Service = setDefaultMocks(web3Service)

	_, _, err = web3Service.BlockExists(context.Background(), common.BytesToHash([]byte{0}))
	require.NotNil(t, err, "Expected BlockExists to error with invalid hash")
}

func TestBlockExists_UsesCachedBlockInfo(t *testing.T) {
	beaconDB := dbutil.SetupDB(t)
	server, endpoint, err := mockPOW.SetupRPCServer()
	require.NoError(t, err)
	t.Cleanup(func() {
		server.Stop()
	})
	web3Service, err := NewService(context.Background(),
		WithHttpEndpoints([]string{endpoint}),
		WithDatabase(beaconDB),
	)
	require.NoError(t, err, "unable to setup web3 ETH1.0 chain service")
	// nil eth1DataFetcher would panic if cached value not used
	//web3Service.eth1DataFetcher = nil
	nr_0 := uint64(1)
	header := &gethTypes.Header{
		Number: &nr_0,
	}

	err = web3Service.headerCache.AddHeader(header)
	require.NoError(t, err)

	exists, height, err := web3Service.BlockExists(context.Background(), header.Hash())
	require.NoError(t, err, "Could not get block hash with given height")
	require.Equal(t, true, exists)
	require.Equal(t, 0, height.Cmp(new(big.Int).SetUint64(header.Nr())))
}

func TestBlockExistsWithCache_UsesCachedHeaderInfo(t *testing.T) {
	beaconDB := dbutil.SetupDB(t)
	server, endpoint, err := mockPOW.SetupRPCServer()
	require.NoError(t, err)
	t.Cleanup(func() {
		server.Stop()
	})
	web3Service, err := NewService(context.Background(),
		WithHttpEndpoints([]string{endpoint}),
		WithDatabase(beaconDB),
	)
	require.NoError(t, err, "unable to setup web3 ETH1.0 chain service")

	nr_0 := uint64(0)
	header := &gethTypes.Header{
		Number: &nr_0,
	}

	err = web3Service.headerCache.AddHeader(header)
	require.NoError(t, err)

	exists, height, err := web3Service.BlockExistsWithCache(context.Background(), header.Hash())
	require.NoError(t, err, "Could not get block hash with given height")
	require.Equal(t, true, exists)
	require.Equal(t, 0, height.Cmp(new(big.Int).SetUint64(header.Nr())))
}

func TestBlockExistsWithCache_HeaderNotCached(t *testing.T) {
	beaconDB := dbutil.SetupDB(t)
	server, endpoint, err := mockPOW.SetupRPCServer()
	require.NoError(t, err)
	t.Cleanup(func() {
		server.Stop()
	})
	web3Service, err := NewService(context.Background(),
		WithHttpEndpoints([]string{endpoint}),
		WithDatabase(beaconDB),
	)
	require.NoError(t, err, "unable to setup web3 ETH1.0 chain service")

	exists, height, err := web3Service.BlockExistsWithCache(context.Background(), common.BytesToHash([]byte("hash")))
	require.NoError(t, err, "Could not get block hash with given height")
	require.Equal(t, false, exists)
	require.Equal(t, (*big.Int)(nil), height)
}

func TestService_BlockTimeByHeight_ReturnsError_WhenNoEth1Client(t *testing.T) {
	beaconDB := dbutil.SetupDB(t)
	server, endpoint, err := mockPOW.SetupRPCServer()
	require.NoError(t, err)
	t.Cleanup(func() {
		server.Stop()
	})
	web3Service, err := NewService(context.Background(),
		WithHttpEndpoints([]string{endpoint}),
		WithDatabase(beaconDB),
	)
	require.NoError(t, err, "unable to setup web3 ETH1.0 chain service")

	web3Service = setDefaultMocks(web3Service)
	web3Service.eth1DataFetcher = nil
	ctx := context.Background()

	_, err = web3Service.BlockTimeByHeight(ctx, big.NewInt(0))
	require.ErrorContains(t, "nil eth1DataFetcher", err)
}

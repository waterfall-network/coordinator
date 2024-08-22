package p2p

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/prysmaticlabs/go-bitfield"
	mock "gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/blockchain/testing"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/cache"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/core/feed"
	statefeed "gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/core/feed/state"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/cmd/beacon-chain/flags"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/config/params"
	pb "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1/wrapper"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/assert"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/require"
	"gitlab.waterfall.network/waterfall/protocol/gwat/p2p/discover"
	"gitlab.waterfall.network/waterfall/protocol/gwat/p2p/enode"
	"gitlab.waterfall.network/waterfall/protocol/gwat/p2p/enr"
)

func TestStartDiscV5_DiscoverPeersWithSubnets(t *testing.T) {
	//// This test needs to be entirely rewritten and should be done in a follow up PR from #7885.
	//t.Skip("This test is now failing after PR 7885 due to false positive")
	gFlags := new(flags.GlobalFlags)
	gFlags.MinimumPeersPerSubnet = 4
	flags.Init(gFlags)
	// Reset config.
	defer flags.Init(new(flags.GlobalFlags))
	port := 2000
	ipAddr, pkey := createAddrAndPrivKey(t)
	genesisTime := time.Now()
	genesisValidatorsRoot := make([]byte, 32)
	s := &Service{
		cfg:                   &Config{UDPPort: uint(port)},
		genesisTime:           genesisTime,
		genesisValidatorsRoot: genesisValidatorsRoot,
	}
	bootListener, err := s.createListener(ipAddr, pkey)
	require.NoError(t, err)
	defer bootListener.Close()

	bootNode := bootListener.Self()
	// Use shorter period for testing.
	currentPeriod := pollingPeriod
	pollingPeriod = 1 * time.Second
	defer func() {
		pollingPeriod = currentPeriod
	}()

	var listeners []*discover.UDPv5
	for i := 1; i <= 3; i++ {
		port = 3000 + i
		cfg := &Config{
			BootstrapNodeAddr:   []string{bootNode.String()},
			Discv5BootStrapAddr: []string{bootNode.String()},
			MaxPeers:            30,
			UDPPort:             uint(port),
		}
		ipAddr, pkey := createAddrAndPrivKey(t)
		s = &Service{
			cfg:                   cfg,
			genesisTime:           genesisTime,
			genesisValidatorsRoot: genesisValidatorsRoot,
		}
		listener, err := s.startDiscoveryV5(ipAddr, pkey)
		assert.NoError(t, err, "Could not start discovery for node")
		bitV := bitfield.NewBitvector64()
		bitV.SetBitAt(uint64(i), true)

		entry := enr.WithEntry(attSubnetEnrKey, &bitV)
		listener.LocalNode().Set(entry)
		listeners = append(listeners, listener)
	}
	defer func() {
		// Close down all peers.
		for _, listener := range listeners {
			listener.Close()
		}
	}()

	// Make one service on port 4001.
	port = 4001
	cfg := &Config{
		BootstrapNodeAddr:   []string{bootNode.String()},
		Discv5BootStrapAddr: []string{bootNode.String()},
		MaxPeers:            30,
		UDPPort:             uint(port),
	}
	cfg.StateNotifier = &mock.MockStateNotifier{}
	s, err = NewService(context.Background(), cfg)
	require.NoError(t, err)
	exitRoutine := make(chan bool)
	go func() {
		s.Start()
		<-exitRoutine
	}()
	time.Sleep(50 * time.Millisecond)
	// Send in a loop to ensure it is delivered (busy wait for the service to subscribe to the state feed).
	for sent := 0; sent == 0; {
		sent = s.stateNotifier.StateFeed().Send(&feed.Event{
			Type: statefeed.Initialized,
			Data: &statefeed.InitializedData{
				StartTime:             time.Now(),
				GenesisValidatorsRoot: make([]byte, 32),
			},
		})
	}

	// Wait for the nodes to have their local routing tables to be populated with the other nodes
	time.Sleep(6 * discoveryWaitTime)

	// look up 3 different subnets
	ctx := context.Background()
	exists, err := s.FindPeersWithSubnet(ctx, GossipAttestationMessage, 1, flags.Get().MinimumPeersPerSubnet)
	require.NoError(t, err)
	exists2, err := s.FindPeersWithSubnet(ctx, GossipAttestationMessage, 2, flags.Get().MinimumPeersPerSubnet)
	require.NoError(t, err)
	exists3, err := s.FindPeersWithSubnet(ctx, GossipAttestationMessage, 3, flags.Get().MinimumPeersPerSubnet)
	require.NoError(t, err)
	if !exists || !exists2 || !exists3 {
		t.Fatal("Peer with subnet doesn't exist")
	}

	// Update ENR of a peer.
	testService := &Service{
		dv5Listener: listeners[0],
		metaData: wrapper.WrappedMetadataV0(&pb.MetaDataV0{
			Attnets: bitfield.NewBitvector64(),
		}),
	}
	cache.SubnetIDs.AddAttesterSubnetID(0, 10)
	testService.RefreshENR()
	time.Sleep(2 * time.Second)

	exists, err = s.FindPeersWithSubnet(ctx, GossipAttestationMessage, 2, flags.Get().MinimumPeersPerSubnet)
	require.NoError(t, err)

	assert.Equal(t, true, exists, "Peer with subnet doesn't exist")
	assert.NoError(t, s.Stop())
	exitRoutine <- true
}

func TestStartDiscV5_FindPeersWithSubnet(t *testing.T) {
	// Topology of this test:
	//
	//
	// Node 1 (subscribed to subnet 1)  --\
	//									  |
	// Node 2 (subscribed to subnet 2)  --+--> BootNode (not subscribed to any subnet) <------- Node 0 (not subscribed to any subnet)
	//									  |
	// Node 3 (subscribed to subnet 3)  --/
	//
	// The purpose of this test is to ensure that the "Node 0" (connected only to the boot node) is able to
	// find and connect to a node already subscribed to a specific subnet.
	// In our case: The node i is subscribed to subnet i, with i = 1, 2, 3

	// Define the genesis validators root, to ensure everybody is on the same network.
	const genesisValidatorRootStr = "0xdeadbeefcafecafedeadbeefcafecafedeadbeefcafecafedeadbeefcafecafe"
	genesisValidatorsRoot, err := hex.DecodeString(genesisValidatorRootStr[2:])
	require.NoError(t, err)

	// Create a context.
	ctx := context.Background()

	// Use shorter period for testing.
	currentPeriod := pollingPeriod
	pollingPeriod = 1 * time.Second
	defer func() {
		pollingPeriod = currentPeriod
	}()

	// Create flags.
	params.SetupTestConfigCleanup(t)
	gFlags := new(flags.GlobalFlags)
	gFlags.MinimumPeersPerSubnet = 1
	flags.Init(gFlags)

	params.BeaconNetworkConfig().MinimumPeersInSubnetSearch = 1

	// Reset config.
	defer flags.Init(new(flags.GlobalFlags))

	// First, generate a bootstrap node.
	ipAddr, pkey := createAddrAndPrivKey(t)
	genesisTime := time.Now()

	bootNodeService := &Service{
		cfg:                   &Config{TCPPort: 2000, UDPPort: 3000},
		genesisTime:           genesisTime,
		genesisValidatorsRoot: genesisValidatorsRoot,
	}

	bootNodeForkDigest, err := bootNodeService.currentForkDigest()
	require.NoError(t, err)

	bootListener, err := bootNodeService.createListener(ipAddr, pkey)
	require.NoError(t, err)
	defer bootListener.Close()

	bootNodeENR := bootListener.Self().String()

	// Create 3 nodes, each subscribed to a different subnet.
	// Each node is connected to the boostrap node.
	services := make([]*Service, 0, 3)

	for i := 1; i <= 3; i++ {
		subnet := uint64(i)
		service, err := NewService(ctx, &Config{
			Discv5BootStrapAddr: []string{bootNodeENR},
			//BootstrapNodeAddr:   []string{bootNodeENR},
			MaxPeers: 30,
			TCPPort:  uint(2000 + i),
			UDPPort:  uint(3000 + i),
		})

		require.NoError(t, err)

		service.genesisTime = genesisTime
		service.genesisValidatorsRoot = genesisValidatorsRoot

		nodeForkDigest, err := service.currentForkDigest()
		require.NoError(t, err)
		require.Equal(t, true, nodeForkDigest == bootNodeForkDigest, "fork digest of the node doesn't match the boot node")

		// Start the service.
		service.Start()

		// Set the ENR `attnets`, used by Prysm to filter peers by subnet.
		bitV := bitfield.NewBitvector64()
		bitV.SetBitAt(subnet, true)
		entry := enr.WithEntry(attSubnetEnrKey, &bitV)
		service.dv5Listener.LocalNode().Set(entry)

		// Join and subscribe to the subnet, needed by libp2p.
		topic, err := service.pubsub.Join(fmt.Sprintf(AttestationSubnetTopicFormat, bootNodeForkDigest, subnet) + "/ssz_snappy")
		require.NoError(t, err)

		_, err = topic.Subscribe()
		require.NoError(t, err)

		// Memoize the service.
		services = append(services, service)
	}

	// Stop the services.
	defer func() {
		for _, service := range services {
			err := service.Stop()
			require.NoError(t, err)
		}
	}()

	cfg := &Config{
		Discv5BootStrapAddr: []string{bootNodeENR},
		//BootstrapNodeAddr:   []string{bootNodeENR},
		MaxPeers: 30,
		TCPPort:  2010,
		UDPPort:  3010,
	}

	service, err := NewService(ctx, cfg)
	require.NoError(t, err)

	service.genesisTime = genesisTime
	service.genesisValidatorsRoot = genesisValidatorsRoot

	service.Start()
	defer func() {
		err := service.Stop()
		require.NoError(t, err)
	}()

	// Look up 3 different subnets.
	exists := make([]bool, 0, 3)
	for i := 1; i <= 3; i++ {
		subnet := uint64(i)
		topic := fmt.Sprintf(AttestationSubnetTopicFormat, bootNodeForkDigest, subnet)

		exist := false

		// This for loop is used to ensure we don't get stuck in `FindPeersWithSubnet`.
		// Read the documentation of `FindPeersWithSubnet` for more details.
		for j := 0; j < 3; j++ {
			ctxWithTimeOut, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			exist, err = service.FindPeersWithSubnet(ctxWithTimeOut, topic, subnet, 1)
			require.NoError(t, err)

			if exist {
				break
			}
		}

		require.NoError(t, err)
		exists = append(exists, exist)

	}

	// Check if all peers are found.
	for _, exist := range exists {
		require.Equal(t, true, exist, "Peer with subnet doesn't exist")
	}
}

func Test_AttSubnets(t *testing.T) {
	params.SetupTestConfigCleanup(t)
	tests := []struct {
		name        string
		record      func(t *testing.T) *enr.Record
		want        map[uint64]bool
		wantErr     bool
		errContains string
	}{
		{
			name: "valid record",
			record: func(t *testing.T) *enr.Record {
				db, err := enode.OpenDB("")
				assert.NoError(t, err)
				priv, _, err := crypto.GenerateSecp256k1Key(rand.Reader)
				assert.NoError(t, err)
				convertedKey, err := ConvertFromInterfacePrivKey(priv)
				assert.NoError(t, err)
				localNode := enode.NewLocalNode(db, convertedKey)
				localNode = initializeAttSubnets(localNode)
				return localNode.Node().Record()
			},
			want:    map[uint64]bool{},
			wantErr: false,
		},
		{
			name: "too small subnet",
			record: func(t *testing.T) *enr.Record {
				db, err := enode.OpenDB("")
				assert.NoError(t, err)
				priv, _, err := crypto.GenerateSecp256k1Key(rand.Reader)
				assert.NoError(t, err)
				convertedKey, err := ConvertFromInterfacePrivKey(priv)
				assert.NoError(t, err)
				localNode := enode.NewLocalNode(db, convertedKey)
				entry := enr.WithEntry(attSubnetEnrKey, []byte{})
				localNode.Set(entry)
				return localNode.Node().Record()
			},
			want:        map[uint64]bool{},
			wantErr:     true,
			errContains: "invalid bitvector provided, it has a size of",
		},
		{
			name: "half sized subnet",
			record: func(t *testing.T) *enr.Record {
				db, err := enode.OpenDB("")
				assert.NoError(t, err)
				priv, _, err := crypto.GenerateSecp256k1Key(rand.Reader)
				assert.NoError(t, err)
				convertedKey, err := ConvertFromInterfacePrivKey(priv)
				assert.NoError(t, err)
				localNode := enode.NewLocalNode(db, convertedKey)
				entry := enr.WithEntry(attSubnetEnrKey, make([]byte, 4))
				localNode.Set(entry)
				return localNode.Node().Record()
			},
			want:        map[uint64]bool{},
			wantErr:     true,
			errContains: "invalid bitvector provided, it has a size of",
		},
		{
			name: "too large subnet",
			record: func(t *testing.T) *enr.Record {
				db, err := enode.OpenDB("")
				assert.NoError(t, err)
				priv, _, err := crypto.GenerateSecp256k1Key(rand.Reader)
				assert.NoError(t, err)
				convertedKey, err := ConvertFromInterfacePrivKey(priv)
				assert.NoError(t, err)
				localNode := enode.NewLocalNode(db, convertedKey)
				entry := enr.WithEntry(attSubnetEnrKey, make([]byte, byteCount(int(attestationSubnetCount))+1))
				localNode.Set(entry)
				return localNode.Node().Record()
			},
			want:        map[uint64]bool{},
			wantErr:     true,
			errContains: "invalid bitvector provided, it has a size of",
		},
		{
			name: "very large subnet",
			record: func(t *testing.T) *enr.Record {
				db, err := enode.OpenDB("")
				assert.NoError(t, err)
				priv, _, err := crypto.GenerateSecp256k1Key(rand.Reader)
				assert.NoError(t, err)
				convertedKey, err := ConvertFromInterfacePrivKey(priv)
				assert.NoError(t, err)
				localNode := enode.NewLocalNode(db, convertedKey)
				entry := enr.WithEntry(attSubnetEnrKey, make([]byte, byteCount(int(attestationSubnetCount))+100))
				localNode.Set(entry)
				return localNode.Node().Record()
			},
			want:        map[uint64]bool{},
			wantErr:     true,
			errContains: "invalid bitvector provided, it has a size of",
		},
		{
			name: "single subnet",
			record: func(t *testing.T) *enr.Record {
				db, err := enode.OpenDB("")
				assert.NoError(t, err)
				priv, _, err := crypto.GenerateSecp256k1Key(rand.Reader)
				assert.NoError(t, err)
				convertedKey, err := ConvertFromInterfacePrivKey(priv)
				assert.NoError(t, err)
				localNode := enode.NewLocalNode(db, convertedKey)
				bitV := bitfield.NewBitvector64()
				bitV.SetBitAt(0, true)
				entry := enr.WithEntry(attSubnetEnrKey, bitV.Bytes())
				localNode.Set(entry)
				return localNode.Node().Record()
			},
			want:    map[uint64]bool{0: true},
			wantErr: false,
		},
		{
			name: "multiple subnets",
			record: func(t *testing.T) *enr.Record {
				db, err := enode.OpenDB("")
				assert.NoError(t, err)
				priv, _, err := crypto.GenerateSecp256k1Key(rand.Reader)
				assert.NoError(t, err)
				convertedKey, err := ConvertFromInterfacePrivKey(priv)
				assert.NoError(t, err)
				localNode := enode.NewLocalNode(db, convertedKey)
				bitV := bitfield.NewBitvector64()
				for i := uint64(0); i < bitV.Len(); i++ {
					// skip 2 subnets
					if (i+1)%2 == 0 {
						continue
					}
					bitV.SetBitAt(i, true)
				}
				bitV.SetBitAt(0, true)
				entry := enr.WithEntry(attSubnetEnrKey, bitV.Bytes())
				localNode.Set(entry)
				return localNode.Node().Record()
			},
			want: map[uint64]bool{0: true, 2: true, 4: true, 6: true, 8: true, 10: true, 12: true, 14: true, 16: true, 18: true,
				20: true, 22: true, 24: true, 26: true, 28: true, 30: true, 32: true, 34: true, 36: true, 38: true,
				40: true, 42: true, 44: true, 46: true, 48: true, 50: true, 52: true, 54: true, 56: true, 58: true,
				60: true, 62: true},
			wantErr: false,
		},
		{
			name: "all subnets",
			record: func(t *testing.T) *enr.Record {
				db, err := enode.OpenDB("")
				assert.NoError(t, err)
				priv, _, err := crypto.GenerateSecp256k1Key(rand.Reader)
				assert.NoError(t, err)
				convertedKey, err := ConvertFromInterfacePrivKey(priv)
				assert.NoError(t, err)
				localNode := enode.NewLocalNode(db, convertedKey)
				bitV := bitfield.NewBitvector64()
				for i := uint64(0); i < bitV.Len(); i++ {
					bitV.SetBitAt(i, true)
				}
				entry := enr.WithEntry(attSubnetEnrKey, bitV.Bytes())
				localNode.Set(entry)
				return localNode.Node().Record()
			},
			want: map[uint64]bool{0: true, 1: true, 2: true, 3: true, 4: true, 5: true, 6: true, 7: true, 8: true, 9: true,
				10: true, 11: true, 12: true, 13: true, 14: true, 15: true, 16: true, 17: true, 18: true, 19: true,
				20: true, 21: true, 22: true, 23: true, 24: true, 25: true, 26: true, 27: true, 28: true, 29: true,
				30: true, 31: true, 32: true, 33: true, 34: true, 35: true, 36: true, 37: true, 38: true, 39: true,
				40: true, 41: true, 42: true, 43: true, 44: true, 45: true, 46: true, 47: true, 48: true, 49: true,
				50: true, 51: true, 52: true, 53: true, 54: true, 55: true, 56: true, 57: true, 58: true, 59: true,
				60: true, 61: true, 62: true, 63: true},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := attSubnets(tt.record(t))
			if (err != nil) != tt.wantErr {
				t.Errorf("syncSubnets() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				assert.ErrorContains(t, tt.errContains, err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("syncSubnets() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_SyncSubnets(t *testing.T) {
	params.SetupTestConfigCleanup(t)
	tests := []struct {
		name        string
		record      func(t *testing.T) *enr.Record
		want        []uint64
		wantErr     bool
		errContains string
	}{
		{
			name: "valid record",
			record: func(t *testing.T) *enr.Record {
				db, err := enode.OpenDB("")
				assert.NoError(t, err)
				priv, _, err := crypto.GenerateSecp256k1Key(rand.Reader)
				assert.NoError(t, err)
				convertedKey, err := ConvertFromInterfacePrivKey(priv)
				assert.NoError(t, err)
				localNode := enode.NewLocalNode(db, convertedKey)
				localNode = initializeSyncCommSubnets(localNode)
				return localNode.Node().Record()
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "too small subnet",
			record: func(t *testing.T) *enr.Record {
				db, err := enode.OpenDB("")
				assert.NoError(t, err)
				priv, _, err := crypto.GenerateSecp256k1Key(rand.Reader)
				assert.NoError(t, err)
				convertedKey, err := ConvertFromInterfacePrivKey(priv)
				assert.NoError(t, err)
				localNode := enode.NewLocalNode(db, convertedKey)
				entry := enr.WithEntry(syncCommsSubnetEnrKey, []byte{})
				localNode.Set(entry)
				return localNode.Node().Record()
			},
			want:        []uint64{},
			wantErr:     true,
			errContains: "invalid bitvector provided, it has a size of",
		},
		{
			name: "too large subnet",
			record: func(t *testing.T) *enr.Record {
				db, err := enode.OpenDB("")
				assert.NoError(t, err)
				priv, _, err := crypto.GenerateSecp256k1Key(rand.Reader)
				assert.NoError(t, err)
				convertedKey, err := ConvertFromInterfacePrivKey(priv)
				assert.NoError(t, err)
				localNode := enode.NewLocalNode(db, convertedKey)
				entry := enr.WithEntry(syncCommsSubnetEnrKey, make([]byte, byteCount(int(syncCommsSubnetCount))+1))
				localNode.Set(entry)
				return localNode.Node().Record()
			},
			want:        []uint64{},
			wantErr:     true,
			errContains: "invalid bitvector provided, it has a size of",
		},
		{
			name: "very large subnet",
			record: func(t *testing.T) *enr.Record {
				db, err := enode.OpenDB("")
				assert.NoError(t, err)
				priv, _, err := crypto.GenerateSecp256k1Key(rand.Reader)
				assert.NoError(t, err)
				convertedKey, err := ConvertFromInterfacePrivKey(priv)
				assert.NoError(t, err)
				localNode := enode.NewLocalNode(db, convertedKey)
				entry := enr.WithEntry(syncCommsSubnetEnrKey, make([]byte, byteCount(int(syncCommsSubnetCount))+100))
				localNode.Set(entry)
				return localNode.Node().Record()
			},
			want:        []uint64{},
			wantErr:     true,
			errContains: "invalid bitvector provided, it has a size of",
		},
		{
			name: "single subnet",
			record: func(t *testing.T) *enr.Record {
				db, err := enode.OpenDB("")
				assert.NoError(t, err)
				priv, _, err := crypto.GenerateSecp256k1Key(rand.Reader)
				assert.NoError(t, err)
				convertedKey, err := ConvertFromInterfacePrivKey(priv)
				assert.NoError(t, err)
				localNode := enode.NewLocalNode(db, convertedKey)
				bitV := bitfield.Bitvector4{byte(0x00)}
				bitV.SetBitAt(0, true)
				entry := enr.WithEntry(syncCommsSubnetEnrKey, bitV.Bytes())
				localNode.Set(entry)
				return localNode.Node().Record()
			},
			want:    []uint64{0},
			wantErr: false,
		},
		{
			name: "multiple subnets",
			record: func(t *testing.T) *enr.Record {
				db, err := enode.OpenDB("")
				assert.NoError(t, err)
				priv, _, err := crypto.GenerateSecp256k1Key(rand.Reader)
				assert.NoError(t, err)
				convertedKey, err := ConvertFromInterfacePrivKey(priv)
				assert.NoError(t, err)
				localNode := enode.NewLocalNode(db, convertedKey)
				bitV := bitfield.Bitvector4{byte(0x00)}
				for i := uint64(0); i < bitV.Len(); i++ {
					// skip 2 subnets
					if (i+1)%2 == 0 {
						continue
					}
					bitV.SetBitAt(i, true)
				}
				bitV.SetBitAt(0, true)
				entry := enr.WithEntry(syncCommsSubnetEnrKey, bitV.Bytes())
				localNode.Set(entry)
				return localNode.Node().Record()
			},
			want:    []uint64{0, 2},
			wantErr: false,
		},
		{
			name: "all subnets",
			record: func(t *testing.T) *enr.Record {
				db, err := enode.OpenDB("")
				assert.NoError(t, err)
				priv, _, err := crypto.GenerateSecp256k1Key(rand.Reader)
				assert.NoError(t, err)
				convertedKey, err := ConvertFromInterfacePrivKey(priv)
				assert.NoError(t, err)
				localNode := enode.NewLocalNode(db, convertedKey)
				bitV := bitfield.Bitvector4{byte(0x00)}
				for i := uint64(0); i < bitV.Len(); i++ {
					bitV.SetBitAt(i, true)
				}
				entry := enr.WithEntry(syncCommsSubnetEnrKey, bitV.Bytes())
				localNode.Set(entry)
				return localNode.Node().Record()
			},
			want:    []uint64{0, 1, 2, 3},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := syncSubnets(tt.record(t))
			if (err != nil) != tt.wantErr {
				t.Errorf("syncSubnets() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				assert.ErrorContains(t, tt.errContains, err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("syncSubnets() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSubnetComputation(t *testing.T) {
	db, err := enode.OpenDB("")
	assert.NoError(t, err)
	defer db.Close()
	priv, _, err := crypto.GenerateSecp256k1Key(rand.Reader)
	assert.NoError(t, err)
	convertedKey, err := ConvertFromInterfacePrivKey(priv)
	assert.NoError(t, err)
	localNode := enode.NewLocalNode(db, convertedKey)

	retrievedSubnets, err := computeSubscribedSubnets(localNode.ID(), 1000)
	assert.NoError(t, err)
	assert.Equal(t, retrievedSubnets[0]+1, retrievedSubnets[1])
}

func TestInitializePersistentSubnets(t *testing.T) {
	cache.SubnetIDs.EmptyAllCaches()
	defer cache.SubnetIDs.EmptyAllCaches()

	db, err := enode.OpenDB("")
	assert.NoError(t, err)
	defer db.Close()
	priv, _, err := crypto.GenerateSecp256k1Key(rand.Reader)
	assert.NoError(t, err)
	convertedKey, err := ConvertFromInterfacePrivKey(priv)
	assert.NoError(t, err)
	localNode := enode.NewLocalNode(db, convertedKey)

	assert.NoError(t, initializePersistentSubnets(localNode.ID(), 10000))
	subs, ok, expTime := cache.SubnetIDs.GetPersistentSubnets()
	assert.Equal(t, true, ok)
	assert.Equal(t, 2, len(subs))
	assert.Equal(t, true, expTime.After(time.Now()))
}

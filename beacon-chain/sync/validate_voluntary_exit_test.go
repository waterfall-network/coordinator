package sync

import (
	"bytes"
	"context"
	"crypto/rand"
	"reflect"
	"testing"
	"time"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	pubsubpb "github.com/libp2p/go-libp2p-pubsub/pb"
	mock "gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/blockchain/testing"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/core/feed"
	opfeed "gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/core/feed/operation"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/p2p"
	p2ptest "gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/p2p/testing"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/state"
	v1 "gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/state/v1"
	mockSync "gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/sync/initial-sync/testing"
	lruwrpr "gitlab.waterfall.network/waterfall/protocol/coordinator/cache/lru"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/config/params"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/crypto/bls"
	ethpb "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/assert"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/require"
)

func setupValidExit(t *testing.T) (*ethpb.VoluntaryExit, state.BeaconState) {
	exit := &ethpb.VoluntaryExit{
		ValidatorIndex: 0,
		Epoch:          1 + params.BeaconConfig().ShardCommitteePeriod,
	}
	registry := []*ethpb.Validator{
		{
			ExitEpoch:       params.BeaconConfig().FarFutureEpoch,
			ActivationEpoch: 0,
		},
	}
	state, err := v1.InitializeFromProto(&ethpb.BeaconState{
		Validators: registry,
		Fork: &ethpb.Fork{
			CurrentVersion:  params.BeaconConfig().GenesisForkVersion,
			PreviousVersion: params.BeaconConfig().GenesisForkVersion,
		},
		Slot: params.BeaconConfig().SlotsPerEpoch * 5,
	})
	require.NoError(t, err)
	err = state.SetSlot(state.Slot() + params.BeaconConfig().SlotsPerEpoch.Mul(uint64(params.BeaconConfig().ShardCommitteePeriod)))
	require.NoError(t, err)

	priv, err := bls.RandKey()
	require.NoError(t, err)

	val, err := state.ValidatorAtIndex(0)
	require.NoError(t, err)
	val.PublicKey = priv.PublicKey().Marshal()
	require.NoError(t, state.UpdateValidatorAtIndex(0, val))

	b := make([]byte, 32)
	_, err = rand.Read(b)
	require.NoError(t, err)

	return exit, state
}

func TestValidateVoluntaryExit_ValidExit(t *testing.T) {
	p := p2ptest.NewTestP2P(t)
	ctx := context.Background()

	exit, s := setupValidExit(t)

	r := &Service{
		cfg: &config{
			p2p: p,
			chain: &mock.ChainService{
				State:   s,
				Genesis: time.Now(),
			},
			initialSync:       &mockSync.Sync{IsSyncing: false},
			operationNotifier: (&mock.ChainService{}).OperationNotifier(),
		},
		seenExitCache: lruwrpr.New(10),
	}

	buf := new(bytes.Buffer)
	_, err := p.Encoding().EncodeGossip(buf, exit)
	require.NoError(t, err)
	topic := p2p.GossipTypeMapping[reflect.TypeOf(exit)]
	d, err := r.currentForkDigest()
	assert.NoError(t, err)
	topic = r.addDigestToTopic(topic, d)
	m := &pubsub.Message{
		Message: &pubsubpb.Message{
			Data:  buf.Bytes(),
			Topic: &topic,
		},
	}

	// Subscribe to operation notifications.
	opChannel := make(chan *feed.Event, 1)
	opSub := r.cfg.operationNotifier.OperationFeed().Subscribe(opChannel)
	defer opSub.Unsubscribe()

	res, err := r.validateVoluntaryExit(ctx, "", m)
	assert.NoError(t, err)
	valid := res == pubsub.ValidationAccept
	assert.Equal(t, true, valid, "Failed validation")
	assert.NotNil(t, m.ValidatorData, "Decoded message was not set on the message validator data")

	// Ensure the state notification was broadcast.
	notificationFound := false
	for !notificationFound {
		select {
		case event := <-opChannel:
			if event.Type == opfeed.ExitReceived {
				notificationFound = true
				_, ok := event.Data.(*opfeed.ExitReceivedData)
				assert.Equal(t, true, ok, "Entity is not of type *opfeed.ExitReceivedData")
			}
		case <-opSub.Err():
			t.Error("Subscription to state notifier failed")
			return
		}
	}
}

func TestValidateVoluntaryExit_InvalidExitSlot(t *testing.T) {
	p := p2ptest.NewTestP2P(t)
	ctx := context.Background()

	exit, s := setupValidExit(t)
	// Set state slot to 1 to cause exit object fail to verify.
	require.NoError(t, s.SetSlot(1))
	r := &Service{
		cfg: &config{
			p2p: p,
			chain: &mock.ChainService{
				State: s,
			},
			initialSync: &mockSync.Sync{IsSyncing: false},
		},
		seenExitCache: lruwrpr.New(10),
	}

	buf := new(bytes.Buffer)
	_, err := p.Encoding().EncodeGossip(buf, exit)
	require.NoError(t, err)
	topic := p2p.GossipTypeMapping[reflect.TypeOf(exit)]
	m := &pubsub.Message{
		Message: &pubsubpb.Message{
			Data:  buf.Bytes(),
			Topic: &topic,
		},
	}
	res, err := r.validateVoluntaryExit(ctx, "", m)
	_ = err
	valid := res == pubsub.ValidationAccept
	assert.Equal(t, false, valid, "passed validation")
}

func TestValidateVoluntaryExit_ValidExit_Syncing(t *testing.T) {
	p := p2ptest.NewTestP2P(t)
	ctx := context.Background()

	exit, s := setupValidExit(t)

	r := &Service{
		cfg: &config{
			p2p: p,
			chain: &mock.ChainService{
				State: s,
			},
			initialSync: &mockSync.Sync{IsSyncing: true},
		},
	}
	buf := new(bytes.Buffer)
	_, err := p.Encoding().EncodeGossip(buf, exit)
	require.NoError(t, err)
	topic := p2p.GossipTypeMapping[reflect.TypeOf(exit)]
	m := &pubsub.Message{
		Message: &pubsubpb.Message{
			Data:  buf.Bytes(),
			Topic: &topic,
		},
	}
	res, err := r.validateVoluntaryExit(ctx, "", m)
	_ = err
	valid := res == pubsub.ValidationAccept
	assert.Equal(t, false, valid, "Validation should have failed")
}

func TestValidateVoluntaryExit_Optimistic(t *testing.T) {
	p := p2ptest.NewTestP2P(t)
	ctx := context.Background()

	exit, s := setupValidExit(t)

	r := &Service{
		cfg: &config{
			p2p: p,
			chain: &mock.ChainService{
				State:      s,
				Optimistic: true,
			},
			initialSync: &mockSync.Sync{IsSyncing: false},
		},
	}
	buf := new(bytes.Buffer)
	_, err := p.Encoding().EncodeGossip(buf, exit)
	require.NoError(t, err)
	topic := p2p.GossipTypeMapping[reflect.TypeOf(exit)]
	m := &pubsub.Message{
		Message: &pubsubpb.Message{
			Data:  buf.Bytes(),
			Topic: &topic,
		},
	}
	res, err := r.validateVoluntaryExit(ctx, "", m)
	assert.NoError(t, err)
	valid := res == pubsub.ValidationIgnore
	assert.Equal(t, true, valid, "Validation should have ignored the message")
}

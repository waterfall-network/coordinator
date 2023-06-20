package prevote

import (
	"github.com/patrickmn/go-cache"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/config/params"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/crypto/hash"
	ethpb "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1"
	"sync"
	"time"
)

var hashFn = hash.HashProto

// PrevoteCache defines the caches used to satisfy prevote pool interface.
// These caches are KV store for prevotes
type PrevoteCache struct {
	prevoteCacheLock sync.RWMutex
	prevoteCache     map[[32]byte]*ethpb.PreVote
	seenPrevote      *cache.Cache
}

func NewPrevoteCache() *PrevoteCache {
	secsInEpoch := time.Duration(params.BeaconConfig().SlotsPerEpoch.Mul(params.BeaconConfig().SecondsPerSlot))
	c := cache.New(secsInEpoch*time.Second, 2*secsInEpoch*time.Second)
	pool := &PrevoteCache{
		prevoteCache: make(map[[32]byte]*ethpb.PreVote),
		seenPrevote:  c,
	}

	return pool
}

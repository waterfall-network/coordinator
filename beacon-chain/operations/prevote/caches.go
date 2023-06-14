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

// PrevoteCaches defines the caches used to satisfy prevote pool interface.
// These caches are KV store for various prevotes
// such are unaggregated, aggregated
type PrevoteCaches struct {
	aggregatedPrevoteLock  sync.RWMutex
	aggregatedPrevote      map[[32]byte][]*ethpb.PreVote
	unAggregatePrevoteLock sync.RWMutex
	unAggregatedPrevote    map[[32]byte]*ethpb.PreVote
	seenPrevote            *cache.Cache
}

func NewPrevoteCaches() *PrevoteCaches {
	secsInEpoch := time.Duration(params.BeaconConfig().SlotsPerEpoch.Mul(params.BeaconConfig().SecondsPerSlot))
	c := cache.New(secsInEpoch*time.Second, 2*secsInEpoch*time.Second)
	pool := &PrevoteCaches{
		unAggregatedPrevote: make(map[[32]byte]*ethpb.PreVote),
		aggregatedPrevote:   make(map[[32]byte][]*ethpb.PreVote),
		seenPrevote:         c,
	}

	return pool
}

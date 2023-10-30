package stategen

import (
	"context"
	"sync"

	lru "github.com/hashicorp/golang-lru"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/state"
	lruwrpr "gitlab.waterfall.network/waterfall/protocol/coordinator/cache/lru"
)

var (
	// HotStateCacheSize defines the max number of hot state this can cache.
	HotStateCacheSize = 64
	// Metrics
	hotStateCacheHit = promauto.NewCounter(prometheus.CounterOpts{
		Name: "hot_state_cache_hit",
		Help: "The total number of cache hits on the hot state cache.",
	})
	hotStateCacheMiss = promauto.NewCounter(prometheus.CounterOpts{
		Name: "hot_state_cache_miss",
		Help: "The total number of cache misses on the hot state cache.",
	})
)

// hotStateCache is used to store the processed beacon state after finalized check point..
type hotStateCache struct {
	cache           *lru.Cache
	blockStateRoots *lru.Cache
	lock            sync.RWMutex
}

// newHotStateCache initializes the map and underlying cache.
func newHotStateCache() *hotStateCache {
	return &hotStateCache{
		cache:           lruwrpr.New(HotStateCacheSize),
		blockStateRoots: lruwrpr.New(HotStateCacheSize * 2),
	}
}

// Get returns a cached response via input block root, if any.
// The response is copied by default.
func (c *hotStateCache) get(root [32]byte) state.BeaconState {
	c.lock.RLock()
	defer c.lock.RUnlock()
	item, exists := c.cache.Get(root)

	if exists && item != nil {
		hotStateCacheHit.Inc()
		return item.(state.BeaconState).Copy()
	}
	hotStateCacheMiss.Inc()
	return nil
}

func (c *hotStateCache) ByRoot(root [32]byte) (state.BeaconState, error) {
	st := c.get(root)
	if st == nil {
		return nil, ErrNotInCache
	}
	return st, nil
}

// GetWithoutCopy returns a non-copied cached response via input block root.
func (c *hotStateCache) getWithoutCopy(root [32]byte) state.BeaconState {
	c.lock.RLock()
	defer c.lock.RUnlock()
	item, exists := c.cache.Get(root)
	if exists && item != nil {
		hotStateCacheHit.Inc()
		return item.(state.BeaconState)
	}
	hotStateCacheMiss.Inc()
	return nil
}

// put the response in the cache.
func (c *hotStateCache) put(root [32]byte, state state.BeaconState) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if state != nil {
		stRoot, err := state.HashTreeRoot(context.Background())
		if err == nil {
			c.blockStateRoots.Add(root, stRoot)
		}
	}
	c.cache.Add(root, state)
}

func (c *hotStateCache) isNotMutated(root [32]byte) bool {
	c.lock.Lock()
	defer c.lock.Unlock()

	rtVal, exists := c.blockStateRoots.Get(root)
	if !exists || rtVal == nil {
		return false
	}
	checkRoot, ok := rtVal.([32]byte)
	if !ok {
		c.blockStateRoots.Remove(root)
		return false
	}

	item, exists := c.cache.Get(root)
	if !exists || item == nil {
		c.blockStateRoots.Remove(root)
		return false
	}
	st, ok := item.(state.BeaconState)
	if !ok {
		c.blockStateRoots.Remove(root)
		return false
	}

	stRoot, err := st.HashTreeRoot(context.Background())
	if err != nil {
		return false
	}

	if stRoot != checkRoot {
		c.blockStateRoots.Remove(root)
	}

	return stRoot == checkRoot
}

// has returns true if the key exists in the cache.
func (c *hotStateCache) has(root [32]byte) bool {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.cache.Contains(root)
}

// delete deletes the key exists in the cache.
func (c *hotStateCache) delete(root [32]byte) bool {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.blockStateRoots.Remove(root)
	return c.cache.Remove(root)
}

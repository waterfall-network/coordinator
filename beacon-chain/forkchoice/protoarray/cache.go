package protoarray

import (
	"sort"
	"sync"

	lru "github.com/hashicorp/golang-lru"
	types "github.com/prysmaticlabs/eth2-types"
	lruwrpr "gitlab.waterfall.network/waterfall/protocol/coordinator/cache/lru"
	gwatCommon "gitlab.waterfall.network/waterfall/protocol/gwat/common"
)

const (
	// maxForkChoiceCacheSize defines the max number of items can cache.
	maxForkChoiceCacheSize = int(4)
)

// cacheForkChoice main cache instance
var cacheForkChoice *ForkChoiceCache = NewForkChoiceCache()

// ForkChoiceCache is a struct with 1 LRU cache for looking up forkchoice.
type ForkChoiceCache struct {
	cache *lru.Cache
	lock  sync.RWMutex
}

// NewForkChoiceCache creates a new cache of forkchoice.
func NewForkChoiceCache() *ForkChoiceCache {
	return &ForkChoiceCache{
		cache: lruwrpr.New(maxForkChoiceCacheSize),
	}
}

// Add adds a new ForkChoice entry into the cache.
func (c *ForkChoiceCache) Add(fc *ForkChoice) {
	c.lock.Lock()
	defer c.lock.Unlock()

	cpy := fc.Copy()
	key := cacheKeyByRootIndexMap(cpy.store.nodesIndices)
	_ = c.cache.Add(key, cpy)
	return
}

// Get returns the forkchoice in cache.
func (c *ForkChoiceCache) Get(rootIndexMap map[[32]byte]uint64) *ForkChoice {
	c.lock.RLock()
	defer c.lock.RUnlock()

	key := cacheKeyByRootIndexMap(rootIndexMap)
	value, exists := c.cache.Get(key)
	if !exists {
		return nil
	}
	return value.(*ForkChoice).Copy()
}

func cacheKeyByRootIndexMap(rootIndexMap map[[32]byte]uint64) [32]byte {
	if len(rootIndexMap) == 0 {
		return [32]byte{}
	}
	roots := make(gwatCommon.HashArray, len(rootIndexMap))
	i := 0
	for r := range rootIndexMap {
		roots[i] = r
		i++
	}
	return roots.Key()
}

// maxSlotKeyMap calcs the max slots of forkchoices in the cache.
func (c *ForkChoiceCache) maxSlotKeyMap() map[[32]byte]types.Slot {
	slotKeyMap := make(map[[32]byte]types.Slot)
	for _, key := range c.cache.Keys() {
		value, exists := c.cache.Get(key)
		if exists {
			continue
		}
		fc, ok := value.(*ForkChoice)
		if !ok {
			continue
		}
		var maxSlot types.Slot
		for _, n := range fc.store.nodes {
			if n.slot > maxSlot {
				maxSlot = n.slot
			}
		}
		slotKeyMap[key.([32]byte)] = maxSlot
	}
	return slotKeyMap
}

// SearchCompatibleFc searches cached forkchoice compatible with rootIndexMap
// and calculate nodes that are not included in forkchoice.
func (c *ForkChoiceCache) SearchCompatibleFc(rootIndexMap map[[32]byte]uint64) (fc *ForkChoice, diff map[[32]byte]uint64) {
	diff = make(map[[32]byte]uint64)

	// descending sort node's indexes
	nodeIndexes := make(gwatCommon.SorterDescU64, 0, len(rootIndexMap))
	indexRootMap := make(map[uint64]gwatCommon.Hash, len(rootIndexMap))
	for r, index := range rootIndexMap {
		nodeIndexes = append(nodeIndexes, index)
		indexRootMap[index] = r
	}
	sort.Sort(nodeIndexes)
	// collect descending sorted nodes' roots
	roots := make(gwatCommon.HashArray, len(nodeIndexes))
	for i, index := range nodeIndexes {
		roots[i] = indexRootMap[index]
	}
	// search compatible key and
	for i, r := range roots {
		keyRoots := roots[i:]
		var key [32]byte = keyRoots.Key()
		value, exists := c.cache.Get(key)
		if exists {
			fc, ok := value.(*ForkChoice)
			if ok {
				return fc.Copy(), diff
			}
		}
		diff[r] = rootIndexMap[r]
	}
	return nil, diff
}

// getCompatibleFc searches/create forkchoice inctance compatible with rootIndexMap
// and calculate nodes that are not included in forkchoice.
// Helper function for workflow optimization.
func getCompatibleFc(nodesRootIndexMap map[[32]byte]uint64, currFc *ForkChoice) (fc *ForkChoice, diff map[[32]byte]uint64) {
	// if current fc is equivalent target fc
	if cacheKeyByRootIndexMap(currFc.store.nodesIndices) == cacheKeyByRootIndexMap(nodesRootIndexMap) {
		fc = currFc.Copy()
		diff = map[[32]byte]uint64{}
		return fc, diff
	}
	// search cached fc
	fc, diff = cacheForkChoice.SearchCompatibleFc(nodesRootIndexMap)
	if fc != nil {
		return fc, diff
	}
	// create new ForkChoice instance
	fc = New(currFc.store.justifiedEpoch, currFc.store.finalizedEpoch)
	diff = nodesRootIndexMap
	return fc, diff
}

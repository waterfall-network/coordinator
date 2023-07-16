package prevote

import (
	"context"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	types "github.com/prysmaticlabs/eth2-types"
	"github.com/prysmaticlabs/go-bitfield"
	ethpb "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/time/slots"
	"go.opencensus.io/trace"
)

func (c *PrevoteCache) HasPrevote(pv *ethpb.PreVote) (bool, error) {
	if pv == nil || pv.Data == nil {
		return false, errors.New("Prevote data cannot be nil")
	}

	c.prevoteCacheLock.RLock()
	defer c.prevoteCacheLock.RUnlock()
	if p, ok := c.prevoteCache[pv.Data.Slot]; ok {
		for _, v := range p {
			if c, err := v.AggregationBits.Contains(pv.AggregationBits); err != nil {
				return false, err
			} else if c {
				return true, nil
			}
		}
	}

	return false, nil
}

func (c *PrevoteCache) SavePrevote(pv *ethpb.PreVote) error {
	if pv == nil {
		return nil
	}

	seen, err := c.hasSeenBit(pv)
	if err != nil {
		return err
	}
	if seen {
		return nil
	}

	copiedPv := ethpb.CopyPrevote(pv) // Copied.

	c.prevoteCacheLock.Lock()
	defer c.prevoteCacheLock.Unlock()

	if val, exists := c.prevoteCache[pv.Data.Slot]; exists {
		newVal := append(val, copiedPv)
		c.prevoteCache[pv.Data.Slot] = newVal
	} else {
		c.prevoteCache[pv.Data.Slot] = append(c.prevoteCache[pv.Data.Slot], copiedPv)
	}

	return nil
}

func (c *PrevoteCache) hasSeenBit(pv *ethpb.PreVote) (bool, error) {
	r, err := hashFn(pv.Data)
	if err != nil {
		return false, err
	}

	v, ok := c.seenPrevote.Get(string(r[:]))
	if ok {
		seenBits, ok := v.([]bitfield.Bitlist)
		if !ok {
			return false, errors.New("could not convert to bitlist type")
		}
		for _, bit := range seenBits {
			if c, err := bit.Contains(pv.AggregationBits); err != nil {
				return false, err
			} else if c {
				return true, nil
			}
		}
	}
	return false, nil
}

func (c *PrevoteCache) GetPrevoteBySlot(ctx context.Context, slot types.Slot) ([]*ethpb.PreVote, error) {
	_, span := trace.StartSpan(ctx, "operations.prevote.GetPrevoteBySlot")
	defer span.End()

	c.prevoteCacheLock.RLock()
	defer c.prevoteCacheLock.RUnlock()

	pv := make([]*ethpb.PreVote, 0)

	for k, v := range c.prevoteCache {
		if k == slot {
			pv = v
		}
	}

	if len(pv) == 0 {
		return []*ethpb.PreVote{}, errors.Errorf("No prevote data for slot %v", slot)
	}

	return pv, nil
}

func (c *PrevoteCache) PurgeOutdatedPrevote(t time.Time) error {
	c.prevoteCacheLock.RLock()
	defer c.prevoteCacheLock.RUnlock()

	for k, v := range c.prevoteCache {
		if k < slots.CurrentSlot(uint64(t.Unix())) {
			for _, p := range v {
				err := c.insertSeenBit(p)
				if err != nil {
					return err
				}
			}
			delete(c.prevoteCache, k)
		}
	}
	return nil
}

func (c *PrevoteCache) insertSeenBit(pv *ethpb.PreVote) error {
	r, err := hashFn(pv.Data)
	if err != nil {
		return err
	}

	v, ok := c.seenPrevote.Get(string(r[:]))
	if ok {
		seenBits, ok := v.([]bitfield.Bitlist)
		if !ok {
			return errors.New("could not convert to bitlist type")
		}
		alreadyExists := false
		for _, bit := range seenBits {
			if c, err := bit.Contains(pv.AggregationBits); err != nil {
				return err
			} else if c {
				alreadyExists = true
				break
			}
		}
		if !alreadyExists {
			seenBits = append(seenBits, pv.AggregationBits)
		}
		c.seenPrevote.Set(string(r[:]), seenBits, cache.DefaultExpiration /* one epoch */)
		return nil
	}

	c.seenPrevote.Set(string(r[:]), []bitfield.Bitlist{pv.AggregationBits}, cache.DefaultExpiration /* one epoch */)
	return nil
}

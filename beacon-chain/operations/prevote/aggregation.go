package prevote

import (
	"github.com/pkg/errors"
	"github.com/prysmaticlabs/go-bitfield"
	ethpb "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1"
)

func (c *PrevoteCache) HasPrevote(pv *ethpb.PreVote) (bool, error) {
	if pv == nil || pv.Data == nil {
		return false, errors.New("Prevote data cannot be nil")
	}
	r, err := hashFn(pv.Data)
	if err != nil {
		return false, errors.Wrap(err, "could not tree hash prevote")
	}

	c.prevoteCacheLock.RLock()
	defer c.prevoteCacheLock.RUnlock()
	if pv, ok := c.prevoteCache[r]; ok {
		if c, err := pv.AggregationBits.Contains(pv.AggregationBits); err != nil {
			return false, err
		} else if c {
			return true, nil
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

	r, err := hashFn(pv)
	if err != nil {
		return errors.Wrap(err, "could not tree hash prevote")
	}
	pv = ethpb.CopyPrevote(pv) // Copied.
	c.prevoteCacheLock.Lock()
	defer c.prevoteCacheLock.Unlock()
	c.prevoteCache[r] = pv

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

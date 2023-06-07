package db

import (
	"context"

	types "github.com/prysmaticlabs/eth2-types"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/config/params"
)

func BlockInfoFetcherFunc(db ReadOnlyDatabase) params.CtxBlockFetcher {
	return func(ctx context.Context, blockRoot [32]byte) (types.ValidatorIndex, types.Slot, uint64, error) {
		block, err := db.Block(ctx, blockRoot)
		if err != nil {
			return 0, 0, 0, err
		}

		votesIncluded := uint64(0)
		for _, att := range block.Block().Body().Attestations() {
			votesIncluded += att.AggregationBits.Count()
		}

		return block.Block().ProposerIndex(), block.Block().Slot(), votesIncluded, nil
	}
}

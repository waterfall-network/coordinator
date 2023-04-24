package db

import (
	"context"

	types "github.com/prysmaticlabs/eth2-types"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/config/params"
)

func BlockInfoFetcherFunc(db ReadOnlyDatabase) params.BlockInfoFetcherContextValue {
	return func(ctx context.Context, blockRoot [32]byte) (types.ValidatorIndex, types.Slot, error) {
		block, err := db.Block(ctx, blockRoot)
		if err != nil {
			return 0, 0, err
		}

		return block.Block().ProposerIndex(), block.Block().Slot(), nil
	}
}

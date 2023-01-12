package db

import "gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/db/kv"

var _ Database = (*kv.Store)(nil)

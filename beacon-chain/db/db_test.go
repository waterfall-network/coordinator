package db

import "github.com/waterfall-foundation/coordinator/beacon-chain/db/kv"

var _ Database = (*kv.Store)(nil)

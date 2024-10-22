package kv

import (
	"bytes"
	"context"
	"strconv"

	"gitlab.waterfall.network/waterfall/protocol/coordinator/encoding/bytesutil"
	bolt "go.etcd.io/bbolt"
)

var migrationBlockSlotIndex0Key = []byte("block_slot_index_0")

func migrateBlockSlotIndex(ctx context.Context, db *bolt.DB) error {
	if updateErr := db.Batch(func(tx *bolt.Tx) error {
		mb := tx.Bucket(migrationsBucket)
		if b := mb.Get(migrationBlockSlotIndex0Key); bytes.Equal(b, migrationCompleted) {
			return nil // Migration already completed.
		}

		bkt := tx.Bucket(blockSlotIndicesBucket)

		// Convert indices from strings to big endian integers.
		if err := bkt.ForEach(func(k, v []byte) error {
			key, err := strconv.ParseUint(string(k), 10, 64)
			if err != nil {
				return err
			}
			if err = bkt.Delete(k); err != nil {
				return err
			}
			if err = bkt.Put(bytesutil.Uint64ToBytesBigEndian(key), v); err != nil {
				return err
			}
			// check if context is canceled in between
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return nil
		}); err != nil {
			return err
		}

		return mb.Put(migrationBlockSlotIndex0Key, migrationCompleted)
	}); updateErr != nil {
		log.WithError(updateErr).Errorf("could not migrate bucket: %s", blockSlotIndicesBucket)
		return updateErr
	}
	return nil
}

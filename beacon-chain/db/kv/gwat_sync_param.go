package kv

import (
	"context"

	types "github.com/prysmaticlabs/eth2-types"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/encoding/bytesutil"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1/wrapper"
	bolt "go.etcd.io/bbolt"
	"go.opencensus.io/trace"
)

// GwatSyncParam retrieval by epoch.
func (s *Store) GwatSyncParam(ctx context.Context, epoch types.Epoch) (*wrapper.GwatSyncParam, error) {
	ctx, span := trace.StartSpan(ctx, "BeaconDB.GwatSyncParam")
	defer span.End()

	var err error
	var data *wrapper.GwatSyncParam
	err = s.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(gwatSyncParamBucket)
		key := bytesutil.EpochToBytesBigEndian(epoch)
		enc := bkt.Get(key)
		if enc == nil {
			return nil
		}
		data, err = wrapper.BytesToGwatSyncParam(enc)
		return err
	})
	return data, err
}

// SaveGwatSyncParam to the db.
func (s *Store) SaveGwatSyncParam(ctx context.Context, gsp wrapper.GwatSyncParam) error {
	ctx, span := trace.StartSpan(ctx, "BeaconDB.SaveGwatSyncParams")
	defer span.End()

	return s.db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(gwatSyncParamBucket)
		key := bytesutil.EpochToBytesBigEndian(gsp.Epoch())
		enc, err := gsp.Bytes()
		if err != nil {
			return err
		}
		if err = bkt.Put(key, enc); err != nil {
			return err
		}
		return nil
	})
}

// DeleteGwatSyncParam from the db
func (s *Store) DeleteGwatSyncParam(ctx context.Context, epoch types.Epoch) error {
	ctx, span := trace.StartSpan(ctx, "BeaconDB.DeleteGwatSyncParam")
	defer span.End()

	return s.db.Update(func(tx *bolt.Tx) error {
		key := bytesutil.EpochToBytesBigEndian(epoch)
		if err := tx.Bucket(gwatSyncParamBucket).Delete(key); err != nil {
			return err
		}
		return nil
	})
}

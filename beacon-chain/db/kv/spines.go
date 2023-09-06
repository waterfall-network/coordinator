package kv

import (
	"context"

	"gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1/wrapper"
	bolt "go.etcd.io/bbolt"
	"go.opencensus.io/trace"
)

// ReadSpines retrieval spines represented by plain bytes array.
func (s *Store) ReadSpines(ctx context.Context, key [32]byte) (wrapper.Spines, error) {
	ctx, span := trace.StartSpan(ctx, "BeaconDB.ReadSpines")
	defer span.End()

	var err error
	var data wrapper.Spines

	if key == (wrapper.Spines{}).Key() {
		return wrapper.Spines{}, nil
	}

	// Return from cache if it exists.
	if v, ok := s.spinesCache.Get(key); v != nil && ok {
		return v.(wrapper.Spines), nil
	}

	err = s.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(spinesBucket)
		data = bkt.Get(key[:])
		return err
	})
	// cache it.
	s.spinesCache.Add(key, data)
	return data, err
}

// WriteSpines to the db.
func (s *Store) WriteSpines(ctx context.Context, spines wrapper.Spines) ([32]byte, error) {
	ctx, span := trace.StartSpan(ctx, "BeaconDB.WriteSpines")
	defer span.End()
	key := spines.Key()
	return key, s.db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(spinesBucket)
		if err := bkt.Put(key[:], spines); err != nil {
			return err
		}
		// cache it.
		s.spinesCache.Add(key, spines)
		return nil
	})
}

// DeleteSpines from the db
func (s *Store) DeleteSpines(ctx context.Context, key [32]byte) error {
	ctx, span := trace.StartSpan(ctx, "BeaconDB.DeleteSpines")
	defer span.End()

	s.spinesCache.Remove(key)
	return s.db.Update(func(tx *bolt.Tx) error {
		if err := tx.Bucket(spinesBucket).Delete(key[:]); err != nil {
			return err
		}
		return nil
	})
}

package testing

import (
	"context"
	"testing"

	fieldparams "github.com/waterfall-foundation/coordinator/config/fieldparams"
	"github.com/waterfall-foundation/coordinator/validator/db/iface"
	"github.com/waterfall-foundation/coordinator/validator/db/kv"
)

// SetupDB instantiates and returns a DB instance for the validator client.
func SetupDB(t testing.TB, pubkeys [][fieldparams.BLSPubkeyLength]byte) iface.ValidatorDB {
	db, err := kv.NewKVStore(context.Background(), t.TempDir(), &kv.Config{
		PubKeys: pubkeys,
	})
	if err != nil {
		t.Fatalf("Failed to instantiate DB: %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Fatalf("Failed to close database: %v", err)
		}
		if err := db.ClearDB(); err != nil {
			t.Fatalf("Failed to clear database: %v", err)
		}
	})
	return db
}

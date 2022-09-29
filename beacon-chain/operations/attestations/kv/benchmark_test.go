package kv_test

import (
	"testing"

	"github.com/waterfall-foundation/coordinator/beacon-chain/operations/attestations/kv"
	ethpb "github.com/waterfall-foundation/coordinator/proto/prysm/v1alpha1"
	"github.com/waterfall-foundation/coordinator/testing/assert"
)

func BenchmarkAttCaches(b *testing.B) {
	ac := kv.NewAttCaches()

	att := &ethpb.Attestation{}

	for i := 0; i < b.N; i++ {
		assert.NoError(b, ac.SaveUnaggregatedAttestation(att))
		assert.NoError(b, ac.DeleteAggregatedAttestation(att))
	}
}

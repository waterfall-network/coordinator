package kv_test

import (
	"testing"

	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/operations/attestations/kv"
	ethpb "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/assert"
)

func BenchmarkAttCaches(b *testing.B) {
	ac := kv.NewAttCaches()

	att := &ethpb.Attestation{}

	for i := 0; i < b.N; i++ {
		assert.NoError(b, ac.SaveUnaggregatedAttestation(att))
		assert.NoError(b, ac.DeleteAggregatedAttestation(att))
	}
}

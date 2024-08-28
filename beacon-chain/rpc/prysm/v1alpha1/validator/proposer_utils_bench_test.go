package validator

import (
	"fmt"
	"testing"

	"github.com/prysmaticlabs/go-bitfield"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/config/features"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/config/params"
	ethpb "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1"
	aggtesting "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1/attestation/aggregation/testing"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/require"
)

func BenchmarkProposerAtts_sortByProfitability(b *testing.B) {
	bitlistLen := params.BeaconConfig().MaxValidatorsPerCommittee

	tests := []struct {
		name   string
		inputs []bitfield.Bitlist
	}{
		{
			name:   "256 attestations with single bit set",
			inputs: aggtesting.BitlistsWithSingleBitSet(256, bitlistLen),
		},
		{
			name:   "256 attestations with 64 random bits set",
			inputs: aggtesting.BitlistsWithSingleBitSet(256, bitlistLen),
		},
		{
			name:   "512 attestations with single bit set",
			inputs: aggtesting.BitlistsWithSingleBitSet(512, bitlistLen),
		},
		{
			name:   "1024 attestations with 64 random bits set",
			inputs: aggtesting.BitlistsWithMultipleBitSet(b, 1024, bitlistLen, 64),
		},
		{
			name:   "1024 attestations with 512 random bits set",
			inputs: aggtesting.BitlistsWithMultipleBitSet(b, 1024, bitlistLen, 512),
		},
		{
			name:   "1024 attestations with 1000 random bits set",
			inputs: aggtesting.BitlistsWithMultipleBitSet(b, 1024, bitlistLen, 1000),
		},
	}

	runner := func(atts []*ethpb.Attestation) {
		attsCopy := make(proposerAtts, len(atts))
		for i, att := range atts {
			attsCopy[i] = ethpb.CopyAttestation(att)
		}
		_, err := attsCopy.sortByProfitability()
		require.NoError(b, err, "Could not sort attestations by profitability")
	}

	for _, tt := range tests {
		b.Run(fmt.Sprintf("max-cover_%s", tt.name), func(b *testing.B) {
			b.StopTimer()
			resetCfg := features.InitWithReset(&features.Flags{})
			defer resetCfg()
			atts := aggtesting.MakeAttestationsFromBitlists(tt.inputs)
			b.StartTimer()
			for i := 0; i < b.N; i++ {
				runner(atts)
			}
		})
	}
}

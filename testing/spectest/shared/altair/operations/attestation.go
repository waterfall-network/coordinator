package operations

import (
	"context"
	"errors"
	"path"
	"testing"

	"github.com/golang/snappy"
	types "github.com/prysmaticlabs/eth2-types"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/core/altair"
	b "gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/core/blocks"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/state"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/config/params"
	ethpb "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1/block"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/require"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/spectest/utils"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/util"
)

func RunAttestationTest(t *testing.T, config string) {
	require.NoError(t, utils.SetConfig(t, config))
	testFolders, testsFolderPath := utils.TestFolders(t, config, "altair", "operations/attestation/pyspec_tests")
	for _, folder := range testFolders {
		t.Run(folder.Name(), func(t *testing.T) {
			folderPath := path.Join(testsFolderPath, folder.Name())
			attestationFile, err := util.BazelFileBytes(folderPath, "attestation.ssz_snappy")
			require.NoError(t, err)
			attestationSSZ, err := snappy.Decode(nil /* dst */, attestationFile)
			require.NoError(t, err, "Failed to decompress")
			att := &ethpb.Attestation{}
			require.NoError(t, att.UnmarshalSSZ(attestationSSZ), "Failed to unmarshal")

			body := &ethpb.BeaconBlockBodyAltair{Attestations: []*ethpb.Attestation{att}}
			processAtt := func(ctx context.Context, st state.BeaconState, blk block.SignedBeaconBlock) (state.BeaconState, error) {
				ctxBlockFetcher := params.CtxBlockFetcher(func(ctx context.Context, blockRoot [32]byte) (types.ValidatorIndex, types.Slot, uint64, error) {
					block := blk
					votesIncluded := uint64(0)
					for _, att := range block.Block().Body().Attestations() {
						votesIncluded += att.AggregationBits.Count()
					}

					return block.Block().ProposerIndex(), block.Block().Slot(), votesIncluded, nil
				})

				ctxWithFetcher := context.WithValue(context.Background(),
					params.BeaconConfig().CtxBlockFetcherKey,
					ctxBlockFetcher)
				st, err = altair.ProcessAttestationsNoVerifySignature(ctxWithFetcher, st, blk)
				if err != nil {
					return nil, err
				}
				aSet, err := b.AttestationSignatureBatch(ctx, st, blk.Block().Body().Attestations())
				if err != nil {
					return nil, err
				}
				verified, err := aSet.Verify()
				if err != nil {
					return nil, err
				}
				if !verified {
					return nil, errors.New("could not batch verify attestation signature")
				}
				return st, nil
			}

			RunBlockOperationTest(t, folderPath, body, processAtt)
		})
	}
}

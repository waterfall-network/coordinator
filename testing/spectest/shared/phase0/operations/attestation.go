package operations

import (
	"context"
	"path"
	"testing"

	"github.com/golang/snappy"
	"github.com/pkg/errors"
	b "github.com/waterfall-foundation/coordinator/beacon-chain/core/blocks"
	"github.com/waterfall-foundation/coordinator/beacon-chain/state"
	ethpb "github.com/waterfall-foundation/coordinator/proto/prysm/v1alpha1"
	"github.com/waterfall-foundation/coordinator/proto/prysm/v1alpha1/block"
	"github.com/waterfall-foundation/coordinator/testing/require"
	"github.com/waterfall-foundation/coordinator/testing/spectest/utils"
	"github.com/waterfall-foundation/coordinator/testing/util"
)

// RunAttestationTest executes "operations/attestation" tests.
func RunAttestationTest(t *testing.T, config string) {
	require.NoError(t, utils.SetConfig(t, config))
	testFolders, testsFolderPath := utils.TestFolders(t, config, "phase0", "operations/attestation/pyspec_tests")
	for _, folder := range testFolders {
		t.Run(folder.Name(), func(t *testing.T) {
			folderPath := path.Join(testsFolderPath, folder.Name())
			attestationFile, err := util.BazelFileBytes(folderPath, "attestation.ssz_snappy")
			require.NoError(t, err)
			attestationSSZ, err := snappy.Decode(nil /* dst */, attestationFile)
			require.NoError(t, err, "Failed to decompress")
			att := &ethpb.Attestation{}
			require.NoError(t, att.UnmarshalSSZ(attestationSSZ), "Failed to unmarshal")

			body := &ethpb.BeaconBlockBody{Attestations: []*ethpb.Attestation{att}}
			processAtt := func(ctx context.Context, st state.BeaconState, blk block.SignedBeaconBlock) (state.BeaconState, error) {
				st, err = b.ProcessAttestationsNoVerifySignature(ctx, st, blk)
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

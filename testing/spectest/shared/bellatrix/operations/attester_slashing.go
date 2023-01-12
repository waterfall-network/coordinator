package operations

import (
	"context"
	"path"
	"testing"

	"github.com/golang/snappy"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/core/blocks"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/core/validators"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/state"
	ethpb "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1/block"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/require"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/spectest/utils"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/util"
)

func RunAttesterSlashingTest(t *testing.T, config string) {
	require.NoError(t, utils.SetConfig(t, config))
	testFolders, testsFolderPath := utils.TestFolders(t, config, "bellatrix", "operations/attester_slashing/pyspec_tests")
	for _, folder := range testFolders {
		t.Run(folder.Name(), func(t *testing.T) {
			folderPath := path.Join(testsFolderPath, folder.Name())
			attSlashingFile, err := util.BazelFileBytes(folderPath, "attester_slashing.ssz_snappy")
			require.NoError(t, err)
			attSlashingSSZ, err := snappy.Decode(nil /* dst */, attSlashingFile)
			require.NoError(t, err, "Failed to decompress")
			attSlashing := &ethpb.AttesterSlashing{}
			require.NoError(t, attSlashing.UnmarshalSSZ(attSlashingSSZ), "Failed to unmarshal")

			body := &ethpb.BeaconBlockBodyBellatrix{AttesterSlashings: []*ethpb.AttesterSlashing{attSlashing}}
			RunBlockOperationTest(t, folderPath, body, func(ctx context.Context, s state.BeaconState, b block.SignedBeaconBlock) (state.BeaconState, error) {
				return blocks.ProcessAttesterSlashings(ctx, s, b.Block().Body().AttesterSlashings(), validators.SlashValidator)
			})
		})
	}
}

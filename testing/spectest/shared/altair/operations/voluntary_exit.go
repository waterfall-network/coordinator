package operations

import (
	"context"
	"path"
	"testing"

	"github.com/golang/snappy"
	"github.com/waterfall-foundation/coordinator/beacon-chain/core/blocks"
	"github.com/waterfall-foundation/coordinator/beacon-chain/state"
	ethpb "github.com/waterfall-foundation/coordinator/proto/prysm/v1alpha1"
	"github.com/waterfall-foundation/coordinator/proto/prysm/v1alpha1/block"
	"github.com/waterfall-foundation/coordinator/testing/require"
	"github.com/waterfall-foundation/coordinator/testing/spectest/utils"
	"github.com/waterfall-foundation/coordinator/testing/util"
)

func RunVoluntaryExitTest(t *testing.T, config string) {
	require.NoError(t, utils.SetConfig(t, config))
	testFolders, testsFolderPath := utils.TestFolders(t, config, "altair", "operations/voluntary_exit/pyspec_tests")
	for _, folder := range testFolders {
		t.Run(folder.Name(), func(t *testing.T) {
			folderPath := path.Join(testsFolderPath, folder.Name())
			exitFile, err := util.BazelFileBytes(folderPath, "voluntary_exit.ssz_snappy")
			require.NoError(t, err)
			exitSSZ, err := snappy.Decode(nil /* dst */, exitFile)
			require.NoError(t, err, "Failed to decompress")
			voluntaryExit := &ethpb.SignedVoluntaryExit{}
			require.NoError(t, voluntaryExit.UnmarshalSSZ(exitSSZ), "Failed to unmarshal")

			body := &ethpb.BeaconBlockBodyAltair{VoluntaryExits: []*ethpb.SignedVoluntaryExit{voluntaryExit}}
			RunBlockOperationTest(t, folderPath, body, func(ctx context.Context, s state.BeaconState, b block.SignedBeaconBlock) (state.BeaconState, error) {
				return blocks.ProcessVoluntaryExits(ctx, s, b.Block().Body().VoluntaryExits())
			})
		})
	}
}

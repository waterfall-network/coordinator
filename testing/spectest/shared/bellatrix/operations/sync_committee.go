package operations

import (
	"context"
	"path"
	"testing"

	"github.com/golang/snappy"
	"github.com/waterfall-foundation/coordinator/beacon-chain/core/altair"
	"github.com/waterfall-foundation/coordinator/beacon-chain/state"
	ethpb "github.com/waterfall-foundation/coordinator/proto/prysm/v1alpha1"
	"github.com/waterfall-foundation/coordinator/proto/prysm/v1alpha1/block"
	"github.com/waterfall-foundation/coordinator/testing/require"
	"github.com/waterfall-foundation/coordinator/testing/spectest/utils"
	"github.com/waterfall-foundation/coordinator/testing/util"
)

func RunSyncCommitteeTest(t *testing.T, config string) {
	require.NoError(t, utils.SetConfig(t, config))
	testFolders, testsFolderPath := utils.TestFolders(t, config, "bellatrix", "operations/sync_aggregate/pyspec_tests")
	for _, folder := range testFolders {
		t.Run(folder.Name(), func(t *testing.T) {
			folderPath := path.Join(testsFolderPath, folder.Name())
			syncCommitteeFile, err := util.BazelFileBytes(folderPath, "sync_aggregate.ssz_snappy")
			require.NoError(t, err)
			syncCommitteeSSZ, err := snappy.Decode(nil /* dst */, syncCommitteeFile)
			require.NoError(t, err, "Failed to decompress")
			sc := &ethpb.SyncAggregate{}
			require.NoError(t, sc.UnmarshalSSZ(syncCommitteeSSZ), "Failed to unmarshal")

			body := &ethpb.BeaconBlockBodyBellatrix{SyncAggregate: sc}
			RunBlockOperationTest(t, folderPath, body, func(ctx context.Context, s state.BeaconState, b block.SignedBeaconBlock) (state.BeaconState, error) {
				return altair.ProcessSyncAggregate(context.Background(), s, body.SyncAggregate)
			})
		})
	}
}

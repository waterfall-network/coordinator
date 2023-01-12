package operations

import (
	"context"
	"path"
	"testing"

	"github.com/golang/snappy"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/core/altair"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/state"
	ethpb "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1/block"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/require"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/spectest/utils"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/util"
)

func RunDepositTest(t *testing.T, config string) {
	require.NoError(t, utils.SetConfig(t, config))
	testFolders, testsFolderPath := utils.TestFolders(t, config, "bellatrix", "operations/deposit/pyspec_tests")
	for _, folder := range testFolders {
		t.Run(folder.Name(), func(t *testing.T) {
			folderPath := path.Join(testsFolderPath, folder.Name())
			depositFile, err := util.BazelFileBytes(folderPath, "deposit.ssz_snappy")
			require.NoError(t, err)
			depositSSZ, err := snappy.Decode(nil /* dst */, depositFile)
			require.NoError(t, err, "Failed to decompress")
			deposit := &ethpb.Deposit{}
			require.NoError(t, deposit.UnmarshalSSZ(depositSSZ), "Failed to unmarshal")

			body := &ethpb.BeaconBlockBodyBellatrix{Deposits: []*ethpb.Deposit{deposit}}
			processDepositsFunc := func(ctx context.Context, s state.BeaconState, b block.SignedBeaconBlock) (state.BeaconState, error) {
				return altair.ProcessDeposits(ctx, s, b.Block().Body().Deposits())
			}
			RunBlockOperationTest(t, folderPath, body, processDepositsFunc)
		})
	}
}

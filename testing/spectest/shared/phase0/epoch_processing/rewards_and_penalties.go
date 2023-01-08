package epoch_processing

import (
	"context"
	"path"
	"testing"

	"github.com/waterfall-foundation/coordinator/beacon-chain/core/epoch/precompute"
	"github.com/waterfall-foundation/coordinator/beacon-chain/core/helpers"
	"github.com/waterfall-foundation/coordinator/beacon-chain/state"
	"github.com/waterfall-foundation/coordinator/testing/require"
	"github.com/waterfall-foundation/coordinator/testing/spectest/utils"
	"gitlab.waterfall.network/waterfall/protocol/gwat/log"
)

// RunRewardsAndPenaltiesTests executes "epoch_processing/rewards_and_penalties" tests.
func RunRewardsAndPenaltiesTests(t *testing.T, config string) {
	require.NoError(t, utils.SetConfig(t, config))

	testPath := "epoch_processing/rewards_and_penalties/pyspec_tests"
	testFolders, testsFolderPath := utils.TestFolders(t, config, "phase0", testPath)
	for _, folder := range testFolders {
		helpers.ClearCache()
		t.Run(folder.Name(), func(t *testing.T) {
			folderPath := path.Join(testsFolderPath, folder.Name())
			RunEpochOperationTest(t, folderPath, processRewardsAndPenaltiesPrecomputeWrapper)
		})
	}
}

func processRewardsAndPenaltiesPrecomputeWrapper(t *testing.T, st state.BeaconState) (state.BeaconState, error) {
	ctx := context.Background()
	vp, bp, err := precompute.New(ctx, st)
	require.NoError(t, err)
	vp, bp, err = precompute.ProcessAttestations(ctx, st, vp, bp)
	require.NoError(t, err)

	log.Info("process rewards and penalties (wrapper)", "st.slot", st.Slot())

	st, err = precompute.ProcessRewardsAndPenaltiesPrecompute(st, bp, vp, precompute.AttestationsDelta, precompute.ProposersDelta)
	require.NoError(t, err, "Could not process reward")

	return st, nil
}

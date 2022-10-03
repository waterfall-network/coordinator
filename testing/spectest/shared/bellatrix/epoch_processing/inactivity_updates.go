package epoch_processing

import (
	"context"
	"path"
	"testing"

	"github.com/waterfall-foundation/coordinator/beacon-chain/core/altair"
	"github.com/waterfall-foundation/coordinator/beacon-chain/core/helpers"
	"github.com/waterfall-foundation/coordinator/beacon-chain/state"
	"github.com/waterfall-foundation/coordinator/testing/require"
	"github.com/waterfall-foundation/coordinator/testing/spectest/utils"
)

// RunInactivityUpdatesTest executes "epoch_processing/inactivity_updates" tests.
func RunInactivityUpdatesTest(t *testing.T, config string) {
	require.NoError(t, utils.SetConfig(t, config))

	testPath := "epoch_processing/inactivity_updates/pyspec_tests"
	testFolders, testsFolderPath := utils.TestFolders(t, config, "bellatrix", testPath)
	for _, folder := range testFolders {
		helpers.ClearCache()
		t.Run(folder.Name(), func(t *testing.T) {
			folderPath := path.Join(testsFolderPath, folder.Name())
			RunEpochOperationTest(t, folderPath, processInactivityUpdates)
		})
	}
}

func processInactivityUpdates(t *testing.T, st state.BeaconState) (state.BeaconState, error) {
	ctx := context.Background()
	vp, bp, err := altair.InitializePrecomputeValidators(ctx, st)
	require.NoError(t, err)
	vp, _, err = altair.ProcessEpochParticipation(ctx, st, bp, vp)
	require.NoError(t, err)

	st, _, err = altair.ProcessInactivityScores(ctx, st, vp)
	require.NoError(t, err, "Could not process reward")

	return st, nil
}

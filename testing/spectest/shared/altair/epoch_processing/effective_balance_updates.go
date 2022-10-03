package epoch_processing

import (
	"path"
	"testing"

	"github.com/waterfall-foundation/coordinator/beacon-chain/core/epoch"
	"github.com/waterfall-foundation/coordinator/beacon-chain/state"
	"github.com/waterfall-foundation/coordinator/testing/require"
	"github.com/waterfall-foundation/coordinator/testing/spectest/utils"
)

// RunEffectiveBalanceUpdatesTests executes "epoch_processing/effective_balance_updates" tests.
func RunEffectiveBalanceUpdatesTests(t *testing.T, config string) {
	require.NoError(t, utils.SetConfig(t, config))

	testFolders, testsFolderPath := utils.TestFolders(t, config, "altair", "epoch_processing/effective_balance_updates/pyspec_tests")
	for _, folder := range testFolders {
		t.Run(folder.Name(), func(t *testing.T) {
			folderPath := path.Join(testsFolderPath, folder.Name())
			RunEpochOperationTest(t, folderPath, processEffectiveBalanceUpdatesWrapper)
		})
	}
}

func processEffectiveBalanceUpdatesWrapper(t *testing.T, state state.BeaconState) (state.BeaconState, error) {
	state, err := epoch.ProcessEffectiveBalanceUpdates(state)
	require.NoError(t, err, "Could not process final updates")
	return state, nil
}

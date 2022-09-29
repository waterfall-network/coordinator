package epoch_processing

import (
	"path"
	"testing"

	"github.com/waterfall-foundation/coordinator/beacon-chain/core/epoch"
	"github.com/waterfall-foundation/coordinator/beacon-chain/state"
	"github.com/waterfall-foundation/coordinator/testing/require"
	"github.com/waterfall-foundation/coordinator/testing/spectest/utils"
)

// RunParticipationRecordUpdatesTests executes "epoch_processing/participation_record_updates" tests.
func RunParticipationRecordUpdatesTests(t *testing.T, config string) {
	require.NoError(t, utils.SetConfig(t, config))

	testFolders, testsFolderPath := utils.TestFolders(t, config, "phase0", "epoch_processing/participation_record_updates/pyspec_tests")
	for _, folder := range testFolders {
		t.Run(folder.Name(), func(t *testing.T) {
			folderPath := path.Join(testsFolderPath, folder.Name())
			RunEpochOperationTest(t, folderPath, processParticipationRecordUpdatesWrapper)
		})
	}
}

func processParticipationRecordUpdatesWrapper(t *testing.T, state state.BeaconState) (state.BeaconState, error) {
	state, err := epoch.ProcessParticipationRecordUpdates(state)
	require.NoError(t, err, "Could not process final updates")
	return state, nil
}

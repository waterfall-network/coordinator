package epoch_processing

import (
	"context"
	"path"
	"testing"

	"github.com/waterfall-foundation/coordinator/beacon-chain/core/epoch/precompute"
	"github.com/waterfall-foundation/coordinator/beacon-chain/state"
	"github.com/waterfall-foundation/coordinator/testing/require"
	"github.com/waterfall-foundation/coordinator/testing/spectest/utils"
)

// RunJustificationAndFinalizationTests executes "epoch_processing/justification_and_finalization" tests.
func RunJustificationAndFinalizationTests(t *testing.T, config string) {
	require.NoError(t, utils.SetConfig(t, config))

	testPath := "epoch_processing/justification_and_finalization/pyspec_tests"
	testFolders, testsFolderPath := utils.TestFolders(t, config, "phase0", testPath)
	for _, folder := range testFolders {
		t.Run(folder.Name(), func(t *testing.T) {
			folderPath := path.Join(testsFolderPath, folder.Name())
			RunEpochOperationTest(t, folderPath, processJustificationAndFinalizationPrecomputeWrapper)
		})
	}
}

func processJustificationAndFinalizationPrecomputeWrapper(t *testing.T, st state.BeaconState) (state.BeaconState, error) {
	ctx := context.Background()
	vp, bp, err := precompute.New(ctx, st)
	require.NoError(t, err)
	_, bp, err = precompute.ProcessAttestations(ctx, st, vp, bp)
	require.NoError(t, err)

	st, err = precompute.ProcessJustificationAndFinalizationPreCompute(st, bp)
	require.NoError(t, err, "Could not process justification")

	return st, nil
}

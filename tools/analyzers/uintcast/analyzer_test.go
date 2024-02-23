package uintcast_test

import (
	"testing"

	"gitlab.waterfall.network/waterfall/protocol/coordinator/build/bazel"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/tools/analyzers/uintcast"
	"golang.org/x/tools/go/analysis/analysistest"
)

func init() {
	if bazel.BuiltWithBazel() {
		bazel.SetGoEnv()
	}
}

func TestAnalyzer(t *testing.T) {
	t.Skip() //go tool not available
	testdata := bazel.TestDataPath(t)
	analysistest.TestData = func() string { return testdata }
	analysistest.Run(t, testdata, uintcast.Analyzer)
}

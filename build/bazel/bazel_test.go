package bazel_test

import (
	"testing"

	"gitlab.waterfall.network/waterfall/protocol/coordinator/build/bazel"
)

func TestBuildWithBazel(t *testing.T) {
	if !bazel.BuiltWithBazel() {
		t.Skip()
		t.Error("not built with Bazel")
	}
}

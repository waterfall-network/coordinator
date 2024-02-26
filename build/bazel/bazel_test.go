package bazel_test

import (
	"testing"

	"gitlab.waterfall.network/waterfall/protocol/coordinator/build/bazel"
)

func TestBuildWithBazel(t *testing.T) {
	t.Skip()
	if !bazel.BuiltWithBazel() {
		t.Error("not built with Bazel")
	}
}

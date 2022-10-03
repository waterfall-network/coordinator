package bazel_test

import (
	"testing"

	"github.com/waterfall-foundation/coordinator/build/bazel"
)

func TestBuildWithBazel(t *testing.T) {
	if !bazel.BuiltWithBazel() {
		t.Error("not built with Bazel")
	}
}

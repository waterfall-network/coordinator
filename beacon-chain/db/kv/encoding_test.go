package kv

import (
	"context"
	"testing"

	testpb "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/testing"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/require"
)

func Test_encode_handlesNilFromFunction(t *testing.T) {
	foo := func() *testpb.Puzzle {
		return nil
	}
	_, err := encode(context.Background(), foo())
	require.ErrorContains(t, "cannot encode nil message", err)
}

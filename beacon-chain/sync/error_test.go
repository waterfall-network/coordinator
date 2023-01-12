package sync

import (
	"bytes"
	"testing"

	p2ptest "gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/p2p/testing"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/p2p/types"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/assert"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/require"
)

func TestRegularSync_generateErrorResponse(t *testing.T) {
	r := &Service{
		cfg: &config{p2p: p2ptest.NewTestP2P(t)},
	}
	data, err := r.generateErrorResponse(responseCodeServerError, "something bad happened")
	require.NoError(t, err)

	buf := bytes.NewBuffer(data)
	b := make([]byte, 1)
	_, err = buf.Read(b)
	require.NoError(t, err)
	assert.Equal(t, responseCodeServerError, b[0], "The first byte was not the status code")
	msg := &types.ErrorMessage{}
	require.NoError(t, r.cfg.p2p.Encoding().DecodeWithMaxLength(buf, msg))
	assert.Equal(t, "something bad happened", string(*msg), "Received the wrong message")
}

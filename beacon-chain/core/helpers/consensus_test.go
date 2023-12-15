package helpers_test

import (
	"fmt"
	"testing"

	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/core/helpers"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/assert"
	gwatCommon "gitlab.waterfall.network/waterfall/protocol/gwat/common"
)

func TestValidateUnpublishedChains(t *testing.T) {
	var err error
	chain_1 := gwatCommon.HashArray{
		gwatCommon.BytesToHash([]byte{'a', '1'}),
	}
	chain_2 := gwatCommon.HashArray{
		gwatCommon.BytesToHash([]byte{'a', '1'}), gwatCommon.BytesToHash([]byte{'b', '2'}),
	}
	chain_3 := gwatCommon.HashArray{
		gwatCommon.BytesToHash([]byte{'a', '1'}), gwatCommon.BytesToHash([]byte{'a', '2'}),
	}

	// No errors
	correctParam := []gwatCommon.HashArray{
		chain_3, chain_2, chain_1,
	}
	err = helpers.ConsensusValidateUnpublishedChains(correctParam)
	assert.NoError(t, err)

	// err "contains empty chain"
	unpubChains := []gwatCommon.HashArray{
		chain_3, {}, chain_1,
	}
	err = helpers.ConsensusValidateUnpublishedChains(unpubChains)
	assert.ErrorContains(t, helpers.ErrBadUnpublishedChains.Error(), err)
	assert.ErrorContains(t, "contains empty chain", err)

	// err "the first values of each chain must be equal"
	// bad chain
	chain_4 := gwatCommon.HashArray{
		gwatCommon.BytesToHash([]byte{'c', '1'}), gwatCommon.BytesToHash([]byte{'a', '2'}),
	}
	unpubChains = []gwatCommon.HashArray{
		chain_4, chain_3, chain_2, chain_1,
	}
	err = helpers.ConsensusValidateUnpublishedChains(unpubChains)
	assert.ErrorContains(t, helpers.ErrBadUnpublishedChains.Error(), err)
	assert.ErrorContains(t, "the first values of chain are not equal", err)

	// err "chain is not uniq"
	chain_5 := gwatCommon.HashArray{
		gwatCommon.BytesToHash([]byte{'a', '1'}), gwatCommon.BytesToHash([]byte{'d', '2'}), gwatCommon.BytesToHash([]byte{'d', '2'}),
	}
	unpubChains = []gwatCommon.HashArray{
		chain_5, chain_3, chain_2, chain_1,
	}
	err = helpers.ConsensusValidateUnpublishedChains(unpubChains)
	assert.ErrorContains(t, helpers.ErrBadUnpublishedChains.Error(), err)
	assert.ErrorContains(t, "chain is not uniq", err)
}

func TestConsensusCalculatePefix_OK(t *testing.T) {
	chain_1 := gwatCommon.HashArray{
		gwatCommon.BytesToHash([]byte("a1")),
	}
	chain_2 := gwatCommon.HashArray{
		gwatCommon.BytesToHash([]byte("a1")), gwatCommon.BytesToHash([]byte("b2")),
	}
	chain_3 := gwatCommon.HashArray{
		gwatCommon.BytesToHash([]byte("a1")), gwatCommon.BytesToHash([]byte("a2")),
	}

	unpubChains := []gwatCommon.HashArray{
		chain_3,
		chain_2,
		chain_1,
	}
	want := gwatCommon.HashArray{
		gwatCommon.BytesToHash([]byte("a1")),
	}
	prefix, err := helpers.ConsensusCalcPrefix(unpubChains)
	assert.NoError(t, err)
	assert.Equal(t, fmt.Sprintf("%v", want), fmt.Sprintf("%v", prefix))
}

func TestConsensusCalculatePefix_AdditionalOK(t *testing.T) {
	chain_1 := gwatCommon.HashArray{
		gwatCommon.BytesToHash([]byte("a1")), gwatCommon.BytesToHash([]byte("a2")), gwatCommon.BytesToHash([]byte("a3")), gwatCommon.BytesToHash([]byte("a4")),
	}
	chain_2 := gwatCommon.HashArray{
		gwatCommon.BytesToHash([]byte("a1")), gwatCommon.BytesToHash([]byte("a2")), gwatCommon.BytesToHash([]byte("a3")), gwatCommon.BytesToHash([]byte("a4")),
	}
	chain_3 := gwatCommon.HashArray{
		gwatCommon.BytesToHash([]byte("a1")), gwatCommon.BytesToHash([]byte("a2")), gwatCommon.BytesToHash([]byte("a3")), gwatCommon.BytesToHash([]byte("a4")),
	}
	chain_4 := gwatCommon.HashArray{
		gwatCommon.BytesToHash([]byte("a1")), gwatCommon.BytesToHash([]byte("a2")), gwatCommon.BytesToHash([]byte("b3")), gwatCommon.BytesToHash([]byte("a4")),
	}

	unpubChains := []gwatCommon.HashArray{
		chain_4, chain_3, chain_2, chain_1,
	}
	want := gwatCommon.HashArray{
		gwatCommon.BytesToHash([]byte("a1")), gwatCommon.BytesToHash([]byte("a2")),
	}
	prefix, err := helpers.ConsensusCalcPrefix(unpubChains)
	assert.NoError(t, err)
	assert.Equal(t, fmt.Sprintf("%v", want), fmt.Sprintf("%v", prefix))
}

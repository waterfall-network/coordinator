package wrapper

import (
	"encoding/binary"
	"fmt"

	types "github.com/prysmaticlabs/eth2-types"
	ethpb "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1"
	gwatTypes "gitlab.waterfall.network/waterfall/protocol/gwat/core/types"
)

type GwatSyncParam struct {
	finEpoch   types.Epoch
	checkpoint *ethpb.Checkpoint
	param      *gwatTypes.FinalizationParams
}

func NewGwatSyncParam(checkpoint *ethpb.Checkpoint, param *gwatTypes.FinalizationParams, finEpoch types.Epoch) *GwatSyncParam {
	return &GwatSyncParam{
		finEpoch:   finEpoch,
		checkpoint: checkpoint,
		param:      param,
	}
}

func (gsp *GwatSyncParam) Bytes() ([]byte, error) {
	finParamBin, err := gsp.param.MarshalJSON()
	if err != nil {
		return nil, err
	}
	cp := gsp.checkpoint
	dataLen := 8 + 8 + 32 + len(finParamBin)
	res := make([]byte, 0, dataLen)

	startEpoch := make([]byte, 8)
	binary.BigEndian.PutUint64(startEpoch, uint64(gsp.finEpoch))
	res = append(res, startEpoch...)

	epoch := make([]byte, 8)
	binary.BigEndian.PutUint64(epoch, uint64(cp.Epoch))
	res = append(res, epoch...)

	res = append(res, cp.Root...)
	res = append(res, finParamBin...)
	return res, nil
}

func BytesToGwatSyncParam(data []byte) (*GwatSyncParam, error) {
	gsp := &GwatSyncParam{
		checkpoint: &ethpb.Checkpoint{},
		param:      &gwatTypes.FinalizationParams{},
	}
	cpLen := 8 + 32
	if len(data) < cpLen {
		return nil, fmt.Errorf("bad bitlen: got=%d req=%d", len(data), cpLen)
	}
	var start, end int

	start = 0
	end += 8
	gsp.finEpoch = types.Epoch(binary.BigEndian.Uint64(data[start:end]))

	start = end
	end += 8
	gsp.checkpoint.Epoch = types.Epoch(binary.BigEndian.Uint64(data[start:end]))

	start = end
	end += 32
	gsp.checkpoint.Root = data[start:end]

	start = end
	finParamBin := data[start:]
	err := gsp.param.UnmarshalJSON(finParamBin)
	if err != nil {
		return nil, err
	}
	return gsp, nil
}

func (gsp *GwatSyncParam) FinEpoch() types.Epoch {
	return gsp.finEpoch
}
func (gsp *GwatSyncParam) Epoch() types.Epoch {
	return gsp.checkpoint.Epoch
}
func (gsp *GwatSyncParam) Root() []byte {
	return gsp.checkpoint.Root
}
func (gsp *GwatSyncParam) Param() *gwatTypes.FinalizationParams {
	return gsp.param
}
func (gsp *GwatSyncParam) Checkpoint() *ethpb.Checkpoint {
	return gsp.checkpoint
}

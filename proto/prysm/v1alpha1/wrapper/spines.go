package wrapper

import (
	gwatCommon "gitlab.waterfall.network/waterfall/protocol/gwat/common"
)

type Spines []byte

func (s Spines) HashArray() gwatCommon.HashArray {
	if len(s) == 0 {
		return gwatCommon.HashArray{}
	}
	return gwatCommon.HashArrayFromBytes(s[:])
}

func (s Spines) Key() [32]byte {
	if len(s) == 0 {
		return [32]byte{}
	}
	return s.HashArray().Key()
}

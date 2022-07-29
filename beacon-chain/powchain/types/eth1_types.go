package types

import (
	"errors"
	"math/big"

	"github.com/prysmaticlabs/prysm/encoding/bytesutil"
	"github.com/waterfall-foundation/gwat/common"
	gethTypes "github.com/waterfall-foundation/gwat/core/types"
)

// HeaderInfo specifies the block header information in the ETH 1.0 chain.
type HeaderInfo struct {
	Number *big.Int
	Hash   common.Hash
	Time   uint64
}

// HeaderToHeaderInfo converts an eth1 header to a header metadata type.
func HeaderToHeaderInfo(hdr *gethTypes.Header) (*HeaderInfo, error) {
	if hdr.Number == nil {
		// A nil number will panic when calling *big.Int.Set(...)
		return nil, errors.New("cannot convert block header with nil block number")
	}

	return &HeaderInfo{
		Hash:   hdr.Hash(),
		Number: new(big.Int).SetUint64(hdr.Nr()),
		Time:   hdr.Time,
	}, nil
}

// Copy sends out a copy of the current header info.
func (h *HeaderInfo) Copy() *HeaderInfo {
	return &HeaderInfo{
		Hash:   bytesutil.ToBytes32(h.Hash[:]),
		Number: new(big.Int).Set(h.Number),
		Time:   h.Time,
	}
}

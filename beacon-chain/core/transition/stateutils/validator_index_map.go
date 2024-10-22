// Package stateutils contains useful tools for faster computation
// of state transitions using maps to represent validators instead
// of slices.
package stateutils

import (
	types "github.com/prysmaticlabs/eth2-types"
	fieldparams "gitlab.waterfall.network/waterfall/protocol/coordinator/config/fieldparams"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/encoding/bytesutil"
	ethpb "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1"
)

// ValidatorIndexMap builds a lookup map for quickly determining the index of
// a validator by their public key.
func ValidatorIndexMap(validators []*ethpb.Validator) map[[fieldparams.BLSPubkeyLength]byte]types.ValidatorIndex {
	m := make(map[[fieldparams.BLSPubkeyLength]byte]types.ValidatorIndex, len(validators))
	if validators == nil {
		return m
	}
	for idx, record := range validators {
		if record == nil {
			continue
		}
		var key [48]byte
		if ptrKey := bytesutil.ToPtrBytes48(record.PublicKey); ptrKey != nil {
			key = *ptrKey
		} else {
			key = bytesutil.ToBytes48(record.PublicKey)
		}
		m[key] = types.ValidatorIndex(idx)
	}
	return m
}

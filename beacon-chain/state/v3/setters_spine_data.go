package v3

import (
	ethpb "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1"
)

// SetSpineData for the beacon state.
func (b *BeaconState) SetSpineData(val *ethpb.SpineData) error {
	if !b.hasInnerState() {
		return ErrNilInnerState
	}
	b.lock.Lock()
	defer b.lock.Unlock()

	b.state.SpineData = val
	b.markFieldAsDirty(spineData)
	return nil
}

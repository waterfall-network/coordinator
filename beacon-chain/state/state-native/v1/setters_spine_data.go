package v1

import (
	ethpb "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1"
)

// SetSpineData for the beacon state.
func (b *BeaconState) SetSpineData(val *ethpb.SpineData) error {
	b.lock.Lock()
	defer b.lock.Unlock()

	b.spineData = val
	b.markFieldAsDirty(spineData)
	return nil
}

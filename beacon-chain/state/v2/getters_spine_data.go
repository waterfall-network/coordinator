package v2

import (
	ethpb "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1"
)

// SpineData obtain spine information stored in the beacon state.
func (b *BeaconState) SpineData() *ethpb.SpineData {
	if !b.hasInnerState() {
		return nil
	}
	if b.state.SpineData == nil {
		return nil
	}

	b.lock.RLock()
	defer b.lock.RUnlock()

	return b.spineData()
}

// spineDataVal corresponding to spine information stored in the beacon state.
// This assumes that a lock is already held on BeaconState.
func (b *BeaconState) spineData() *ethpb.SpineData {
	if !b.hasInnerState() {
		return nil
	}
	if b.state.SpineData == nil {
		return nil
	}

	return ethpb.CopySpineData(b.state.SpineData)
}

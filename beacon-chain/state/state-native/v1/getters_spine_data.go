package v1

import (
	ethpb "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1"
)

// SpineData obtain spine information stored in the beacon state.
func (b *BeaconState) SpineData() *ethpb.SpineData {
	if b.spineData == nil {
		return &ethpb.SpineData{}
	}
	b.lock.RLock()
	defer b.lock.RUnlock()

	return b.spineDataVal()
}

// spineDataVal corresponding to spine information stored in the beacon state.
// This assumes that a lock is already held on BeaconState.
func (b *BeaconState) spineDataVal() *ethpb.SpineData {
	if b.spineData == nil {
		return nil
	}

	return ethpb.CopySpineData(b.spineData)
}

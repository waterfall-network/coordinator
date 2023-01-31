// Code generated by fastssz. DO NOT EDIT.
// Hash: c8c7890d7f89a5b2e82d576c7f6c5b7c7938e9fb929cdcdb00971b8f1a8896fd
package v1

import (
	ssz "github.com/ferranbt/fastssz"
	eth2types "github.com/prysmaticlabs/eth2-types"
	ethpb "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1"
)

// MarshalSSZ ssz marshals the BeaconState object
func (b *BeaconState) MarshalSSZ() ([]byte, error) {
	return ssz.MarshalSSZ(b)
}

// MarshalSSZTo ssz marshals the BeaconState object to a target array
func (b *BeaconState) MarshalSSZTo(buf []byte) (dst []byte, err error) {
	dst = buf
	offset := int(2687313)

	// Field (0) 'genesisTime'
	dst = ssz.MarshalUint64(dst, b.genesisTime)

	// Field (1) 'genesisValidatorsRoot'
	dst = append(dst, b.genesisValidatorsRoot[:]...)

	// Field (2) 'slot'
	dst = ssz.MarshalUint64(dst, uint64(b.slot))

	// Field (3) 'fork'
	if b.fork == nil {
		b.fork = new(ethpb.Fork)
	}
	if dst, err = b.fork.MarshalSSZTo(dst); err != nil {
		return
	}

	// Field (4) 'latestBlockHeader'
	if b.latestBlockHeader == nil {
		b.latestBlockHeader = new(ethpb.BeaconBlockHeader)
	}
	if dst, err = b.latestBlockHeader.MarshalSSZTo(dst); err != nil {
		return
	}

	// Field (5) 'blockRoots'
	for ii := 0; ii < 8192; ii++ {
		dst = append(dst, b.blockRoots[ii][:]...)
	}

	// Field (6) 'stateRoots'
	for ii := 0; ii < 8192; ii++ {
		dst = append(dst, b.stateRoots[ii][:]...)
	}

	// Offset (7) 'historicalRoots'
	dst = ssz.WriteOffset(dst, offset)
	offset += len(b.historicalRoots) * 32

	// Offset (8) 'eth1Data'
	dst = ssz.WriteOffset(dst, offset)
	if b.eth1Data == nil {
		b.eth1Data = new(ethpb.Eth1Data)
	}
	offset += b.eth1Data.SizeSSZ()

	// Offset (9) 'eth1DataVotes'
	dst = ssz.WriteOffset(dst, offset)
	for ii := 0; ii < len(b.eth1DataVotes); ii++ {
		offset += 4
		offset += b.eth1DataVotes[ii].SizeSSZ()
	}

	// Field (10) 'eth1DepositIndex'
	dst = ssz.MarshalUint64(dst, b.eth1DepositIndex)

	// Offset (11) 'blockVoting'
	dst = ssz.WriteOffset(dst, offset)
	for ii := 0; ii < len(b.blockVoting); ii++ {
		offset += 4
		offset += b.blockVoting[ii].SizeSSZ()
	}

	// Offset (12) 'validators'
	dst = ssz.WriteOffset(dst, offset)
	offset += len(b.validators) * 121

	// Offset (13) 'balances'
	dst = ssz.WriteOffset(dst, offset)
	offset += len(b.balances) * 8

	// Field (14) 'randaoMixes'
	for ii := 0; ii < 65536; ii++ {
		dst = append(dst, b.randaoMixes[ii][:]...)
	}

	// Field (15) 'slashings'
	if len(b.slashings) != 8192 {
		err = ssz.ErrVectorLength
		return
	}
	for ii := 0; ii < 8192; ii++ {
		dst = ssz.MarshalUint64(dst, b.slashings[ii])
	}

	// Offset (16) 'previousEpochAttestations'
	dst = ssz.WriteOffset(dst, offset)
	for ii := 0; ii < len(b.previousEpochAttestations); ii++ {
		offset += 4
		offset += b.previousEpochAttestations[ii].SizeSSZ()
	}

	// Offset (17) 'currentEpochAttestations'
	dst = ssz.WriteOffset(dst, offset)
	for ii := 0; ii < len(b.currentEpochAttestations); ii++ {
		offset += 4
		offset += b.currentEpochAttestations[ii].SizeSSZ()
	}

	// Field (18) 'justificationBits'
	if len(b.justificationBits) != 1 {
		err = ssz.ErrBytesLength
		return
	}
	dst = append(dst, b.justificationBits...)

	// Field (19) 'previousJustifiedCheckpoint'
	if b.previousJustifiedCheckpoint == nil {
		b.previousJustifiedCheckpoint = new(ethpb.Checkpoint)
	}
	if dst, err = b.previousJustifiedCheckpoint.MarshalSSZTo(dst); err != nil {
		return
	}

	// Field (20) 'currentJustifiedCheckpoint'
	if b.currentJustifiedCheckpoint == nil {
		b.currentJustifiedCheckpoint = new(ethpb.Checkpoint)
	}
	if dst, err = b.currentJustifiedCheckpoint.MarshalSSZTo(dst); err != nil {
		return
	}

	// Field (21) 'finalizedCheckpoint'
	if b.finalizedCheckpoint == nil {
		b.finalizedCheckpoint = new(ethpb.Checkpoint)
	}
	if dst, err = b.finalizedCheckpoint.MarshalSSZTo(dst); err != nil {
		return
	}

	// Field (7) 'historicalRoots'
	if len(b.historicalRoots) > 16777216 {
		err = ssz.ErrListTooBig
		return
	}
	for ii := 0; ii < len(b.historicalRoots); ii++ {
		dst = append(dst, b.historicalRoots[ii][:]...)
	}

	// Field (8) 'eth1Data'
	if dst, err = b.eth1Data.MarshalSSZTo(dst); err != nil {
		return
	}

	// Field (9) 'eth1DataVotes'
	if len(b.eth1DataVotes) > 2048 {
		err = ssz.ErrListTooBig
		return
	}
	{
		offset = 4 * len(b.eth1DataVotes)
		for ii := 0; ii < len(b.eth1DataVotes); ii++ {
			dst = ssz.WriteOffset(dst, offset)
			offset += b.eth1DataVotes[ii].SizeSSZ()
		}
	}
	for ii := 0; ii < len(b.eth1DataVotes); ii++ {
		if dst, err = b.eth1DataVotes[ii].MarshalSSZTo(dst); err != nil {
			return
		}
	}

	// Field (11) 'blockVoting'
	if len(b.blockVoting) > 2048 {
		err = ssz.ErrListTooBig
		return
	}
	{
		offset = 4 * len(b.blockVoting)
		for ii := 0; ii < len(b.blockVoting); ii++ {
			dst = ssz.WriteOffset(dst, offset)
			offset += b.blockVoting[ii].SizeSSZ()
		}
	}
	for ii := 0; ii < len(b.blockVoting); ii++ {
		if dst, err = b.blockVoting[ii].MarshalSSZTo(dst); err != nil {
			return
		}
	}

	// Field (12) 'validators'
	if len(b.validators) > 1099511627776 {
		err = ssz.ErrListTooBig
		return
	}
	for ii := 0; ii < len(b.validators); ii++ {
		if dst, err = b.validators[ii].MarshalSSZTo(dst); err != nil {
			return
		}
	}

	// Field (13) 'balances'
	if len(b.balances) > 1099511627776 {
		err = ssz.ErrListTooBig
		return
	}
	for ii := 0; ii < len(b.balances); ii++ {
		dst = ssz.MarshalUint64(dst, b.balances[ii])
	}

	// Field (16) 'previousEpochAttestations'
	if len(b.previousEpochAttestations) > 4096 {
		err = ssz.ErrListTooBig
		return
	}
	{
		offset = 4 * len(b.previousEpochAttestations)
		for ii := 0; ii < len(b.previousEpochAttestations); ii++ {
			dst = ssz.WriteOffset(dst, offset)
			offset += b.previousEpochAttestations[ii].SizeSSZ()
		}
	}
	for ii := 0; ii < len(b.previousEpochAttestations); ii++ {
		if dst, err = b.previousEpochAttestations[ii].MarshalSSZTo(dst); err != nil {
			return
		}
	}

	// Field (17) 'currentEpochAttestations'
	if len(b.currentEpochAttestations) > 4096 {
		err = ssz.ErrListTooBig
		return
	}
	{
		offset = 4 * len(b.currentEpochAttestations)
		for ii := 0; ii < len(b.currentEpochAttestations); ii++ {
			dst = ssz.WriteOffset(dst, offset)
			offset += b.currentEpochAttestations[ii].SizeSSZ()
		}
	}
	for ii := 0; ii < len(b.currentEpochAttestations); ii++ {
		if dst, err = b.currentEpochAttestations[ii].MarshalSSZTo(dst); err != nil {
			return
		}
	}

	return
}

// UnmarshalSSZ ssz unmarshals the BeaconState object
func (b *BeaconState) UnmarshalSSZ(buf []byte) error {
	var err error
	size := uint64(len(buf))
	if size < 2687313 {
		return ssz.ErrSize
	}

	tail := buf
	var o7, o8, o9, o11, o12, o13, o16, o17 uint64

	// Field (0) 'genesisTime'
	b.genesisTime = ssz.UnmarshallUint64(buf[0:8])

	// Field (1) 'genesisValidatorsRoot'
	copy(b.genesisValidatorsRoot[:], buf[8:40])

	// Field (2) 'slot'
	b.slot = eth2types.Slot(ssz.UnmarshallUint64(buf[40:48]))

	// Field (3) 'fork'
	if b.fork == nil {
		b.fork = new(ethpb.Fork)
	}
	if err = b.fork.UnmarshalSSZ(buf[48:64]); err != nil {
		return err
	}

	// Field (4) 'latestBlockHeader'
	if b.latestBlockHeader == nil {
		b.latestBlockHeader = new(ethpb.BeaconBlockHeader)
	}
	if err = b.latestBlockHeader.UnmarshalSSZ(buf[64:176]); err != nil {
		return err
	}

	// Field (5) 'blockRoots'

	for ii := 0; ii < 8192; ii++ {
		copy(b.blockRoots[ii][:], buf[176:262320][ii*32:(ii+1)*32])
	}

	// Field (6) 'stateRoots'

	for ii := 0; ii < 8192; ii++ {
		copy(b.stateRoots[ii][:], buf[262320:524464][ii*32:(ii+1)*32])
	}

	// Offset (7) 'historicalRoots'
	if o7 = ssz.ReadOffset(buf[524464:524468]); o7 > size {
		return ssz.ErrOffset
	}

	if o7 < 2687313 {
		return ssz.ErrInvalidVariableOffset
	}

	// Offset (8) 'eth1Data'
	if o8 = ssz.ReadOffset(buf[524468:524472]); o8 > size || o7 > o8 {
		return ssz.ErrOffset
	}

	// Offset (9) 'eth1DataVotes'
	if o9 = ssz.ReadOffset(buf[524472:524476]); o9 > size || o8 > o9 {
		return ssz.ErrOffset
	}

	// Field (10) 'eth1DepositIndex'
	b.eth1DepositIndex = ssz.UnmarshallUint64(buf[524476:524484])

	// Offset (11) 'blockVoting'
	if o11 = ssz.ReadOffset(buf[524484:524488]); o11 > size || o9 > o11 {
		return ssz.ErrOffset
	}

	// Offset (12) 'validators'
	if o12 = ssz.ReadOffset(buf[524488:524492]); o12 > size || o11 > o12 {
		return ssz.ErrOffset
	}

	// Offset (13) 'balances'
	if o13 = ssz.ReadOffset(buf[524492:524496]); o13 > size || o12 > o13 {
		return ssz.ErrOffset
	}

	// Field (14) 'randaoMixes'

	for ii := 0; ii < 65536; ii++ {
		copy(b.randaoMixes[ii][:], buf[524496:2621648][ii*32:(ii+1)*32])
	}

	// Field (15) 'slashings'
	b.slashings = ssz.ExtendUint64(b.slashings, 8192)
	for ii := 0; ii < 8192; ii++ {
		b.slashings[ii] = ssz.UnmarshallUint64(buf[2621648:2687184][ii*8 : (ii+1)*8])
	}

	// Offset (16) 'previousEpochAttestations'
	if o16 = ssz.ReadOffset(buf[2687184:2687188]); o16 > size || o13 > o16 {
		return ssz.ErrOffset
	}

	// Offset (17) 'currentEpochAttestations'
	if o17 = ssz.ReadOffset(buf[2687188:2687192]); o17 > size || o16 > o17 {
		return ssz.ErrOffset
	}

	// Field (18) 'justificationBits'
	if cap(b.justificationBits) == 0 {
		b.justificationBits = make([]byte, 0, len(buf[2687192:2687193]))
	}
	b.justificationBits = append(b.justificationBits, buf[2687192:2687193]...)

	// Field (19) 'previousJustifiedCheckpoint'
	if b.previousJustifiedCheckpoint == nil {
		b.previousJustifiedCheckpoint = new(ethpb.Checkpoint)
	}
	if err = b.previousJustifiedCheckpoint.UnmarshalSSZ(buf[2687193:2687233]); err != nil {
		return err
	}

	// Field (20) 'currentJustifiedCheckpoint'
	if b.currentJustifiedCheckpoint == nil {
		b.currentJustifiedCheckpoint = new(ethpb.Checkpoint)
	}
	if err = b.currentJustifiedCheckpoint.UnmarshalSSZ(buf[2687233:2687273]); err != nil {
		return err
	}

	// Field (21) 'finalizedCheckpoint'
	if b.finalizedCheckpoint == nil {
		b.finalizedCheckpoint = new(ethpb.Checkpoint)
	}
	if err = b.finalizedCheckpoint.UnmarshalSSZ(buf[2687273:2687313]); err != nil {
		return err
	}

	// Field (7) 'historicalRoots'
	{
		buf = tail[o7:o8]
		num, err := ssz.DivideInt2(len(buf), 32, 16777216)
		if err != nil {
			return err
		}
		b.historicalRoots = make([][32]byte, num)
		for ii := 0; ii < num; ii++ {
			copy(b.historicalRoots[ii][:], buf[ii*32:(ii+1)*32])
		}
	}

	// Field (8) 'eth1Data'
	{
		buf = tail[o8:o9]
		if b.eth1Data == nil {
			b.eth1Data = new(ethpb.Eth1Data)
		}
		if err = b.eth1Data.UnmarshalSSZ(buf); err != nil {
			return err
		}
	}

	// Field (9) 'eth1DataVotes'
	{
		buf = tail[o9:o11]
		num, err := ssz.DecodeDynamicLength(buf, 2048)
		if err != nil {
			return err
		}
		b.eth1DataVotes = make([]*ethpb.Eth1Data, num)
		err = ssz.UnmarshalDynamic(buf, num, func(indx int, buf []byte) (err error) {
			if b.eth1DataVotes[indx] == nil {
				b.eth1DataVotes[indx] = new(ethpb.Eth1Data)
			}
			if err = b.eth1DataVotes[indx].UnmarshalSSZ(buf); err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return err
		}
	}

	// Field (11) 'blockVoting'
	{
		buf = tail[o11:o12]
		num, err := ssz.DecodeDynamicLength(buf, 2048)
		if err != nil {
			return err
		}
		b.blockVoting = make([]*ethpb.BlockVoting, num)
		err = ssz.UnmarshalDynamic(buf, num, func(indx int, buf []byte) (err error) {
			if b.blockVoting[indx] == nil {
				b.blockVoting[indx] = new(ethpb.BlockVoting)
			}
			if err = b.blockVoting[indx].UnmarshalSSZ(buf); err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return err
		}
	}

	// Field (12) 'validators'
	{
		buf = tail[o12:o13]
		num, err := ssz.DivideInt2(len(buf), 121, 1099511627776)
		if err != nil {
			return err
		}
		b.validators = make([]*ethpb.Validator, num)
		for ii := 0; ii < num; ii++ {
			if b.validators[ii] == nil {
				b.validators[ii] = new(ethpb.Validator)
			}
			if err = b.validators[ii].UnmarshalSSZ(buf[ii*121 : (ii+1)*121]); err != nil {
				return err
			}
		}
	}

	// Field (13) 'balances'
	{
		buf = tail[o13:o16]
		num, err := ssz.DivideInt2(len(buf), 8, 1099511627776)
		if err != nil {
			return err
		}
		b.balances = ssz.ExtendUint64(b.balances, num)
		for ii := 0; ii < num; ii++ {
			b.balances[ii] = ssz.UnmarshallUint64(buf[ii*8 : (ii+1)*8])
		}
	}

	// Field (16) 'previousEpochAttestations'
	{
		buf = tail[o16:o17]
		num, err := ssz.DecodeDynamicLength(buf, 4096)
		if err != nil {
			return err
		}
		b.previousEpochAttestations = make([]*ethpb.PendingAttestation, num)
		err = ssz.UnmarshalDynamic(buf, num, func(indx int, buf []byte) (err error) {
			if b.previousEpochAttestations[indx] == nil {
				b.previousEpochAttestations[indx] = new(ethpb.PendingAttestation)
			}
			if err = b.previousEpochAttestations[indx].UnmarshalSSZ(buf); err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return err
		}
	}

	// Field (17) 'currentEpochAttestations'
	{
		buf = tail[o17:]
		num, err := ssz.DecodeDynamicLength(buf, 4096)
		if err != nil {
			return err
		}
		b.currentEpochAttestations = make([]*ethpb.PendingAttestation, num)
		err = ssz.UnmarshalDynamic(buf, num, func(indx int, buf []byte) (err error) {
			if b.currentEpochAttestations[indx] == nil {
				b.currentEpochAttestations[indx] = new(ethpb.PendingAttestation)
			}
			if err = b.currentEpochAttestations[indx].UnmarshalSSZ(buf); err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return err
		}
	}
	return err
}

// SizeSSZ returns the ssz encoded size in bytes for the BeaconState object
func (b *BeaconState) SizeSSZ() (size int) {
	size = 2687313

	// Field (7) 'historicalRoots'
	size += len(b.historicalRoots) * 32

	// Field (8) 'eth1Data'
	if b.eth1Data == nil {
		b.eth1Data = new(ethpb.Eth1Data)
	}
	size += b.eth1Data.SizeSSZ()

	// Field (9) 'eth1DataVotes'
	for ii := 0; ii < len(b.eth1DataVotes); ii++ {
		size += 4
		size += b.eth1DataVotes[ii].SizeSSZ()
	}

	// Field (11) 'blockVoting'
	for ii := 0; ii < len(b.blockVoting); ii++ {
		size += 4
		size += b.blockVoting[ii].SizeSSZ()
	}

	// Field (12) 'validators'
	size += len(b.validators) * 121

	// Field (13) 'balances'
	size += len(b.balances) * 8

	// Field (16) 'previousEpochAttestations'
	for ii := 0; ii < len(b.previousEpochAttestations); ii++ {
		size += 4
		size += b.previousEpochAttestations[ii].SizeSSZ()
	}

	// Field (17) 'currentEpochAttestations'
	for ii := 0; ii < len(b.currentEpochAttestations); ii++ {
		size += 4
		size += b.currentEpochAttestations[ii].SizeSSZ()
	}

	return
}

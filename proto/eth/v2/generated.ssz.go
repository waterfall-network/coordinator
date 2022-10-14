// Code generated by fastssz. DO NOT EDIT.
// Hash: 9b7f29cef312ab31b0f542a125c746957975f8406c6e899e9b46a5a3015fd891
package eth

import (
	ssz "github.com/ferranbt/fastssz"
	github_com_prysmaticlabs_eth2_types "github.com/prysmaticlabs/eth2-types"
	v11 "github.com/waterfall-foundation/coordinator/proto/engine/v1"
	v1 "github.com/waterfall-foundation/coordinator/proto/eth/v1"
)

// MarshalSSZ ssz marshals the SignedBeaconBlockBellatrix object
func (s *SignedBeaconBlockBellatrix) MarshalSSZ() ([]byte, error) {
	return ssz.MarshalSSZ(s)
}

// MarshalSSZTo ssz marshals the SignedBeaconBlockBellatrix object to a target array
func (s *SignedBeaconBlockBellatrix) MarshalSSZTo(buf []byte) (dst []byte, err error) {
	dst = buf
	offset := int(100)

	// Offset (0) 'Message'
	dst = ssz.WriteOffset(dst, offset)
	if s.Message == nil {
		s.Message = new(BeaconBlockBellatrix)
	}
	offset += s.Message.SizeSSZ()

	// Field (1) 'Signature'
	if len(s.Signature) != 96 {
		err = ssz.ErrBytesLength
		return
	}
	dst = append(dst, s.Signature...)

	// Field (0) 'Message'
	if dst, err = s.Message.MarshalSSZTo(dst); err != nil {
		return
	}

	return
}

// UnmarshalSSZ ssz unmarshals the SignedBeaconBlockBellatrix object
func (s *SignedBeaconBlockBellatrix) UnmarshalSSZ(buf []byte) error {
	var err error
	size := uint64(len(buf))
	if size < 100 {
		return ssz.ErrSize
	}

	tail := buf
	var o0 uint64

	// Offset (0) 'Message'
	if o0 = ssz.ReadOffset(buf[0:4]); o0 > size {
		return ssz.ErrOffset
	}

	if o0 < 100 {
		return ssz.ErrInvalidVariableOffset
	}

	// Field (1) 'Signature'
	if cap(s.Signature) == 0 {
		s.Signature = make([]byte, 0, len(buf[4:100]))
	}
	s.Signature = append(s.Signature, buf[4:100]...)

	// Field (0) 'Message'
	{
		buf = tail[o0:]
		if s.Message == nil {
			s.Message = new(BeaconBlockBellatrix)
		}
		if err = s.Message.UnmarshalSSZ(buf); err != nil {
			return err
		}
	}
	return err
}

// SizeSSZ returns the ssz encoded size in bytes for the SignedBeaconBlockBellatrix object
func (s *SignedBeaconBlockBellatrix) SizeSSZ() (size int) {
	size = 100

	// Field (0) 'Message'
	if s.Message == nil {
		s.Message = new(BeaconBlockBellatrix)
	}
	size += s.Message.SizeSSZ()

	return
}

// HashTreeRoot ssz hashes the SignedBeaconBlockBellatrix object
func (s *SignedBeaconBlockBellatrix) HashTreeRoot() ([32]byte, error) {
	return ssz.HashWithDefaultHasher(s)
}

// HashTreeRootWith ssz hashes the SignedBeaconBlockBellatrix object with a hasher
func (s *SignedBeaconBlockBellatrix) HashTreeRootWith(hh *ssz.Hasher) (err error) {
	indx := hh.Index()

	// Field (0) 'Message'
	if err = s.Message.HashTreeRootWith(hh); err != nil {
		return
	}

	// Field (1) 'Signature'
	if len(s.Signature) != 96 {
		err = ssz.ErrBytesLength
		return
	}
	hh.PutBytes(s.Signature)

	hh.Merkleize(indx)
	return
}

// MarshalSSZ ssz marshals the SignedBeaconBlockAltair object
func (s *SignedBeaconBlockAltair) MarshalSSZ() ([]byte, error) {
	return ssz.MarshalSSZ(s)
}

// MarshalSSZTo ssz marshals the SignedBeaconBlockAltair object to a target array
func (s *SignedBeaconBlockAltair) MarshalSSZTo(buf []byte) (dst []byte, err error) {
	dst = buf
	offset := int(100)

	// Offset (0) 'Message'
	dst = ssz.WriteOffset(dst, offset)
	if s.Message == nil {
		s.Message = new(BeaconBlockAltair)
	}
	offset += s.Message.SizeSSZ()

	// Field (1) 'Signature'
	if len(s.Signature) != 96 {
		err = ssz.ErrBytesLength
		return
	}
	dst = append(dst, s.Signature...)

	// Field (0) 'Message'
	if dst, err = s.Message.MarshalSSZTo(dst); err != nil {
		return
	}

	return
}

// UnmarshalSSZ ssz unmarshals the SignedBeaconBlockAltair object
func (s *SignedBeaconBlockAltair) UnmarshalSSZ(buf []byte) error {
	var err error
	size := uint64(len(buf))
	if size < 100 {
		return ssz.ErrSize
	}

	tail := buf
	var o0 uint64

	// Offset (0) 'Message'
	if o0 = ssz.ReadOffset(buf[0:4]); o0 > size {
		return ssz.ErrOffset
	}

	if o0 < 100 {
		return ssz.ErrInvalidVariableOffset
	}

	// Field (1) 'Signature'
	if cap(s.Signature) == 0 {
		s.Signature = make([]byte, 0, len(buf[4:100]))
	}
	s.Signature = append(s.Signature, buf[4:100]...)

	// Field (0) 'Message'
	{
		buf = tail[o0:]
		if s.Message == nil {
			s.Message = new(BeaconBlockAltair)
		}
		if err = s.Message.UnmarshalSSZ(buf); err != nil {
			return err
		}
	}
	return err
}

// SizeSSZ returns the ssz encoded size in bytes for the SignedBeaconBlockAltair object
func (s *SignedBeaconBlockAltair) SizeSSZ() (size int) {
	size = 100

	// Field (0) 'Message'
	if s.Message == nil {
		s.Message = new(BeaconBlockAltair)
	}
	size += s.Message.SizeSSZ()

	return
}

// HashTreeRoot ssz hashes the SignedBeaconBlockAltair object
func (s *SignedBeaconBlockAltair) HashTreeRoot() ([32]byte, error) {
	return ssz.HashWithDefaultHasher(s)
}

// HashTreeRootWith ssz hashes the SignedBeaconBlockAltair object with a hasher
func (s *SignedBeaconBlockAltair) HashTreeRootWith(hh *ssz.Hasher) (err error) {
	indx := hh.Index()

	// Field (0) 'Message'
	if err = s.Message.HashTreeRootWith(hh); err != nil {
		return
	}

	// Field (1) 'Signature'
	if len(s.Signature) != 96 {
		err = ssz.ErrBytesLength
		return
	}
	hh.PutBytes(s.Signature)

	hh.Merkleize(indx)
	return
}

// MarshalSSZ ssz marshals the BeaconBlockBellatrix object
func (b *BeaconBlockBellatrix) MarshalSSZ() ([]byte, error) {
	return ssz.MarshalSSZ(b)
}

// MarshalSSZTo ssz marshals the BeaconBlockBellatrix object to a target array
func (b *BeaconBlockBellatrix) MarshalSSZTo(buf []byte) (dst []byte, err error) {
	dst = buf
	offset := int(84)

	// Field (0) 'Slot'
	dst = ssz.MarshalUint64(dst, uint64(b.Slot))

	// Field (1) 'ProposerIndex'
	dst = ssz.MarshalUint64(dst, uint64(b.ProposerIndex))

	// Field (2) 'ParentRoot'
	if len(b.ParentRoot) != 32 {
		err = ssz.ErrBytesLength
		return
	}
	dst = append(dst, b.ParentRoot...)

	// Field (3) 'StateRoot'
	if len(b.StateRoot) != 32 {
		err = ssz.ErrBytesLength
		return
	}
	dst = append(dst, b.StateRoot...)

	// Offset (4) 'Body'
	dst = ssz.WriteOffset(dst, offset)
	if b.Body == nil {
		b.Body = new(BeaconBlockBodyBellatrix)
	}
	offset += b.Body.SizeSSZ()

	// Field (4) 'Body'
	if dst, err = b.Body.MarshalSSZTo(dst); err != nil {
		return
	}

	return
}

// UnmarshalSSZ ssz unmarshals the BeaconBlockBellatrix object
func (b *BeaconBlockBellatrix) UnmarshalSSZ(buf []byte) error {
	var err error
	size := uint64(len(buf))
	if size < 84 {
		return ssz.ErrSize
	}

	tail := buf
	var o4 uint64

	// Field (0) 'Slot'
	b.Slot = github_com_prysmaticlabs_eth2_types.Slot(ssz.UnmarshallUint64(buf[0:8]))

	// Field (1) 'ProposerIndex'
	b.ProposerIndex = github_com_prysmaticlabs_eth2_types.ValidatorIndex(ssz.UnmarshallUint64(buf[8:16]))

	// Field (2) 'ParentRoot'
	if cap(b.ParentRoot) == 0 {
		b.ParentRoot = make([]byte, 0, len(buf[16:48]))
	}
	b.ParentRoot = append(b.ParentRoot, buf[16:48]...)

	// Field (3) 'StateRoot'
	if cap(b.StateRoot) == 0 {
		b.StateRoot = make([]byte, 0, len(buf[48:80]))
	}
	b.StateRoot = append(b.StateRoot, buf[48:80]...)

	// Offset (4) 'Body'
	if o4 = ssz.ReadOffset(buf[80:84]); o4 > size {
		return ssz.ErrOffset
	}

	if o4 < 84 {
		return ssz.ErrInvalidVariableOffset
	}

	// Field (4) 'Body'
	{
		buf = tail[o4:]
		if b.Body == nil {
			b.Body = new(BeaconBlockBodyBellatrix)
		}
		if err = b.Body.UnmarshalSSZ(buf); err != nil {
			return err
		}
	}
	return err
}

// SizeSSZ returns the ssz encoded size in bytes for the BeaconBlockBellatrix object
func (b *BeaconBlockBellatrix) SizeSSZ() (size int) {
	size = 84

	// Field (4) 'Body'
	if b.Body == nil {
		b.Body = new(BeaconBlockBodyBellatrix)
	}
	size += b.Body.SizeSSZ()

	return
}

// HashTreeRoot ssz hashes the BeaconBlockBellatrix object
func (b *BeaconBlockBellatrix) HashTreeRoot() ([32]byte, error) {
	return ssz.HashWithDefaultHasher(b)
}

// HashTreeRootWith ssz hashes the BeaconBlockBellatrix object with a hasher
func (b *BeaconBlockBellatrix) HashTreeRootWith(hh *ssz.Hasher) (err error) {
	indx := hh.Index()

	// Field (0) 'Slot'
	hh.PutUint64(uint64(b.Slot))

	// Field (1) 'ProposerIndex'
	hh.PutUint64(uint64(b.ProposerIndex))

	// Field (2) 'ParentRoot'
	if len(b.ParentRoot) != 32 {
		err = ssz.ErrBytesLength
		return
	}
	hh.PutBytes(b.ParentRoot)

	// Field (3) 'StateRoot'
	if len(b.StateRoot) != 32 {
		err = ssz.ErrBytesLength
		return
	}
	hh.PutBytes(b.StateRoot)

	// Field (4) 'Body'
	if err = b.Body.HashTreeRootWith(hh); err != nil {
		return
	}

	hh.Merkleize(indx)
	return
}

// MarshalSSZ ssz marshals the BeaconBlockAltair object
func (b *BeaconBlockAltair) MarshalSSZ() ([]byte, error) {
	return ssz.MarshalSSZ(b)
}

// MarshalSSZTo ssz marshals the BeaconBlockAltair object to a target array
func (b *BeaconBlockAltair) MarshalSSZTo(buf []byte) (dst []byte, err error) {
	dst = buf
	offset := int(84)

	// Field (0) 'Slot'
	dst = ssz.MarshalUint64(dst, uint64(b.Slot))

	// Field (1) 'ProposerIndex'
	dst = ssz.MarshalUint64(dst, uint64(b.ProposerIndex))

	// Field (2) 'ParentRoot'
	if len(b.ParentRoot) != 32 {
		err = ssz.ErrBytesLength
		return
	}
	dst = append(dst, b.ParentRoot...)

	// Field (3) 'StateRoot'
	if len(b.StateRoot) != 32 {
		err = ssz.ErrBytesLength
		return
	}
	dst = append(dst, b.StateRoot...)

	// Offset (4) 'Body'
	dst = ssz.WriteOffset(dst, offset)
	if b.Body == nil {
		b.Body = new(BeaconBlockBodyAltair)
	}
	offset += b.Body.SizeSSZ()

	// Field (4) 'Body'
	if dst, err = b.Body.MarshalSSZTo(dst); err != nil {
		return
	}

	return
}

// UnmarshalSSZ ssz unmarshals the BeaconBlockAltair object
func (b *BeaconBlockAltair) UnmarshalSSZ(buf []byte) error {
	var err error
	size := uint64(len(buf))
	if size < 84 {
		return ssz.ErrSize
	}

	tail := buf
	var o4 uint64

	// Field (0) 'Slot'
	b.Slot = github_com_prysmaticlabs_eth2_types.Slot(ssz.UnmarshallUint64(buf[0:8]))

	// Field (1) 'ProposerIndex'
	b.ProposerIndex = github_com_prysmaticlabs_eth2_types.ValidatorIndex(ssz.UnmarshallUint64(buf[8:16]))

	// Field (2) 'ParentRoot'
	if cap(b.ParentRoot) == 0 {
		b.ParentRoot = make([]byte, 0, len(buf[16:48]))
	}
	b.ParentRoot = append(b.ParentRoot, buf[16:48]...)

	// Field (3) 'StateRoot'
	if cap(b.StateRoot) == 0 {
		b.StateRoot = make([]byte, 0, len(buf[48:80]))
	}
	b.StateRoot = append(b.StateRoot, buf[48:80]...)

	// Offset (4) 'Body'
	if o4 = ssz.ReadOffset(buf[80:84]); o4 > size {
		return ssz.ErrOffset
	}

	if o4 < 84 {
		return ssz.ErrInvalidVariableOffset
	}

	// Field (4) 'Body'
	{
		buf = tail[o4:]
		if b.Body == nil {
			b.Body = new(BeaconBlockBodyAltair)
		}
		if err = b.Body.UnmarshalSSZ(buf); err != nil {
			return err
		}
	}
	return err
}

// SizeSSZ returns the ssz encoded size in bytes for the BeaconBlockAltair object
func (b *BeaconBlockAltair) SizeSSZ() (size int) {
	size = 84

	// Field (4) 'Body'
	if b.Body == nil {
		b.Body = new(BeaconBlockBodyAltair)
	}
	size += b.Body.SizeSSZ()

	return
}

// HashTreeRoot ssz hashes the BeaconBlockAltair object
func (b *BeaconBlockAltair) HashTreeRoot() ([32]byte, error) {
	return ssz.HashWithDefaultHasher(b)
}

// HashTreeRootWith ssz hashes the BeaconBlockAltair object with a hasher
func (b *BeaconBlockAltair) HashTreeRootWith(hh *ssz.Hasher) (err error) {
	indx := hh.Index()

	// Field (0) 'Slot'
	hh.PutUint64(uint64(b.Slot))

	// Field (1) 'ProposerIndex'
	hh.PutUint64(uint64(b.ProposerIndex))

	// Field (2) 'ParentRoot'
	if len(b.ParentRoot) != 32 {
		err = ssz.ErrBytesLength
		return
	}
	hh.PutBytes(b.ParentRoot)

	// Field (3) 'StateRoot'
	if len(b.StateRoot) != 32 {
		err = ssz.ErrBytesLength
		return
	}
	hh.PutBytes(b.StateRoot)

	// Field (4) 'Body'
	if err = b.Body.HashTreeRootWith(hh); err != nil {
		return
	}

	hh.Merkleize(indx)
	return
}

// MarshalSSZ ssz marshals the BeaconBlockBodyBellatrix object
func (b *BeaconBlockBodyBellatrix) MarshalSSZ() ([]byte, error) {
	return ssz.MarshalSSZ(b)
}

// MarshalSSZTo ssz marshals the BeaconBlockBodyBellatrix object to a target array
func (b *BeaconBlockBodyBellatrix) MarshalSSZTo(buf []byte) (dst []byte, err error) {
	dst = buf
	offset := int(316)

	// Field (0) 'RandaoReveal'
	if len(b.RandaoReveal) != 96 {
		err = ssz.ErrBytesLength
		return
	}
	dst = append(dst, b.RandaoReveal...)

	// Offset (1) 'Eth1Data'
	dst = ssz.WriteOffset(dst, offset)
	if b.Eth1Data == nil {
		b.Eth1Data = new(v1.Eth1Data)
	}
	offset += b.Eth1Data.SizeSSZ()

	// Field (2) 'Graffiti'
	if len(b.Graffiti) != 32 {
		err = ssz.ErrBytesLength
		return
	}
	dst = append(dst, b.Graffiti...)

	// Offset (3) 'ProposerSlashings'
	dst = ssz.WriteOffset(dst, offset)
	offset += len(b.ProposerSlashings) * 416

	// Offset (4) 'AttesterSlashings'
	dst = ssz.WriteOffset(dst, offset)
	for ii := 0; ii < len(b.AttesterSlashings); ii++ {
		offset += 4
		offset += b.AttesterSlashings[ii].SizeSSZ()
	}

	// Offset (5) 'Attestations'
	dst = ssz.WriteOffset(dst, offset)
	for ii := 0; ii < len(b.Attestations); ii++ {
		offset += 4
		offset += b.Attestations[ii].SizeSSZ()
	}

	// Offset (6) 'Deposits'
	dst = ssz.WriteOffset(dst, offset)
	offset += len(b.Deposits) * 1240

	// Offset (7) 'VoluntaryExits'
	dst = ssz.WriteOffset(dst, offset)
	offset += len(b.VoluntaryExits) * 112

	// Field (8) 'SyncAggregate'
	if b.SyncAggregate == nil {
		b.SyncAggregate = new(v1.SyncAggregate)
	}
	if dst, err = b.SyncAggregate.MarshalSSZTo(dst); err != nil {
		return
	}

	// Offset (9) 'ExecutionPayload'
	dst = ssz.WriteOffset(dst, offset)
	if b.ExecutionPayload == nil {
		b.ExecutionPayload = new(v11.ExecutionPayload)
	}
	offset += b.ExecutionPayload.SizeSSZ()

	// Field (1) 'Eth1Data'
	if dst, err = b.Eth1Data.MarshalSSZTo(dst); err != nil {
		return
	}

	// Field (3) 'ProposerSlashings'
	if len(b.ProposerSlashings) > 16 {
		err = ssz.ErrListTooBig
		return
	}
	for ii := 0; ii < len(b.ProposerSlashings); ii++ {
		if dst, err = b.ProposerSlashings[ii].MarshalSSZTo(dst); err != nil {
			return
		}
	}

	// Field (4) 'AttesterSlashings'
	if len(b.AttesterSlashings) > 2 {
		err = ssz.ErrListTooBig
		return
	}
	{
		offset = 4 * len(b.AttesterSlashings)
		for ii := 0; ii < len(b.AttesterSlashings); ii++ {
			dst = ssz.WriteOffset(dst, offset)
			offset += b.AttesterSlashings[ii].SizeSSZ()
		}
	}
	for ii := 0; ii < len(b.AttesterSlashings); ii++ {
		if dst, err = b.AttesterSlashings[ii].MarshalSSZTo(dst); err != nil {
			return
		}
	}

	// Field (5) 'Attestations'
	if len(b.Attestations) > 128 {
		err = ssz.ErrListTooBig
		return
	}
	{
		offset = 4 * len(b.Attestations)
		for ii := 0; ii < len(b.Attestations); ii++ {
			dst = ssz.WriteOffset(dst, offset)
			offset += b.Attestations[ii].SizeSSZ()
		}
	}
	for ii := 0; ii < len(b.Attestations); ii++ {
		if dst, err = b.Attestations[ii].MarshalSSZTo(dst); err != nil {
			return
		}
	}

	// Field (6) 'Deposits'
	if len(b.Deposits) > 16 {
		err = ssz.ErrListTooBig
		return
	}
	for ii := 0; ii < len(b.Deposits); ii++ {
		if dst, err = b.Deposits[ii].MarshalSSZTo(dst); err != nil {
			return
		}
	}

	// Field (7) 'VoluntaryExits'
	if len(b.VoluntaryExits) > 16 {
		err = ssz.ErrListTooBig
		return
	}
	for ii := 0; ii < len(b.VoluntaryExits); ii++ {
		if dst, err = b.VoluntaryExits[ii].MarshalSSZTo(dst); err != nil {
			return
		}
	}

	// Field (9) 'ExecutionPayload'
	if dst, err = b.ExecutionPayload.MarshalSSZTo(dst); err != nil {
		return
	}

	return
}

// UnmarshalSSZ ssz unmarshals the BeaconBlockBodyBellatrix object
func (b *BeaconBlockBodyBellatrix) UnmarshalSSZ(buf []byte) error {
	var err error
	size := uint64(len(buf))
	if size < 316 {
		return ssz.ErrSize
	}

	tail := buf
	var o1, o3, o4, o5, o6, o7, o9 uint64

	// Field (0) 'RandaoReveal'
	if cap(b.RandaoReveal) == 0 {
		b.RandaoReveal = make([]byte, 0, len(buf[0:96]))
	}
	b.RandaoReveal = append(b.RandaoReveal, buf[0:96]...)

	// Offset (1) 'Eth1Data'
	if o1 = ssz.ReadOffset(buf[96:100]); o1 > size {
		return ssz.ErrOffset
	}

	if o1 < 316 {
		return ssz.ErrInvalidVariableOffset
	}

	// Field (2) 'Graffiti'
	if cap(b.Graffiti) == 0 {
		b.Graffiti = make([]byte, 0, len(buf[100:132]))
	}
	b.Graffiti = append(b.Graffiti, buf[100:132]...)

	// Offset (3) 'ProposerSlashings'
	if o3 = ssz.ReadOffset(buf[132:136]); o3 > size || o1 > o3 {
		return ssz.ErrOffset
	}

	// Offset (4) 'AttesterSlashings'
	if o4 = ssz.ReadOffset(buf[136:140]); o4 > size || o3 > o4 {
		return ssz.ErrOffset
	}

	// Offset (5) 'Attestations'
	if o5 = ssz.ReadOffset(buf[140:144]); o5 > size || o4 > o5 {
		return ssz.ErrOffset
	}

	// Offset (6) 'Deposits'
	if o6 = ssz.ReadOffset(buf[144:148]); o6 > size || o5 > o6 {
		return ssz.ErrOffset
	}

	// Offset (7) 'VoluntaryExits'
	if o7 = ssz.ReadOffset(buf[148:152]); o7 > size || o6 > o7 {
		return ssz.ErrOffset
	}

	// Field (8) 'SyncAggregate'
	if b.SyncAggregate == nil {
		b.SyncAggregate = new(v1.SyncAggregate)
	}
	if err = b.SyncAggregate.UnmarshalSSZ(buf[152:312]); err != nil {
		return err
	}

	// Offset (9) 'ExecutionPayload'
	if o9 = ssz.ReadOffset(buf[312:316]); o9 > size || o7 > o9 {
		return ssz.ErrOffset
	}

	// Field (1) 'Eth1Data'
	{
		buf = tail[o1:o3]
		if b.Eth1Data == nil {
			b.Eth1Data = new(v1.Eth1Data)
		}
		if err = b.Eth1Data.UnmarshalSSZ(buf); err != nil {
			return err
		}
	}

	// Field (3) 'ProposerSlashings'
	{
		buf = tail[o3:o4]
		num, err := ssz.DivideInt2(len(buf), 416, 16)
		if err != nil {
			return err
		}
		b.ProposerSlashings = make([]*v1.ProposerSlashing, num)
		for ii := 0; ii < num; ii++ {
			if b.ProposerSlashings[ii] == nil {
				b.ProposerSlashings[ii] = new(v1.ProposerSlashing)
			}
			if err = b.ProposerSlashings[ii].UnmarshalSSZ(buf[ii*416 : (ii+1)*416]); err != nil {
				return err
			}
		}
	}

	// Field (4) 'AttesterSlashings'
	{
		buf = tail[o4:o5]
		num, err := ssz.DecodeDynamicLength(buf, 2)
		if err != nil {
			return err
		}
		b.AttesterSlashings = make([]*v1.AttesterSlashing, num)
		err = ssz.UnmarshalDynamic(buf, num, func(indx int, buf []byte) (err error) {
			if b.AttesterSlashings[indx] == nil {
				b.AttesterSlashings[indx] = new(v1.AttesterSlashing)
			}
			if err = b.AttesterSlashings[indx].UnmarshalSSZ(buf); err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return err
		}
	}

	// Field (5) 'Attestations'
	{
		buf = tail[o5:o6]
		num, err := ssz.DecodeDynamicLength(buf, 128)
		if err != nil {
			return err
		}
		b.Attestations = make([]*v1.Attestation, num)
		err = ssz.UnmarshalDynamic(buf, num, func(indx int, buf []byte) (err error) {
			if b.Attestations[indx] == nil {
				b.Attestations[indx] = new(v1.Attestation)
			}
			if err = b.Attestations[indx].UnmarshalSSZ(buf); err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return err
		}
	}

	// Field (6) 'Deposits'
	{
		buf = tail[o6:o7]
		num, err := ssz.DivideInt2(len(buf), 1240, 16)
		if err != nil {
			return err
		}
		b.Deposits = make([]*v1.Deposit, num)
		for ii := 0; ii < num; ii++ {
			if b.Deposits[ii] == nil {
				b.Deposits[ii] = new(v1.Deposit)
			}
			if err = b.Deposits[ii].UnmarshalSSZ(buf[ii*1240 : (ii+1)*1240]); err != nil {
				return err
			}
		}
	}

	// Field (7) 'VoluntaryExits'
	{
		buf = tail[o7:o9]
		num, err := ssz.DivideInt2(len(buf), 112, 16)
		if err != nil {
			return err
		}
		b.VoluntaryExits = make([]*v1.SignedVoluntaryExit, num)
		for ii := 0; ii < num; ii++ {
			if b.VoluntaryExits[ii] == nil {
				b.VoluntaryExits[ii] = new(v1.SignedVoluntaryExit)
			}
			if err = b.VoluntaryExits[ii].UnmarshalSSZ(buf[ii*112 : (ii+1)*112]); err != nil {
				return err
			}
		}
	}

	// Field (9) 'ExecutionPayload'
	{
		buf = tail[o9:]
		if b.ExecutionPayload == nil {
			b.ExecutionPayload = new(v11.ExecutionPayload)
		}
		if err = b.ExecutionPayload.UnmarshalSSZ(buf); err != nil {
			return err
		}
	}
	return err
}

// SizeSSZ returns the ssz encoded size in bytes for the BeaconBlockBodyBellatrix object
func (b *BeaconBlockBodyBellatrix) SizeSSZ() (size int) {
	size = 316

	// Field (1) 'Eth1Data'
	if b.Eth1Data == nil {
		b.Eth1Data = new(v1.Eth1Data)
	}
	size += b.Eth1Data.SizeSSZ()

	// Field (3) 'ProposerSlashings'
	size += len(b.ProposerSlashings) * 416

	// Field (4) 'AttesterSlashings'
	for ii := 0; ii < len(b.AttesterSlashings); ii++ {
		size += 4
		size += b.AttesterSlashings[ii].SizeSSZ()
	}

	// Field (5) 'Attestations'
	for ii := 0; ii < len(b.Attestations); ii++ {
		size += 4
		size += b.Attestations[ii].SizeSSZ()
	}

	// Field (6) 'Deposits'
	size += len(b.Deposits) * 1240

	// Field (7) 'VoluntaryExits'
	size += len(b.VoluntaryExits) * 112

	// Field (9) 'ExecutionPayload'
	if b.ExecutionPayload == nil {
		b.ExecutionPayload = new(v11.ExecutionPayload)
	}
	size += b.ExecutionPayload.SizeSSZ()

	return
}

// HashTreeRoot ssz hashes the BeaconBlockBodyBellatrix object
func (b *BeaconBlockBodyBellatrix) HashTreeRoot() ([32]byte, error) {
	return ssz.HashWithDefaultHasher(b)
}

// HashTreeRootWith ssz hashes the BeaconBlockBodyBellatrix object with a hasher
func (b *BeaconBlockBodyBellatrix) HashTreeRootWith(hh *ssz.Hasher) (err error) {
	indx := hh.Index()

	// Field (0) 'RandaoReveal'
	if len(b.RandaoReveal) != 96 {
		err = ssz.ErrBytesLength
		return
	}
	hh.PutBytes(b.RandaoReveal)

	// Field (1) 'Eth1Data'
	if err = b.Eth1Data.HashTreeRootWith(hh); err != nil {
		return
	}

	// Field (2) 'Graffiti'
	if len(b.Graffiti) != 32 {
		err = ssz.ErrBytesLength
		return
	}
	hh.PutBytes(b.Graffiti)

	// Field (3) 'ProposerSlashings'
	{
		subIndx := hh.Index()
		num := uint64(len(b.ProposerSlashings))
		if num > 16 {
			err = ssz.ErrIncorrectListSize
			return
		}
		for _, elem := range b.ProposerSlashings {
			if err = elem.HashTreeRootWith(hh); err != nil {
				return
			}
		}
		hh.MerkleizeWithMixin(subIndx, num, 16)
	}

	// Field (4) 'AttesterSlashings'
	{
		subIndx := hh.Index()
		num := uint64(len(b.AttesterSlashings))
		if num > 2 {
			err = ssz.ErrIncorrectListSize
			return
		}
		for _, elem := range b.AttesterSlashings {
			if err = elem.HashTreeRootWith(hh); err != nil {
				return
			}
		}
		hh.MerkleizeWithMixin(subIndx, num, 2)
	}

	// Field (5) 'Attestations'
	{
		subIndx := hh.Index()
		num := uint64(len(b.Attestations))
		if num > 128 {
			err = ssz.ErrIncorrectListSize
			return
		}
		for _, elem := range b.Attestations {
			if err = elem.HashTreeRootWith(hh); err != nil {
				return
			}
		}
		hh.MerkleizeWithMixin(subIndx, num, 128)
	}

	// Field (6) 'Deposits'
	{
		subIndx := hh.Index()
		num := uint64(len(b.Deposits))
		if num > 16 {
			err = ssz.ErrIncorrectListSize
			return
		}
		for _, elem := range b.Deposits {
			if err = elem.HashTreeRootWith(hh); err != nil {
				return
			}
		}
		hh.MerkleizeWithMixin(subIndx, num, 16)
	}

	// Field (7) 'VoluntaryExits'
	{
		subIndx := hh.Index()
		num := uint64(len(b.VoluntaryExits))
		if num > 16 {
			err = ssz.ErrIncorrectListSize
			return
		}
		for _, elem := range b.VoluntaryExits {
			if err = elem.HashTreeRootWith(hh); err != nil {
				return
			}
		}
		hh.MerkleizeWithMixin(subIndx, num, 16)
	}

	// Field (8) 'SyncAggregate'
	if err = b.SyncAggregate.HashTreeRootWith(hh); err != nil {
		return
	}

	// Field (9) 'ExecutionPayload'
	if err = b.ExecutionPayload.HashTreeRootWith(hh); err != nil {
		return
	}

	hh.Merkleize(indx)
	return
}

// MarshalSSZ ssz marshals the BeaconBlockBodyAltair object
func (b *BeaconBlockBodyAltair) MarshalSSZ() ([]byte, error) {
	return ssz.MarshalSSZ(b)
}

// MarshalSSZTo ssz marshals the BeaconBlockBodyAltair object to a target array
func (b *BeaconBlockBodyAltair) MarshalSSZTo(buf []byte) (dst []byte, err error) {
	dst = buf
	offset := int(312)

	// Field (0) 'RandaoReveal'
	if len(b.RandaoReveal) != 96 {
		err = ssz.ErrBytesLength
		return
	}
	dst = append(dst, b.RandaoReveal...)

	// Offset (1) 'Eth1Data'
	dst = ssz.WriteOffset(dst, offset)
	if b.Eth1Data == nil {
		b.Eth1Data = new(v1.Eth1Data)
	}
	offset += b.Eth1Data.SizeSSZ()

	// Field (2) 'Graffiti'
	if len(b.Graffiti) != 32 {
		err = ssz.ErrBytesLength
		return
	}
	dst = append(dst, b.Graffiti...)

	// Offset (3) 'ProposerSlashings'
	dst = ssz.WriteOffset(dst, offset)
	offset += len(b.ProposerSlashings) * 416

	// Offset (4) 'AttesterSlashings'
	dst = ssz.WriteOffset(dst, offset)
	for ii := 0; ii < len(b.AttesterSlashings); ii++ {
		offset += 4
		offset += b.AttesterSlashings[ii].SizeSSZ()
	}

	// Offset (5) 'Attestations'
	dst = ssz.WriteOffset(dst, offset)
	for ii := 0; ii < len(b.Attestations); ii++ {
		offset += 4
		offset += b.Attestations[ii].SizeSSZ()
	}

	// Offset (6) 'Deposits'
	dst = ssz.WriteOffset(dst, offset)
	offset += len(b.Deposits) * 1240

	// Offset (7) 'VoluntaryExits'
	dst = ssz.WriteOffset(dst, offset)
	offset += len(b.VoluntaryExits) * 112

	// Field (8) 'SyncAggregate'
	if b.SyncAggregate == nil {
		b.SyncAggregate = new(v1.SyncAggregate)
	}
	if dst, err = b.SyncAggregate.MarshalSSZTo(dst); err != nil {
		return
	}

	// Field (1) 'Eth1Data'
	if dst, err = b.Eth1Data.MarshalSSZTo(dst); err != nil {
		return
	}

	// Field (3) 'ProposerSlashings'
	if len(b.ProposerSlashings) > 16 {
		err = ssz.ErrListTooBig
		return
	}
	for ii := 0; ii < len(b.ProposerSlashings); ii++ {
		if dst, err = b.ProposerSlashings[ii].MarshalSSZTo(dst); err != nil {
			return
		}
	}

	// Field (4) 'AttesterSlashings'
	if len(b.AttesterSlashings) > 2 {
		err = ssz.ErrListTooBig
		return
	}
	{
		offset = 4 * len(b.AttesterSlashings)
		for ii := 0; ii < len(b.AttesterSlashings); ii++ {
			dst = ssz.WriteOffset(dst, offset)
			offset += b.AttesterSlashings[ii].SizeSSZ()
		}
	}
	for ii := 0; ii < len(b.AttesterSlashings); ii++ {
		if dst, err = b.AttesterSlashings[ii].MarshalSSZTo(dst); err != nil {
			return
		}
	}

	// Field (5) 'Attestations'
	if len(b.Attestations) > 128 {
		err = ssz.ErrListTooBig
		return
	}
	{
		offset = 4 * len(b.Attestations)
		for ii := 0; ii < len(b.Attestations); ii++ {
			dst = ssz.WriteOffset(dst, offset)
			offset += b.Attestations[ii].SizeSSZ()
		}
	}
	for ii := 0; ii < len(b.Attestations); ii++ {
		if dst, err = b.Attestations[ii].MarshalSSZTo(dst); err != nil {
			return
		}
	}

	// Field (6) 'Deposits'
	if len(b.Deposits) > 16 {
		err = ssz.ErrListTooBig
		return
	}
	for ii := 0; ii < len(b.Deposits); ii++ {
		if dst, err = b.Deposits[ii].MarshalSSZTo(dst); err != nil {
			return
		}
	}

	// Field (7) 'VoluntaryExits'
	if len(b.VoluntaryExits) > 16 {
		err = ssz.ErrListTooBig
		return
	}
	for ii := 0; ii < len(b.VoluntaryExits); ii++ {
		if dst, err = b.VoluntaryExits[ii].MarshalSSZTo(dst); err != nil {
			return
		}
	}

	return
}

// UnmarshalSSZ ssz unmarshals the BeaconBlockBodyAltair object
func (b *BeaconBlockBodyAltair) UnmarshalSSZ(buf []byte) error {
	var err error
	size := uint64(len(buf))
	if size < 312 {
		return ssz.ErrSize
	}

	tail := buf
	var o1, o3, o4, o5, o6, o7 uint64

	// Field (0) 'RandaoReveal'
	if cap(b.RandaoReveal) == 0 {
		b.RandaoReveal = make([]byte, 0, len(buf[0:96]))
	}
	b.RandaoReveal = append(b.RandaoReveal, buf[0:96]...)

	// Offset (1) 'Eth1Data'
	if o1 = ssz.ReadOffset(buf[96:100]); o1 > size {
		return ssz.ErrOffset
	}

	if o1 < 312 {
		return ssz.ErrInvalidVariableOffset
	}

	// Field (2) 'Graffiti'
	if cap(b.Graffiti) == 0 {
		b.Graffiti = make([]byte, 0, len(buf[100:132]))
	}
	b.Graffiti = append(b.Graffiti, buf[100:132]...)

	// Offset (3) 'ProposerSlashings'
	if o3 = ssz.ReadOffset(buf[132:136]); o3 > size || o1 > o3 {
		return ssz.ErrOffset
	}

	// Offset (4) 'AttesterSlashings'
	if o4 = ssz.ReadOffset(buf[136:140]); o4 > size || o3 > o4 {
		return ssz.ErrOffset
	}

	// Offset (5) 'Attestations'
	if o5 = ssz.ReadOffset(buf[140:144]); o5 > size || o4 > o5 {
		return ssz.ErrOffset
	}

	// Offset (6) 'Deposits'
	if o6 = ssz.ReadOffset(buf[144:148]); o6 > size || o5 > o6 {
		return ssz.ErrOffset
	}

	// Offset (7) 'VoluntaryExits'
	if o7 = ssz.ReadOffset(buf[148:152]); o7 > size || o6 > o7 {
		return ssz.ErrOffset
	}

	// Field (8) 'SyncAggregate'
	if b.SyncAggregate == nil {
		b.SyncAggregate = new(v1.SyncAggregate)
	}
	if err = b.SyncAggregate.UnmarshalSSZ(buf[152:312]); err != nil {
		return err
	}

	// Field (1) 'Eth1Data'
	{
		buf = tail[o1:o3]
		if b.Eth1Data == nil {
			b.Eth1Data = new(v1.Eth1Data)
		}
		if err = b.Eth1Data.UnmarshalSSZ(buf); err != nil {
			return err
		}
	}

	// Field (3) 'ProposerSlashings'
	{
		buf = tail[o3:o4]
		num, err := ssz.DivideInt2(len(buf), 416, 16)
		if err != nil {
			return err
		}
		b.ProposerSlashings = make([]*v1.ProposerSlashing, num)
		for ii := 0; ii < num; ii++ {
			if b.ProposerSlashings[ii] == nil {
				b.ProposerSlashings[ii] = new(v1.ProposerSlashing)
			}
			if err = b.ProposerSlashings[ii].UnmarshalSSZ(buf[ii*416 : (ii+1)*416]); err != nil {
				return err
			}
		}
	}

	// Field (4) 'AttesterSlashings'
	{
		buf = tail[o4:o5]
		num, err := ssz.DecodeDynamicLength(buf, 2)
		if err != nil {
			return err
		}
		b.AttesterSlashings = make([]*v1.AttesterSlashing, num)
		err = ssz.UnmarshalDynamic(buf, num, func(indx int, buf []byte) (err error) {
			if b.AttesterSlashings[indx] == nil {
				b.AttesterSlashings[indx] = new(v1.AttesterSlashing)
			}
			if err = b.AttesterSlashings[indx].UnmarshalSSZ(buf); err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return err
		}
	}

	// Field (5) 'Attestations'
	{
		buf = tail[o5:o6]
		num, err := ssz.DecodeDynamicLength(buf, 128)
		if err != nil {
			return err
		}
		b.Attestations = make([]*v1.Attestation, num)
		err = ssz.UnmarshalDynamic(buf, num, func(indx int, buf []byte) (err error) {
			if b.Attestations[indx] == nil {
				b.Attestations[indx] = new(v1.Attestation)
			}
			if err = b.Attestations[indx].UnmarshalSSZ(buf); err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return err
		}
	}

	// Field (6) 'Deposits'
	{
		buf = tail[o6:o7]
		num, err := ssz.DivideInt2(len(buf), 1240, 16)
		if err != nil {
			return err
		}
		b.Deposits = make([]*v1.Deposit, num)
		for ii := 0; ii < num; ii++ {
			if b.Deposits[ii] == nil {
				b.Deposits[ii] = new(v1.Deposit)
			}
			if err = b.Deposits[ii].UnmarshalSSZ(buf[ii*1240 : (ii+1)*1240]); err != nil {
				return err
			}
		}
	}

	// Field (7) 'VoluntaryExits'
	{
		buf = tail[o7:]
		num, err := ssz.DivideInt2(len(buf), 112, 16)
		if err != nil {
			return err
		}
		b.VoluntaryExits = make([]*v1.SignedVoluntaryExit, num)
		for ii := 0; ii < num; ii++ {
			if b.VoluntaryExits[ii] == nil {
				b.VoluntaryExits[ii] = new(v1.SignedVoluntaryExit)
			}
			if err = b.VoluntaryExits[ii].UnmarshalSSZ(buf[ii*112 : (ii+1)*112]); err != nil {
				return err
			}
		}
	}
	return err
}

// SizeSSZ returns the ssz encoded size in bytes for the BeaconBlockBodyAltair object
func (b *BeaconBlockBodyAltair) SizeSSZ() (size int) {
	size = 312

	// Field (1) 'Eth1Data'
	if b.Eth1Data == nil {
		b.Eth1Data = new(v1.Eth1Data)
	}
	size += b.Eth1Data.SizeSSZ()

	// Field (3) 'ProposerSlashings'
	size += len(b.ProposerSlashings) * 416

	// Field (4) 'AttesterSlashings'
	for ii := 0; ii < len(b.AttesterSlashings); ii++ {
		size += 4
		size += b.AttesterSlashings[ii].SizeSSZ()
	}

	// Field (5) 'Attestations'
	for ii := 0; ii < len(b.Attestations); ii++ {
		size += 4
		size += b.Attestations[ii].SizeSSZ()
	}

	// Field (6) 'Deposits'
	size += len(b.Deposits) * 1240

	// Field (7) 'VoluntaryExits'
	size += len(b.VoluntaryExits) * 112

	return
}

// HashTreeRoot ssz hashes the BeaconBlockBodyAltair object
func (b *BeaconBlockBodyAltair) HashTreeRoot() ([32]byte, error) {
	return ssz.HashWithDefaultHasher(b)
}

// HashTreeRootWith ssz hashes the BeaconBlockBodyAltair object with a hasher
func (b *BeaconBlockBodyAltair) HashTreeRootWith(hh *ssz.Hasher) (err error) {
	indx := hh.Index()

	// Field (0) 'RandaoReveal'
	if len(b.RandaoReveal) != 96 {
		err = ssz.ErrBytesLength
		return
	}
	hh.PutBytes(b.RandaoReveal)

	// Field (1) 'Eth1Data'
	if err = b.Eth1Data.HashTreeRootWith(hh); err != nil {
		return
	}

	// Field (2) 'Graffiti'
	if len(b.Graffiti) != 32 {
		err = ssz.ErrBytesLength
		return
	}
	hh.PutBytes(b.Graffiti)

	// Field (3) 'ProposerSlashings'
	{
		subIndx := hh.Index()
		num := uint64(len(b.ProposerSlashings))
		if num > 16 {
			err = ssz.ErrIncorrectListSize
			return
		}
		for _, elem := range b.ProposerSlashings {
			if err = elem.HashTreeRootWith(hh); err != nil {
				return
			}
		}
		hh.MerkleizeWithMixin(subIndx, num, 16)
	}

	// Field (4) 'AttesterSlashings'
	{
		subIndx := hh.Index()
		num := uint64(len(b.AttesterSlashings))
		if num > 2 {
			err = ssz.ErrIncorrectListSize
			return
		}
		for _, elem := range b.AttesterSlashings {
			if err = elem.HashTreeRootWith(hh); err != nil {
				return
			}
		}
		hh.MerkleizeWithMixin(subIndx, num, 2)
	}

	// Field (5) 'Attestations'
	{
		subIndx := hh.Index()
		num := uint64(len(b.Attestations))
		if num > 128 {
			err = ssz.ErrIncorrectListSize
			return
		}
		for _, elem := range b.Attestations {
			if err = elem.HashTreeRootWith(hh); err != nil {
				return
			}
		}
		hh.MerkleizeWithMixin(subIndx, num, 128)
	}

	// Field (6) 'Deposits'
	{
		subIndx := hh.Index()
		num := uint64(len(b.Deposits))
		if num > 16 {
			err = ssz.ErrIncorrectListSize
			return
		}
		for _, elem := range b.Deposits {
			if err = elem.HashTreeRootWith(hh); err != nil {
				return
			}
		}
		hh.MerkleizeWithMixin(subIndx, num, 16)
	}

	// Field (7) 'VoluntaryExits'
	{
		subIndx := hh.Index()
		num := uint64(len(b.VoluntaryExits))
		if num > 16 {
			err = ssz.ErrIncorrectListSize
			return
		}
		for _, elem := range b.VoluntaryExits {
			if err = elem.HashTreeRootWith(hh); err != nil {
				return
			}
		}
		hh.MerkleizeWithMixin(subIndx, num, 16)
	}

	// Field (8) 'SyncAggregate'
	if err = b.SyncAggregate.HashTreeRootWith(hh); err != nil {
		return
	}

	hh.Merkleize(indx)
	return
}

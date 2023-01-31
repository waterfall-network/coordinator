package stateutil

import (
	"bytes"
	"encoding/binary"

	"github.com/pkg/errors"
	fieldparams "gitlab.waterfall.network/waterfall/protocol/coordinator/config/fieldparams"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/crypto/hash"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/encoding/bytesutil"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/encoding/ssz"
	ethpb "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1"
)

// BlockVotingDataRootWithHasher returns the hash tree root of input `BlockVoting`.
func BlockVotingDataRootWithHasher(hasher ssz.HashFn, blockVoting *ethpb.BlockVoting) ([32]byte, error) {
	if blockVoting == nil {
		return [32]byte{}, errors.New("nil blockVoting data")
	}
	fixedFldsCount := 2

	attBytes := append([]byte{}, blockVoting.GetCandidates()...)
	for _, att := range blockVoting.GetAttestations() {
		attBytes = []byte(att.String())
	}
	finLen := len(attBytes)
	finChunks := finLen / 32
	if finLen%32 > 0 {
		finChunks++
	}
	fieldRoots := make([][32]byte, fixedFldsCount+finChunks)

	for i := 0; i < len(fieldRoots); i++ {
		fieldRoots[i] = [32]byte{}
	}

	if len(blockVoting.Root) > 0 {
		fieldRoots[0] = bytesutil.ToBytes32(blockVoting.Root)
	}
	slotBuf := make([]byte, 8)
	binary.LittleEndian.PutUint64(slotBuf, uint64(blockVoting.Slot))
	fieldRoots[1] = bytesutil.ToBytes32(slotBuf)

	if finLen > 0 {
		for i := 0; i < finChunks; i++ {
			from := i * 32
			to := from + 32
			if to > finLen {
				to = finLen
			}
			val := attBytes[from:to]
			if len(val) > 0 {
				fieldRoots[i+fixedFldsCount] = bytesutil.ToBytes32(val)
			}
		}
	}

	root, err := ssz.BitwiseMerkleize(hasher, fieldRoots, uint64(len(fieldRoots)), uint64(len(fieldRoots)))
	if err != nil {
		return [32]byte{}, err
	}
	return root, nil
}

// BlockVotingsRoot returns the hash tree root of input `blockVoting`.
func BlockVotingsRoot(blockVotings []*ethpb.BlockVoting) ([32]byte, error) {
	hasher := hash.CustomSHA256Hasher()
	BlockVotingRoots := make([][32]byte, 0, len(blockVotings))
	for i := 0; i < len(blockVotings); i++ {
		root, err := BlockVotingDataRootWithHasher(hasher, blockVotings[i])
		if err != nil {
			return [32]byte{}, errors.Wrap(err, "could not compute blockVoting merkleization")
		}
		BlockVotingRoots = append(BlockVotingRoots, root)
	}

	blockVotingRootsRoot, err := ssz.BitwiseMerkleize(
		hasher,
		BlockVotingRoots,
		uint64(len(BlockVotingRoots)),
		fieldparams.BlockVotingLength,
	)
	if err != nil {
		return [32]byte{}, errors.Wrap(err, "could not compute blockVoting votes merkleization")
	}
	blockVotingRootBuf := new(bytes.Buffer)
	if err := binary.Write(blockVotingRootBuf, binary.LittleEndian, uint64(len(blockVotings))); err != nil {
		return [32]byte{}, errors.Wrap(err, "could not marshal blockVoting votes length")
	}
	// We need to mix in the length of the slice.
	blockVotingRootBufRoot := make([]byte, 32)
	copy(blockVotingRootBufRoot, blockVotingRootBuf.Bytes())
	root := ssz.MixInLength(blockVotingRootsRoot, blockVotingRootBufRoot)

	return root, nil
}

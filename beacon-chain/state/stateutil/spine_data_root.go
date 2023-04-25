package stateutil

import (
	"github.com/pkg/errors"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/encoding/bytesutil"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/encoding/ssz"
	ethpb "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1"
)

// SpineDataRootWithHasher returns the hash tree root of input `spineData`.
func SpineDataRootWithHasher(hasher ssz.HashFn, spineData *ethpb.SpineData) ([32]byte, error) {
	if spineData == nil {
		return [32]byte{}, errors.New("nil spine data")
	}

	spSeqRoots := getParentSpinesRoots(spineData.ParentSpines)

	finLen := len(spineData.Spines) + len(spineData.Prefix) + len(spineData.Finalization)
	finChunks := (finLen / 32)
	if finLen%32 > 0 {
		finChunks++
	}
	fieldRoots := make([][32]byte, finChunks+len(spSeqRoots))
	for i := 0; i < len(fieldRoots); i++ {
		fieldRoots[i] = [32]byte{}
	}

	if finLen > 0 {
		mergedData := make([]byte, 0, finLen)
		mergedData = append(mergedData, spineData.Spines...)
		mergedData = append(mergedData, spineData.Prefix...)
		mergedData = append(mergedData, spineData.Finalization...)
		for i := 0; i < finChunks; i++ {
			from := i * 32
			to := from + 32
			if to > finLen {
				to = finLen
			}
			val := mergedData[from:to]
			if len(val) > 0 {
				fieldRoots[i] = bytesutil.ToBytes32(val)
			}
		}
	}

	for i, root := range spSeqRoots {
		fieldRoots[finChunks+i] = root
	}

	root, err := ssz.BitwiseMerkleize(hasher, fieldRoots, uint64(len(fieldRoots)), uint64(len(fieldRoots)))
	if err != nil {
		return [32]byte{}, err
	}
	return root, nil
}

func getParentSpinesRoots(parentSpines []*ethpb.SpinesSeq) [][32]byte {
	if parentSpines == nil || len(parentSpines) == 0 {
		return [][32]byte{}
	}
	allRoots := make([][32]byte, 0)
	for _, ps := range parentSpines {
		finLen := len(ps.Spines)
		finChunks := finLen / 32
		if finLen%32 > 0 {
			finChunks++
		}
		fieldRoots := make([][32]byte, finChunks)
		for i := 0; i < len(fieldRoots); i++ {
			fieldRoots[i] = [32]byte{}
		}
		if finLen > 0 {
			for i := 0; i < finChunks; i++ {
				from := i * 32
				to := from + 32
				if to > finLen {
					to = finLen
				}
				val := ps.Spines[from:to]
				if len(val) > 0 {
					fieldRoots[i] = bytesutil.ToBytes32(val)
				}
			}
		}
		allRoots = append(allRoots, fieldRoots...)
	}
	return allRoots
}

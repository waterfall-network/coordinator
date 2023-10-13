package stateutil

import (
	"encoding/binary"

	"github.com/pkg/errors"
	fieldparams "gitlab.waterfall.network/waterfall/protocol/coordinator/config/fieldparams"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/encoding/bytesutil"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/encoding/ssz"
	ethpb "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1"
)

// ValidatorRootWithHasher describes a method from which the hash tree root
// of a validator is returned.
func ValidatorRootWithHasher(validator *ethpb.Validator) ([32]byte, error) {
	fieldRoots, err := ValidatorFieldRoots(validator)
	if err != nil {
		return [32]byte{}, err
	}
	return ssz.BitwiseMerkleize(fieldRoots, uint64(len(fieldRoots)), uint64(len(fieldRoots)))
}

// ValidatorFieldRoots describes a method from which the hash tree root
// of a validator is returned.
func ValidatorFieldRoots(validator *ethpb.Validator) ([][32]byte, error) {
	var fieldRoots [][32]byte
	if validator != nil {
		pubkey := bytesutil.ToBytes48(validator.PublicKey)
		creatorAddr := bytesutil.ToBytes32(validator.CreatorAddress)
		withdrawCreds := bytesutil.ToBytes32(validator.WithdrawalCredentials)
		var effectiveBalanceBuf [32]byte
		binary.LittleEndian.PutUint64(effectiveBalanceBuf[:8], validator.EffectiveBalance)
		// Slashed.
		var slashBuf [32]byte
		if validator.Slashed {
			slashBuf[0] = uint8(1)
		} else {
			slashBuf[0] = uint8(0)
		}
		var activationEligibilityBuf [32]byte
		binary.LittleEndian.PutUint64(activationEligibilityBuf[:8], uint64(validator.ActivationEligibilityEpoch))

		var activationBuf [32]byte
		binary.LittleEndian.PutUint64(activationBuf[:8], uint64(validator.ActivationEpoch))

		var exitBuf [32]byte
		binary.LittleEndian.PutUint64(exitBuf[:8], uint64(validator.ExitEpoch))

		var withdrawalBuf [32]byte
		binary.LittleEndian.PutUint64(withdrawalBuf[:8], uint64(validator.WithdrawableEpoch))

		// Public key.
		pubKeyRoot, err := merkleizePubkey(pubkey[:])
		if err != nil {
			return [][32]byte{}, err
		}

		activationHashRoot := bytesutil.ToBytes32(validator.ActivationHash)
		exitHashRoot := bytesutil.ToBytes32(validator.ExitHash)
		withdrawalOpsRoot, err := WithdrawalOpsRoot(validator.WithdrawalOps)
		if err != nil {
			return [][32]byte{}, err
		}

		// main root
		mainRoots := [][32]byte{
			pubKeyRoot,
			creatorAddr,
			withdrawCreds,
			effectiveBalanceBuf,
		}

		mainRootsRoot, err := ssz.BitwiseMerkleize(mainRoots, uint64(len(mainRoots)), uint64(len(mainRoots)))
		if err != nil {
			return [][32]byte{}, errors.Wrap(err, "could not compute validator main merkleization")
		}
		mainLengthRoot := make([]byte, 32)
		binary.LittleEndian.PutUint64(mainLengthRoot, uint64(len(mainRootsRoot)))
		mainRoot := ssz.MixInLength(mainRootsRoot, mainLengthRoot)

		////activation root
		//activationRoots := [][32]byte{
		//	activationEligibilityBuf,
		//	activationBuf,
		//	activationHashRoot,
		//}
		//activationRootsRoot, err := ssz.BitwiseMerkleize(activationRoots, uint64(len(activationRoots)), uint64(len(activationRoots)))
		//if err != nil {
		//	return [][32]byte{}, errors.Wrap(err, "could not compute validator activation merkleization")
		//}
		//activationLengthRoot := make([]byte, 32)
		//binary.LittleEndian.PutUint64(activationLengthRoot, uint64(len(activationRootsRoot)))
		//activationRoot := ssz.MixInLength(activationRootsRoot, activationLengthRoot)

		// exit root
		exitRoots := [][32]byte{
			exitBuf,
			exitHashRoot,
		}
		exitRootsRoot, err := ssz.BitwiseMerkleize(exitRoots, uint64(len(exitRoots)), uint64(len(exitRoots)))
		if err != nil {
			return [][32]byte{}, errors.Wrap(err, "could not compute validator exit merkleization")
		}
		exitLengthRoot := make([]byte, 32)
		binary.LittleEndian.PutUint64(exitLengthRoot, uint64(len(exitRootsRoot)))
		exitRoot := ssz.MixInLength(exitRootsRoot, exitLengthRoot)

		fieldRoots = [][32]byte{
			mainRoot,
			slashBuf,
			activationEligibilityBuf,
			activationBuf,
			activationHashRoot,
			exitRoot,
			withdrawalBuf,
			withdrawalOpsRoot,
		}
	}
	return fieldRoots, nil
}

func WithdrawalOpsRoot(wops []*ethpb.WithdrawalOp) ([32]byte, error) {
	woChunks := PackWithdrawalOpsIntoChunks(wops)

	woRootsRoot, err := ssz.BitwiseMerkleize(woChunks, uint64(len(woChunks)), uint64(len(woChunks)))
	if err != nil {
		return [32]byte{}, errors.Wrap(err, "could not compute withdrawal ops merkleization")
	}
	woLengthRoot := make([]byte, 32)
	binary.LittleEndian.PutUint64(woLengthRoot, uint64(len(wops)))
	return ssz.MixInLength(woRootsRoot, woLengthRoot), nil
}

func PackWithdrawalOpsIntoChunks(wops []*ethpb.WithdrawalOp) [][32]byte {
	rotsByWop := 3
	numOfChunks := len(wops) * rotsByWop
	chunkList := make([][32]byte, numOfChunks)
	for idx, wo := range wops {
		var amtBuff, slotBuff, hashRoot [32]byte
		if wo != nil {
			binary.LittleEndian.PutUint64(amtBuff[:8], uint64(wo.Amount))
			binary.LittleEndian.PutUint64(slotBuff[8:16], uint64(wo.Slot))
			hashRoot = bytesutil.ToBytes32(wo.Hash)
		}
		offset := idx * rotsByWop
		chunkList[offset+0] = amtBuff
		chunkList[offset+1] = slotBuff
		chunkList[offset+2] = hashRoot
	}
	return chunkList
}

// Uint64ListRootWithRegistryLimit computes the HashTreeRoot Merkleization of
// a list of uint64 and mixed with registry limit.
func Uint64ListRootWithRegistryLimit(balances []uint64) ([32]byte, error) {
	balancesChunks, err := PackUint64IntoChunks(balances)
	if err != nil {
		return [32]byte{}, errors.Wrap(err, "could not pack balances into chunks")
	}
	balancesRootsRoot, err := ssz.BitwiseMerkleize(balancesChunks, uint64(len(balancesChunks)), ValidatorLimitForBalancesChunks())
	if err != nil {
		return [32]byte{}, errors.Wrap(err, "could not compute balances merkleization")
	}

	balancesLengthRoot := make([]byte, 32)
	binary.LittleEndian.PutUint64(balancesLengthRoot, uint64(len(balances)))
	return ssz.MixInLength(balancesRootsRoot, balancesLengthRoot), nil
}

// ValidatorLimitForBalancesChunks returns the limit of validators after going through the chunking process.
func ValidatorLimitForBalancesChunks() uint64 {
	maxValidatorLimit := uint64(fieldparams.ValidatorRegistryLimit)
	bytesInUint64 := uint64(8)
	return (maxValidatorLimit*bytesInUint64 + 31) / 32 // round to nearest chunk
}

// PackUint64IntoChunks packs a list of uint64 values into 32 byte roots.
func PackUint64IntoChunks(vals []uint64) ([][32]byte, error) {
	// Initialize how many uint64 values we can pack
	// into a single chunk(32 bytes). Each uint64 value
	// would take up 8 bytes.
	numOfElems := 4
	sizeOfElem := 32 / numOfElems
	// Determine total number of chunks to be
	// allocated to provided list of unsigned
	// 64-bit integers.
	numOfChunks := len(vals) / numOfElems
	// Add an extra chunk if the list size
	// is not a perfect multiple of the number
	// of elements.
	if len(vals)%numOfElems != 0 {
		numOfChunks++
	}
	chunkList := make([][32]byte, numOfChunks)
	for idx, b := range vals {
		// In order to determine how to pack in the uint64 value by index into
		// our chunk list we need to determine a few things.
		// 1) The chunk which the particular uint64 value corresponds to.
		// 2) The position of the value in the chunk itself.
		//
		// Once we have determined these 2 values we can simply find the correct
		// section of contiguous bytes to insert the value in the chunk.
		chunkIdx := idx / numOfElems
		idxInChunk := idx % numOfElems
		chunkPos := idxInChunk * sizeOfElem
		binary.LittleEndian.PutUint64(chunkList[chunkIdx][chunkPos:chunkPos+sizeOfElem], b)
	}
	return chunkList, nil
}

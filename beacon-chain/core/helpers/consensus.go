package helpers

import (
	"bytes"
	"math/big"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/state"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/config/params"
	gwatCommon "gitlab.waterfall.network/waterfall/protocol/gwat/common"
)

var ErrBadUnpublishedChains = errors.New("bad unpublished chains")

type mapPublications map[gwatCommon.Hash]int

// ConsensusUpdateStateSpineFinalization update spine data while checkpoints updated.
func ConsensusUpdateStateSpineFinalization(beaconState state.BeaconState, preJustRoot, preFinRoot []byte) (state.BeaconState, error) {
	finRoot := beaconState.FinalizedCheckpoint().GetRoot()
	justRoot := beaconState.CurrentJustifiedCheckpoint().GetRoot()

	if bytes.Equal(justRoot, preJustRoot) {
		return beaconState, nil
	}
	cpFinalized := beaconState.SpineData().GetCpFinalized()
	finalization := beaconState.SpineData().GetFinalization()
	if bytes.Equal(finRoot, preFinRoot) {
		cpFinalized = append(cpFinalized, finalization...)
		finalization = []byte{}
	} else {
		cpFinalized = append(cpFinalized[len(cpFinalized)-32:], finalization...)
		finalization = []byte{}
	}
	//update state.SpineData
	spineData := beaconState.SpineData()
	spineData.Finalization = finalization
	spineData.CpFinalized = cpFinalized
	err := beaconState.SetSpineData(spineData)

	return beaconState, err
}

// CalculateCandidates candidates sequence from optimistic spines for publication in block.
func CalculateCandidates(parentState state.BeaconState, optSpines []gwatCommon.HashArray) gwatCommon.HashArray {
	//find terminal spine
	var terminalSpine gwatCommon.Hash
	sd := parentState.SpineData()
	// 1. from prefix
	if len(sd.Prefix) > 0 {
		terminalSpine = gwatCommon.BytesToHash(sd.Prefix[len(sd.Prefix)-gwatCommon.HashLength:])
	} else {
		// 2. from finalization or checkpoint finalized spines
		terminalSpine = GetTerminalFinalizedSpine(parentState)
	}
	//calc candidates
	candidates := make(gwatCommon.HashArray, 0, len(optSpines))
	for _, spineList := range optSpines {
		// reset candidates if reach terminal finalized spine
		if spineList.Has(terminalSpine) {
			candidates = make(gwatCommon.HashArray, 0, len(optSpines))
			continue
		}
		if len(spineList) > 0 {
			candidates = append(candidates, spineList[0])
		}
	}
	return candidates
}

// GetTerminalFinalizedSpine retrieve last optimistic finalized spine
func GetTerminalFinalizedSpine(beaconState state.BeaconState) gwatCommon.Hash {
	finalization := beaconState.SpineData().Finalization
	if len(finalization) > 0 {
		return gwatCommon.BytesToHash(finalization[len(finalization)-32:])
	}
	cpFinalized := beaconState.SpineData().CpFinalized
	return gwatCommon.BytesToHash(cpFinalized[len(cpFinalized)-32:])
}

// GetTerminalFinalizedSpine returns finalization spines sequence from state.
func GetFinalizationSequence(beaconState state.BeaconState) gwatCommon.HashArray {
	cpFinalized := gwatCommon.HashArrayFromBytes(beaconState.SpineData().CpFinalized)
	finalization := gwatCommon.HashArrayFromBytes(beaconState.SpineData().Finalization)
	baseSpine := cpFinalized[0]
	finalizationSeq := append(cpFinalized, finalization...)
	if baseIx := finalizationSeq.IndexOf(baseSpine); baseIx > -1 {
		finalizationSeq = finalizationSeq[baseIx+1:]
	}
	return finalizationSeq
}

// GetBaseSpine returns base spine.
func GetBaseSpine(beaconState state.BeaconState) gwatCommon.Hash {
	cpFinalized := gwatCommon.HashArrayFromBytes(beaconState.SpineData().CpFinalized)
	baseSpine := cpFinalized[0]
	return baseSpine
}

// ConsensusCalcPrefix calculates sequence of prefix from array of unpublished spines sequences.
func ConsensusCalcPrefix(unpublishedChains []gwatCommon.HashArray) (gwatCommon.HashArray, error) {
	if err := ConsensusValidateUnpublishedChains(unpublishedChains); err != nil {
		return gwatCommon.HashArray{}, err
	}
	var (
		publicationsMap = mapPublications{}
		commonChain     = gwatCommon.HashArray{}
		prefix          = gwatCommon.HashArray{}
	)
	for i, chain := range unpublishedChains {
		if i == 0 {
			commonChain = chain
		} else {
			commonChain = commonChain.SequenceIntersection(chain)
		}
		for _, spine := range chain {
			publicationsMap[spine]++
		}
	}

	for _, spine := range commonChain {
		if publicationsMap[spine] >= params.BeaconConfig().SpinePublicationsPefixSupport {
			prefix = append(prefix, spine)
		}
	}

	log.WithFields(log.Fields{
		"calcPrefix":        prefix,
		"unpublishedChains": unpublishedChains,
		"commonChain":       commonChain,
		"publicationsMap":   publicationsMap,
	}).Info("Calculate pefix")

	return prefix, nil
}

// ConsensusValidateUnpublishedChains validate unpublished chains
func ConsensusValidateUnpublishedChains(unpublishedChains []gwatCommon.HashArray) error {
	var firstVal gwatCommon.Hash
	for _, chain := range unpublishedChains {
		// empty chains must be removed
		if len(chain) == 0 {
			return errors.Wrap(ErrBadUnpublishedChains, "contains empty chain")
		}
		// chains must be uniq
		if !chain.IsUniq() {
			return errors.Wrap(ErrBadUnpublishedChains, "chain is not uniq")
		}
		//	the first values of each chain must be equal
		if firstVal == (gwatCommon.Hash{}) {
			firstVal = chain[0]
			continue
		}
		if chain[0] != firstVal {
			return errors.Wrap(ErrBadUnpublishedChains, "the first values of chain are not equal")
		}
	}
	return nil
}

func ConsensusCopyUnpublishedChains(unpublishedChains []gwatCommon.HashArray) []gwatCommon.HashArray {
	cpy := make([]gwatCommon.HashArray, len(unpublishedChains))
	for i, chain := range unpublishedChains {
		cpy[i] = chain.Copy()
	}
	return cpy
}

func GweiToBig(gwei uint64) *big.Int {
	return new(big.Int).Mul(new(big.Int).SetUint64(gwei), new(big.Int).SetUint64(1000000000))
}

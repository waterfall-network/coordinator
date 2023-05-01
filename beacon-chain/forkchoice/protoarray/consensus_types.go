package protoarray

import (
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/core/helpers"
	ethpb "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1"
	gwatCommon "gitlab.waterfall.network/waterfall/protocol/gwat/common"
)

type Fork struct {
	roots    [][32]byte
	nodesMap map[[32]byte]*Node
}

// AttestationsData represents data related with attestations in Node.
type AttestationsData struct {
	atts          []*ethpb.Attestation
	justifiedRoot [32]byte
	finalizedRoot [32]byte
}

func (ad *AttestationsData) Attesatations() []*ethpb.Attestation { return ad.atts }
func (ad *AttestationsData) JustifiedRoot() [32]byte             { return ad.justifiedRoot }
func (ad *AttestationsData) FinalizedRoot() [32]byte             { return ad.finalizedRoot }

func NewAttestationsData(
	atts []*ethpb.Attestation,
	justifiedRoot [32]byte,
	finalizedRoot [32]byte,
) AttestationsData {
	return AttestationsData{
		atts:          atts,
		justifiedRoot: justifiedRoot,
		finalizedRoot: finalizedRoot,
	}
}

// SpinesData represents data related with spines in Node.
type SpinesData struct {
	spines      gwatCommon.HashArray   // spines from block.Spines
	prefix      gwatCommon.HashArray   // cache for calculated prefix
	finalized   gwatCommon.HashArray   // finalization sequence block.Finalization
	unpubChains []gwatCommon.HashArray // unpublished chains
}

func (rc *SpinesData) Spines() gwatCommon.HashArray              { return rc.spines.Copy() }
func (rc *SpinesData) Prefix() gwatCommon.HashArray              { return rc.prefix.Copy() }
func (rc *SpinesData) Finalized() gwatCommon.HashArray           { return rc.finalized.Copy() }
func (rc *SpinesData) unpublishedChains() []gwatCommon.HashArray { return rc.unpubChains }

func NewSpinesData(
	parentNode *Node,
	spines gwatCommon.HashArray,
	finalized gwatCommon.HashArray,
) (*SpinesData, error) {
	// calculate new finalization sequence
	parentFin := parentNode.SpinesData().Finalized()
	newFin := append(parentFin, finalized...).Uniq()

	// calculate unpublished spines chains
	parentPrefix := parentNode.SpinesData().Prefix()
	parentUnpubChains := parentNode.SpinesData().unpublishedChains()

	// set new spines to the first position
	unpubChains := []gwatCommon.HashArray{spines}
	for _, chain := range parentUnpubChains {
		chainDif := chain.Difference(parentPrefix)
		// the first spine of dif-chain must be equal to the first spine
		// otherwise - skip
		if chainDif[0] == spines[0] {
			unpubChains = append(unpubChains, chain)
		}
	}

	// calculate the new prefix
	prefix, err := helpers.ConsensusCalcPrefix(unpubChains)
	if err != nil {
		return nil, err
	}
	newPrefix := append(parentPrefix, prefix...).Uniq()
	newPrefix = newPrefix.Difference(newFin)

	return &SpinesData{
		spines:      spines,
		prefix:      newPrefix,
		finalized:   newFin,
		unpubChains: unpubChains,
	}, nil
}

package protoarray

import (
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/core/helpers"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/encoding/bytesutil"
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

func (ad *AttestationsData) Attesatations() []*ethpb.Attestation {
	//return ad.atts
	cpy := make([]*ethpb.Attestation, len(ad.atts))
	for i, v := range ad.atts {
		cpy[i] = &ethpb.Attestation{
			AggregationBits: bytesutil.SafeCopyBytes(v.GetAggregationBits()),
			Data: &ethpb.AttestationData{
				Slot:            v.Data.Slot,
				CommitteeIndex:  v.Data.CommitteeIndex,
				BeaconBlockRoot: bytesutil.SafeCopyBytes(v.Data.BeaconBlockRoot),
				Source: &ethpb.Checkpoint{
					Epoch: v.Data.Source.Epoch,
					Root:  bytesutil.SafeCopyBytes(v.Data.Source.Root),
				},
				Target: &ethpb.Checkpoint{
					Epoch: v.Data.Target.Epoch,
					Root:  bytesutil.SafeCopyBytes(v.Data.Target.Root),
				},
			},
			Signature: nil,
		}
	}
	return cpy
}
func (ad *AttestationsData) JustifiedRoot() [32]byte {
	return bytesutil.ToBytes32(bytesutil.SafeCopyBytes(ad.justifiedRoot[:]))
}
func (ad *AttestationsData) FinalizedRoot() [32]byte {
	return bytesutil.ToBytes32(bytesutil.SafeCopyBytes(ad.finalizedRoot[:]))
}

func (ad *AttestationsData) Copy() *AttestationsData {
	if ad == nil {
		return nil
	}
	return &AttestationsData{
		atts:          ad.Attesatations(),
		justifiedRoot: ad.JustifiedRoot(),
		finalizedRoot: ad.FinalizedRoot(),
	}
}

func NewAttestationsData(
	atts []*ethpb.Attestation,
	justifiedRoot [32]byte,
	finalizedRoot [32]byte,
) *AttestationsData {
	return &AttestationsData{
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

func (rc *SpinesData) Spines() gwatCommon.HashArray    { return rc.spines.Copy() }
func (rc *SpinesData) Prefix() gwatCommon.HashArray    { return rc.prefix.Copy() }
func (rc *SpinesData) Finalized() gwatCommon.HashArray { return rc.finalized.Copy() }
func (rc *SpinesData) unpublishedChains() []gwatCommon.HashArray {
	cpy := make([]gwatCommon.HashArray, len(rc.unpubChains))
	for i, v := range rc.unpubChains {
		cpy[i] = v.Copy()
	}
	return cpy
}
func (rc *SpinesData) Copy() *SpinesData {
	if rc == nil {
		return nil
	}
	return &SpinesData{
		spines:      rc.Spines(),
		prefix:      rc.Prefix(),
		finalized:   rc.Finalized(),
		unpubChains: rc.unpublishedChains(),
	}
}

func NewSpinesData(
	parentNode *Node,
	spines gwatCommon.HashArray,
	finalized gwatCommon.HashArray,
) (*SpinesData, error) {
	if spines == nil {
		spines = gwatCommon.HashArray{}
	}
	if finalized == nil {
		finalized = gwatCommon.HashArray{}
	}
	unpubChains := []gwatCommon.HashArray{}
	if len(spines) > 0 {
		unpubChains = []gwatCommon.HashArray{spines}
	}
	if parentNode == nil || parentNode.spinesData == nil {
		return &SpinesData{
			spines:      spines,
			prefix:      gwatCommon.HashArray{},
			finalized:   finalized,
			unpubChains: unpubChains,
		}, nil
	}

	// calculate new finalization sequence
	parentFin := parentNode.SpinesData().Finalized()
	newFin := append(parentFin, finalized...).Uniq()

	// calculate unpublished spines chains
	parentPrefix := parentNode.SpinesData().Prefix()
	parentUnpubChains := parentNode.SpinesData().unpublishedChains()

	// set new spines to the first position
	for _, chain := range parentUnpubChains {
		if len(chain) == 0 || len(spines) == 0 {
			continue
		}
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

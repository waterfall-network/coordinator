package helpers

import (
	"bytes"
	"fmt"

	"gitlab.waterfall.network/waterfall/protocol/coordinator/encoding/bytesutil"
	ethpb "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1"
	gwatCommon "gitlab.waterfall.network/waterfall/protocol/gwat/common"
)

// CountAttestationsVotes counts votes of BlockVoting
func CountAttestationsVotes(attestations []*ethpb.Attestation) uint64 {
	count := uint64(0)
	for _, att := range attestations {
		if IsAggregated(att) {
			count += att.GetAggregationBits().Count()
		} else {
			count++
		}
	}
	return count
}

func BlockVotingCopy(vote *ethpb.BlockVoting) *ethpb.BlockVoting {
	attestations := make([]*ethpb.Attestation, len(vote.Attestations))
	for i, att := range vote.Attestations {
		attestations[i] = &ethpb.Attestation{
			AggregationBits: att.AggregationBits,
			Data: &ethpb.AttestationData{
				Slot:            att.GetData().GetSlot(),
				CommitteeIndex:  att.GetData().GetCommitteeIndex(),
				BeaconBlockRoot: bytesutil.SafeCopyBytes(att.GetData().BeaconBlockRoot),
				Source: &ethpb.Checkpoint{
					Epoch: att.GetData().GetSource().GetEpoch(),
					Root:  bytesutil.SafeCopyBytes(att.GetData().GetSource().GetRoot()),
				},
				Target: &ethpb.Checkpoint{
					Epoch: att.GetData().GetTarget().GetEpoch(),
					Root:  bytesutil.SafeCopyBytes(att.GetData().GetTarget().GetRoot()),
				},
			},
			Signature: bytesutil.SafeCopyBytes(att.Signature),
		}
	}
	return &ethpb.BlockVoting{
		Root:         bytesutil.SafeCopyBytes(vote.Root),
		Slot:         vote.Slot,
		Candidates:   bytesutil.SafeCopyBytes(vote.Candidates),
		Attestations: attestations,
	}
}

func BlockVotingArrCopy(votes []*ethpb.BlockVoting) []*ethpb.BlockVoting {
	cpy := make([]*ethpb.BlockVoting, len(votes))
	for i, vote := range votes {
		cpy[i] = BlockVotingCopy(vote)
	}
	return cpy
}

// BlockVotingArrStateOrder put BlockVoting array to order to calculate state hash.
func BlockVotingArrStateOrder(votes []*ethpb.BlockVoting) ([]*ethpb.BlockVoting, error) {
	var err error
	cpyAttOrd := make([]*ethpb.BlockVoting, len(votes))
	for i, itm := range votes {
		cpyItm := BlockVotingCopy(itm)
		cpyItm.Attestations, err = AttestationArrSort(cpyItm.Attestations)
		if err != nil {
			return nil, err
		}
		cpyAttOrd[i] = cpyItm
	}
	return BlockVotingArrSort(cpyAttOrd)
}

// BlockVotingArrSort sorts BlockVoting array.
func BlockVotingArrSort(votes []*ethpb.BlockVoting) ([]*ethpb.BlockVoting, error) {
	keys := gwatCommon.HashArray{}
	mapKeyData := map[[32]byte]*ethpb.BlockVoting{}
	for _, itm := range votes {
		k, err := itm.HashTreeRoot()
		if err != nil {
			return nil, err
		}
		mapKeyData[k] = itm
	}
	for k := range mapKeyData {
		keys = append(keys, k)
	}
	keys = keys.Sort()
	sorted := make([]*ethpb.BlockVoting, len(keys))
	for i, k := range keys {
		sorted[i] = mapKeyData[k]
	}
	return sorted, nil
}

// AttestationArrSort sorts attestations array.
func AttestationArrSort(atts []*ethpb.Attestation) ([]*ethpb.Attestation, error) {
	keys := gwatCommon.HashArray{}
	mapKeyData := map[[32]byte]*ethpb.Attestation{}
	for _, itm := range atts {
		k, err := itm.HashTreeRoot()
		if err != nil {
			return nil, err
		}
		mapKeyData[k] = itm
	}
	for k := range mapKeyData {
		keys = append(keys, k)
	}
	keys = keys.Sort()
	sorted := make([]*ethpb.Attestation, len(keys))
	for i, k := range keys {
		sorted[i] = mapKeyData[k]
	}
	return sorted, nil
}

func IndexOfRoot(arrRoots [][]byte, root []byte) int {
	for i, r := range arrRoots {
		if bytes.Equal(r, root) {
			return i
		}
	}
	return -1
}

// PrintBlockVotingArr returns formatted string of BlockVoting array.
func PrintBlockVotingArr(blockVotings []*ethpb.BlockVoting) string {
	str := "["
	for i, bv := range blockVotings {
		str += PrintBlockVoting(bv)
		if i < len(blockVotings)-1 {
			str += ","
		}
	}
	str += "]"
	return str
}

// PrintBlockVoting returns formatted string of BlockVoting.
func PrintBlockVoting(blockVoting *ethpb.BlockVoting) string {
	candStr := "["
	cands := gwatCommon.HashArrayFromBytes(blockVoting.GetCandidates())
	for i, c := range cands {
		candStr += fmt.Sprintf("\"%#x\"", c)
		if i < len(cands)-1 {
			candStr += ","
		}
	}
	candStr += "]"

	str := "{"
	str += fmt.Sprintf("root: \"%#x\",", blockVoting.Root)
	str += fmt.Sprintf("candidates: %s,", candStr)
	str += fmt.Sprintf("slot: %d,", blockVoting.GetSlot())
	str += fmt.Sprintf("attestations: [")
	for i, att := range blockVoting.Attestations {
		str += "{"
		str += fmt.Sprintf("aggrBitLen: %d,", att.GetAggregationBits().Count())
		str += fmt.Sprintf("aggrBitCount: %d,", att.GetAggregationBits().Count())
		str += fmt.Sprintf("slot: %d,", att.Data.GetSlot())
		str += fmt.Sprintf("committeeIndex: %d", att.Data.GetCommitteeIndex())
		//str += fmt.Sprintf("beaconBlockRoot: \"%#x\"", att.Data.GetBeaconBlockRoot())
		str += "}"
		if i < len(blockVoting.Attestations)-1 {
			str += ","
		}
	}
	str += fmt.Sprintf("]")
	str += "}"
	return str
}

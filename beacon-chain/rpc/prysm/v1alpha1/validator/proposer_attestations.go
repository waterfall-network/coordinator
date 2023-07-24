package validator

import (
	"context"
	"sort"

	"github.com/pkg/errors"
	types "github.com/prysmaticlabs/eth2-types"
	"github.com/prysmaticlabs/go-bitfield"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/core/altair"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/core/blocks"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/core/helpers"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/db"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/state"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/config/features"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/config/params"
	ethpb "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1/attestation/aggregation"
	attaggregation "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1/attestation/aggregation/attestations"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/runtime/version"
	gwatCommon "gitlab.waterfall.network/waterfall/protocol/gwat/common"
	"go.opencensus.io/trace"
)

type proposerAtts []*ethpb.Attestation

func (vs *Server) packAttestations(ctx context.Context, latestState state.BeaconState, parentRoot [32]byte) ([]*ethpb.Attestation, error) {
	ctx, span := trace.StartSpan(ctx, "ProposerServer.packAttestations")
	defer span.End()

	atts := vs.AttPool.AggregatedAttestations()
	atts, err := vs.validateAndDeleteAttsInPool(ctx, latestState, atts)
	if err != nil {
		return nil, errors.Wrap(err, "could not filter attestations")
	}

	uAtts, err := vs.AttPool.UnaggregatedAttestations()
	if err != nil {
		return nil, errors.Wrap(err, "could not get unaggregated attestations")
	}
	uAtts, err = vs.validateAndDeleteAttsInPool(ctx, latestState, uAtts)
	if err != nil {
		return nil, errors.Wrap(err, "could not filter attestations")
	}
	atts = append(atts, uAtts...)

	// todo temporary commented for test purposes
	//excludedAtts, err := vs.CollectForkExcludedAttestations(ctx, parentRoot)
	//if err != nil {
	//	return nil, errors.Wrap(err, "could not get excluded attestations")
	//}
	//excludedAtts, err = vs.validateAndDeleteAttsInPool(ctx, latestState, excludedAtts)
	//if err != nil {
	//	return nil, errors.Wrap(err, "could not filter excluded attestations")
	//}
	//atts = append(atts, excludedAtts...)

	// Remove duplicates from both aggregated/unaggregated attestations. This
	// prevents inefficient aggregates being created.
	atts, err = proposerAtts(atts).dedup()
	if err != nil {
		return nil, err
	}

	attsByDataRoot := make(map[[32]byte][]*ethpb.Attestation, len(atts))
	for _, att := range atts {
		attDataRoot, err := att.Data.HashTreeRoot()
		if err != nil {
			return nil, err
		}
		attsByDataRoot[attDataRoot] = append(attsByDataRoot[attDataRoot], att)
	}

	attsForInclusion := proposerAtts(make([]*ethpb.Attestation, 0))
	for _, as := range attsByDataRoot {
		as, err := attaggregation.Aggregate(as)
		if err != nil {
			return nil, err
		}
		attsForInclusion = append(attsForInclusion, as...)
	}
	deduped, err := attsForInclusion.dedup()
	if err != nil {
		return nil, err
	}
	sorted, err := deduped.sortByProfitability()
	if err != nil {
		return nil, err
	}
	atts = sorted.limitToMaxAttestations()
	return atts, nil
}

// filter separates attestation list into two groups: valid and invalid attestations.
// The first group passes the all the required checks for attestation to be considered for proposing.
// And attestations from the second group should be deleted.
func (a proposerAtts) filter(ctx context.Context, st state.BeaconState) (proposerAtts, proposerAtts) {
	validAtts := make([]*ethpb.Attestation, 0, len(a))
	invalidAtts := make([]*ethpb.Attestation, 0, len(a))
	var attestationProcessor func(context.Context, state.BeaconState, *ethpb.Attestation) (state.BeaconState, error)

	switch st.Version() {
	case version.Phase0:
		attestationProcessor = blocks.ProcessAttestationNoVerifySignature
	case version.Altair, version.Bellatrix:
		// Use a wrapper here, as go needs strong typing for the function signature.
		attestationProcessor = func(ctx context.Context, st state.BeaconState, attestation *ethpb.Attestation) (state.BeaconState, error) {
			return altair.ProcessAttestationNoVerifySignature(ctx, st, attestation, nil)
		}
	default:
		// Exit early if there is an unknown state type.
		return validAtts, invalidAtts
	}
	for _, att := range a {
		if _, err := attestationProcessor(ctx, st, att); err == nil {
			validAtts = append(validAtts, att)
			continue
		}
		invalidAtts = append(invalidAtts, att)
	}
	return validAtts, invalidAtts
}

// sortByProfitability orders attestations by highest slot and by highest aggregation bit count.
func (a proposerAtts) sortByProfitability() (proposerAtts, error) {
	if len(a) < 2 {
		return a, nil
	}
	if features.Get().ProposerAttsSelectionUsingMaxCover {
		return a.sortByProfitabilityUsingMaxCover()
	}
	sort.Slice(a, func(i, j int) bool {
		if a[i].Data.Slot == a[j].Data.Slot {
			return a[i].AggregationBits.Count() > a[j].AggregationBits.Count()
		}
		return a[i].Data.Slot > a[j].Data.Slot
	})
	return a, nil
}

// sortByProfitabilityUsingMaxCover orders attestations by highest slot and by highest aggregation bit count.
// Duplicate bits are counted only once, using max-cover algorithm.
func (a proposerAtts) sortByProfitabilityUsingMaxCover() (proposerAtts, error) {
	// Separate attestations by slot, as slot number takes higher precedence when sorting.
	var slots []types.Slot
	attsBySlot := map[types.Slot]proposerAtts{}
	for _, att := range a {
		if _, ok := attsBySlot[att.Data.Slot]; !ok {
			slots = append(slots, att.Data.Slot)
		}
		attsBySlot[att.Data.Slot] = append(attsBySlot[att.Data.Slot], att)
	}

	selectAtts := func(atts proposerAtts) (proposerAtts, error) {
		if len(atts) < 2 {
			return atts, nil
		}
		candidates := make([]*bitfield.Bitlist64, len(atts))
		for i := 0; i < len(atts); i++ {
			var err error
			candidates[i], err = atts[i].AggregationBits.ToBitlist64()
			if err != nil {
				return nil, err
			}
		}
		// Add selected candidates on top, those that are not selected - append at bottom.
		selectedKeys, _, err := aggregation.MaxCover(candidates, len(candidates), true /* allowOverlaps */)
		if err == nil {
			// Pick selected attestations first, leftover attestations will be appended at the end.
			// Both lists will be sorted by number of bits set.
			selectedAtts := make(proposerAtts, selectedKeys.Count())
			leftoverAtts := make(proposerAtts, selectedKeys.Not().Count())
			for i, key := range selectedKeys.BitIndices() {
				selectedAtts[i] = atts[key]
			}
			for i, key := range selectedKeys.Not().BitIndices() {
				leftoverAtts[i] = atts[key]
			}
			sort.Slice(selectedAtts, func(i, j int) bool {
				return selectedAtts[i].AggregationBits.Count() > selectedAtts[j].AggregationBits.Count()
			})
			sort.Slice(leftoverAtts, func(i, j int) bool {
				return leftoverAtts[i].AggregationBits.Count() > leftoverAtts[j].AggregationBits.Count()
			})
			return append(selectedAtts, leftoverAtts...), nil
		}
		return atts, nil
	}

	// Select attestations. Slots are sorted from higher to lower at this point. Within slots attestations
	// are sorted to maximize profitability (greedily selected, with previous attestations' bits
	// evaluated before including any new attestation).
	var sortedAtts proposerAtts
	sort.Slice(slots, func(i, j int) bool {
		return slots[i] > slots[j]
	})
	for _, slot := range slots {
		selected, err := selectAtts(attsBySlot[slot])
		if err != nil {
			return nil, err
		}
		sortedAtts = append(sortedAtts, selected...)
	}

	return sortedAtts, nil
}

// limitToMaxAttestations limits attestations to maximum attestations per block.
func (a proposerAtts) limitToMaxAttestations() proposerAtts {
	if uint64(len(a)) > params.BeaconConfig().MaxAttestations {
		return a[:params.BeaconConfig().MaxAttestations]
	}
	return a
}

// dedup removes duplicate attestations (ones with the same bits set on).
// Important: not only exact duplicates are removed, but proper subsets are removed too
// (their known bits are redundant and are already contained in their supersets).
func (a proposerAtts) dedup() (proposerAtts, error) {
	if len(a) < 2 {
		return a, nil
	}
	attsByDataRoot := make(map[[32]byte][]*ethpb.Attestation, len(a))
	for _, att := range a {
		attDataRoot, err := att.Data.HashTreeRoot()
		if err != nil {
			continue
		}
		attsByDataRoot[attDataRoot] = append(attsByDataRoot[attDataRoot], att)
	}

	uniqAtts := make([]*ethpb.Attestation, 0, len(a))
	for _, atts := range attsByDataRoot {
		for i := 0; i < len(atts); i++ {
			a := atts[i]
			for j := i + 1; j < len(atts); j++ {
				b := atts[j]
				if c, err := a.AggregationBits.Contains(b.AggregationBits); err != nil {
					return nil, err
				} else if c {
					// a contains b, b is redundant.
					atts[j] = atts[len(atts)-1]
					atts[len(atts)-1] = nil
					atts = atts[:len(atts)-1]
					j--
				} else if c, err := b.AggregationBits.Contains(a.AggregationBits); err != nil {
					return nil, err
				} else if c {
					// b contains a, a is redundant.
					atts[i] = atts[len(atts)-1]
					atts[len(atts)-1] = nil
					atts = atts[:len(atts)-1]
					i--
					break
				}
			}
		}
		uniqAtts = append(uniqAtts, atts...)
	}

	return uniqAtts, nil
}

// This filters the input attestations to return a list of valid attestations to be packaged inside a beacon block.
func (vs *Server) validateAndDeleteAttsInPool(ctx context.Context, st state.BeaconState, atts []*ethpb.Attestation) ([]*ethpb.Attestation, error) {
	ctx, span := trace.StartSpan(ctx, "ProposerServer.validateAndDeleteAttsInPool")
	defer span.End()

	ctx = context.WithValue(ctx, params.BeaconConfig().CtxBlockFetcherKey, db.BlockInfoFetcherFunc(vs.BeaconDB))
	validAtts, invalidAtts := proposerAtts(atts).filter(ctx, st)
	if err := vs.deleteAttsInPool(ctx, invalidAtts); err != nil {
		return nil, err
	}
	return validAtts, nil
}

// The input attestations are processed and seen by the node, this deletes them from pool
// so proposers don't include them in a block for the future.
func (vs *Server) deleteAttsInPool(ctx context.Context, atts []*ethpb.Attestation) error {
	ctx, span := trace.StartSpan(ctx, "ProposerServer.deleteAttsInPool")
	defer span.End()

	for _, att := range atts {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if helpers.IsAggregated(att) {
			if err := vs.AttPool.DeleteAggregatedAttestation(att); err != nil {
				return err
			}
		} else {
			if err := vs.AttPool.DeleteUnaggregatedAttestation(att); err != nil {
				return err
			}
		}
	}
	return nil
}

// CollectForkExcludedAttestations collect attestations not included into canonical chain
func (vs *Server) CollectForkExcludedAttestations(ctx context.Context, parentRoot [32]byte) ([]*ethpb.Attestation, error) {
	if parentRoot == params.BeaconConfig().ZeroHash {
		return nil, nil
	}
	leaf := gwatCommon.BytesToHash(parentRoot[:])
	// collect blocks excluded from canonical chain
	exRoots := vs.HeadFetcher.ForkChoicer().CollectForkExcludedBlkRoots(leaf)
	//collect attestation
	atts := []*ethpb.Attestation{}
	for _, r := range exRoots {
		blk, err := vs.BeaconDB.Block(ctx, r)
		if err != nil {
			return nil, err
		}
		attestations := blk.Block().Body().Attestations()
		if len(attestations) > 0 {
			atts = append(atts, attestations...)
		}
	}
	return atts, nil
}

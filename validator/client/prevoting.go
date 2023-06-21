package client

import (
	"context"
	"fmt"
	"strings"

	types "github.com/prysmaticlabs/eth2-types"
	"github.com/prysmaticlabs/go-bitfield"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/async"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/core/signing"
	fieldparams "gitlab.waterfall.network/waterfall/protocol/coordinator/config/fieldparams"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/config/params"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/encoding/bytesutil"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/monitoring/tracing"
	ethpb "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1"
	validatorpb "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1/validator-client"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/time/slots"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/validator/client/iface"
	"go.opencensus.io/trace"
)

// SubmitPrevote completes the validator client's prevote responsibility at a given slot.
// It fetches the latest beacon block head along with the latest canonical beacon state
// information in order to sign the block and include information about the validator's
// participation in voting on the block.
func (v *validator) SubmitPrevote(ctx context.Context, slot types.Slot, pubKey [fieldparams.BLSPubkeyLength]byte) {
	ctx, span := trace.StartSpan(ctx, "validator.SubmitPrevote")
	defer span.End()
	span.AddAttributes(trace.StringAttribute("validator", fmt.Sprintf("%#x", pubKey)))

	var b strings.Builder
	if err := b.WriteByte(byte(iface.RoleAttester)); err != nil {
		log.WithError(err).Error("Could not write role byte for lock key while prevote")
		tracing.AnnotateError(span, err)
		return
	}
	_, err := b.Write(pubKey[:])
	if err != nil {
		log.WithError(err).Error("Could not write pubkey bytes for lock key while prevote")
		tracing.AnnotateError(span, err)
		return
	}
	lock := async.NewMultilock(b.String())
	lock.Lock()
	defer lock.Unlock()

	fmtKey := fmt.Sprintf("%#x", pubKey[:])
	log := log.WithField("pubKey", fmt.Sprintf("%#x", bytesutil.Trunc(pubKey[:]))).WithField("slot", slot)
	duty, err := v.duty(pubKey)
	if err != nil {
		log.WithError(err).Error("Could not fetch validator assignment while prevote")
		if v.emitAccountMetrics {
			ValidatorAttestFailVec.WithLabelValues(fmtKey).Inc()
		}
		tracing.AnnotateError(span, err)
		return
	}
	if len(duty.Committee) == 0 {
		log.Warnf("Empty committee for validator prevote")
		return
	}

	req := &ethpb.PreVoteRequest{
		Slot:           slot,
		CommitteeIndex: duty.CommitteeIndex,
	}
	data, err := v.validatorClient.GetPrevoteData(ctx, req)
	if err != nil {
		log.WithError(err).Error("Could not request prevote data to sign at slot")
		if v.emitAccountMetrics {
			ValidatorAttestFailVec.WithLabelValues(fmtKey).Inc()
		}
		tracing.AnnotateError(span, err)
		return
	}

	indexedPrevote := &ethpb.IndexedPreVote{
		AttestingIndices: []uint64{uint64(duty.ValidatorIndex)},
		Data:             data,
	}

	domain, signingRoot, err := v.getDomainAndSigningRootPrevote(ctx, indexedPrevote.Data)
	if err != nil {
		log.WithError(err).Error("Could not get domain and signing root from prevote")
		if v.emitAccountMetrics {
			ValidatorAttestFailVec.WithLabelValues(fmtKey).Inc()
		}
		tracing.AnnotateError(span, err)
		return
	}

	sig, _, err := v.signPrevote(ctx, pubKey, data, slot, domain, signingRoot)
	if err != nil {
		log.WithError(err).Error("Could not sign prevote")
		if v.emitAccountMetrics {
			ValidatorAttestFailVec.WithLabelValues(fmtKey).Inc()
		}
		tracing.AnnotateError(span, err)
		return
	}

	var indexInCommittee uint64
	var found bool
	for i, vID := range duty.Committee {
		if vID == duty.ValidatorIndex {
			indexInCommittee = uint64(i)
			found = true
			break
		}
	}
	if !found {
		log.Errorf("Validator ID %d not found in committee of %v", duty.ValidatorIndex, duty.Committee)
		if v.emitAccountMetrics {
			ValidatorAttestFailVec.WithLabelValues(fmtKey).Inc()
		}
		return
	}

	aggregationBitfield := bitfield.NewBitlist(uint64(len(duty.Committee)))
	aggregationBitfield.SetBitAt(indexInCommittee, true)
	prevote := &ethpb.PreVote{
		Data:            data,
		AggregationBits: aggregationBitfield,
		Signature:       sig,
	}

	// Set the signature of the attestation and send it out to the beacon node.
	indexedPrevote.Signature = sig

	pvResp, err := v.validatorClient.ProposePrevote(ctx, prevote)
	if err != nil {
		log.WithError(err).Error("Could not submit prevote to beacon node")
		if v.emitAccountMetrics {
			ValidatorAttestFailVec.WithLabelValues(fmtKey).Inc()
		}
		tracing.AnnotateError(span, err)
		return
	}

	span.AddAttributes(
		trace.Int64Attribute("slot", int64(slot)), // lint:ignore uintcast -- This conversion is OK for tracing.
		trace.StringAttribute("prevoteHash", fmt.Sprintf("%#x", pvResp.PrevoteDataRoot)),
		trace.Int64Attribute("committeeIndex", int64(data.Index)),
		trace.StringAttribute("bitfield", fmt.Sprintf("%#x", aggregationBitfield)),
	)

	if v.emitAccountMetrics {
		ValidatorAttestSuccessVec.WithLabelValues(fmtKey).Inc()
		ValidatorAttestedSlotsGaugeVec.WithLabelValues(fmtKey).Set(float64(slot))
	}
}

func (v *validator) getDomainAndSigningRootPrevote(ctx context.Context, data *ethpb.PreVoteData) (*ethpb.DomainResponse, [32]byte, error) {
	domain, err := v.domainData(ctx, slots.ToEpoch(data.Slot), params.BeaconConfig().DomainBeaconAttester[:])
	if err != nil {
		return nil, [32]byte{}, err
	}
	root, err := signing.ComputeSigningRoot(data, domain.SignatureDomain)
	if err != nil {
		return nil, [32]byte{}, err
	}
	return domain, root, nil
}

// Given validator's public key, this function returns the signature of a prevote data and its signing root.
func (v *validator) signPrevote(ctx context.Context, pubKey [fieldparams.BLSPubkeyLength]byte, data *ethpb.PreVoteData,
	slot types.Slot, domain *ethpb.DomainResponse, root [32]byte) ([]byte, [32]byte, error) {
	sig, err := v.keyManager.Sign(ctx, &validatorpb.SignRequest{
		PublicKey:       pubKey[:],
		SigningRoot:     root[:],
		SignatureDomain: domain.SignatureDomain,
		Object:          &validatorpb.SignRequest_PrevoteData{PrevoteData: data},
		SigningSlot:     slot,
	})
	if err != nil {
		return nil, [32]byte{}, err
	}

	return sig.Marshal(), root, nil
}

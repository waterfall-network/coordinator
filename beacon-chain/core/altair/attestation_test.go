package altair_test

import (
	"context"
	"fmt"
	"testing"

	fuzz "github.com/google/gofuzz"
	types "github.com/prysmaticlabs/eth2-types"
	"github.com/prysmaticlabs/go-bitfield"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/core/altair"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/core/helpers"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/core/signing"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/core/time"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/state"
	stateAltair "gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/state/v2"
	fieldparams "gitlab.waterfall.network/waterfall/protocol/coordinator/config/fieldparams"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/config/params"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/crypto/bls"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/math"
	ethpb "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1/attestation"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1/wrapper"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/require"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/util"
)

func TestProcessAttestations_InclusionDelayFailure(t *testing.T) {
	attestations := []*ethpb.Attestation{
		util.HydrateAttestation(&ethpb.Attestation{
			Data: &ethpb.AttestationData{
				Target: &ethpb.Checkpoint{Epoch: 0, Root: make([]byte, fieldparams.RootLength)},
				Slot:   5,
			},
		}),
	}
	b := util.NewBeaconBlockAltair()
	b.Block = &ethpb.BeaconBlockAltair{
		Body: &ethpb.BeaconBlockBodyAltair{
			Attestations: attestations,
		},
	}
	beaconState, _ := util.DeterministicGenesisStateAltair(t, 100)

	want := fmt.Sprintf(
		"attestation slot %d + inclusion delay %d > state slot %d",
		attestations[0].Data.Slot,
		params.BeaconConfig().MinAttestationInclusionDelay,
		beaconState.Slot(),
	)
	wsb, err := wrapper.WrappedSignedBeaconBlock(b)
	require.NoError(t, err)
	_, err = altair.ProcessAttestationsNoVerifySignature(context.Background(), beaconState, wsb)
	require.ErrorContains(t, want, err)
}

func TestProcessAttestations_NeitherCurrentNorPrevEpoch(t *testing.T) {
	att := util.HydrateAttestation(&ethpb.Attestation{
		Data: &ethpb.AttestationData{
			Source: &ethpb.Checkpoint{Epoch: 0, Root: []byte("hello-world")},
			Target: &ethpb.Checkpoint{Epoch: 0}}})

	b := util.NewBeaconBlockAltair()
	b.Block = &ethpb.BeaconBlockAltair{
		Body: &ethpb.BeaconBlockBodyAltair{
			Attestations: []*ethpb.Attestation{att},
		},
	}
	beaconState, _ := util.DeterministicGenesisStateAltair(t, 100)
	err := beaconState.SetSlot(beaconState.Slot() + params.BeaconConfig().SlotsPerEpoch*4 + params.BeaconConfig().MinAttestationInclusionDelay)
	require.NoError(t, err)
	pfc := beaconState.PreviousJustifiedCheckpoint()
	pfc.Root = []byte("hello-world")
	require.NoError(t, beaconState.SetPreviousJustifiedCheckpoint(pfc))

	want := fmt.Sprintf(
		"expected target epoch (%d) to be the previous epoch (%d) or the current epoch (%d)",
		att.Data.Target.Epoch,
		time.PrevEpoch(beaconState),
		time.CurrentEpoch(beaconState),
	)
	wsb, err := wrapper.WrappedSignedBeaconBlock(b)
	require.NoError(t, err)
	_, err = altair.ProcessAttestationsNoVerifySignature(context.Background(), beaconState, wsb)
	require.ErrorContains(t, want, err)
}

func TestProcessAttestations_CurrentEpochFFGDataMismatches(t *testing.T) {
	attestations := []*ethpb.Attestation{
		{
			Data: &ethpb.AttestationData{
				Target: &ethpb.Checkpoint{Epoch: 0, Root: make([]byte, fieldparams.RootLength)},
				Source: &ethpb.Checkpoint{Epoch: 1, Root: make([]byte, fieldparams.RootLength)},
			},
			AggregationBits: bitfield.Bitlist{0x09},
		},
	}
	b := util.NewBeaconBlockAltair()
	b.Block = &ethpb.BeaconBlockAltair{
		Body: &ethpb.BeaconBlockBodyAltair{
			Attestations: attestations,
		},
	}
	beaconState, _ := util.DeterministicGenesisStateAltair(t, 100)
	require.NoError(t, beaconState.SetSlot(beaconState.Slot()+params.BeaconConfig().MinAttestationInclusionDelay))
	cfc := beaconState.CurrentJustifiedCheckpoint()
	cfc.Root = []byte("hello-world")
	require.NoError(t, beaconState.SetCurrentJustifiedCheckpoint(cfc))

	want := "source check point not equal to current justified checkpoint"
	wsb, err := wrapper.WrappedSignedBeaconBlock(b)
	require.NoError(t, err)
	_, err = altair.ProcessAttestationsNoVerifySignature(context.Background(), beaconState, wsb)
	require.ErrorContains(t, want, err)
	b.Block.Body.Attestations[0].Data.Source.Epoch = time.CurrentEpoch(beaconState)
	b.Block.Body.Attestations[0].Data.Source.Root = []byte{}
	wsb, err = wrapper.WrappedSignedBeaconBlock(b)
	require.NoError(t, err)
	_, err = altair.ProcessAttestationsNoVerifySignature(context.Background(), beaconState, wsb)
	require.ErrorContains(t, want, err)
}

func TestProcessAttestations_PrevEpochFFGDataMismatches(t *testing.T) {
	beaconState, _ := util.DeterministicGenesisStateAltair(t, 100)

	aggBits := bitfield.NewBitlist(3)
	aggBits.SetBitAt(0, true)
	attestations := []*ethpb.Attestation{
		{
			Data: &ethpb.AttestationData{
				Source: &ethpb.Checkpoint{Epoch: 1, Root: make([]byte, fieldparams.RootLength)},
				Target: &ethpb.Checkpoint{Epoch: 1, Root: make([]byte, fieldparams.RootLength)},
				Slot:   params.BeaconConfig().SlotsPerEpoch,
			},
			AggregationBits: aggBits,
		},
	}
	b := util.NewBeaconBlockAltair()
	b.Block = &ethpb.BeaconBlockAltair{
		Body: &ethpb.BeaconBlockBodyAltair{
			Attestations: attestations,
		},
	}

	err := beaconState.SetSlot(beaconState.Slot() + 2*params.BeaconConfig().SlotsPerEpoch)
	require.NoError(t, err)
	pfc := beaconState.PreviousJustifiedCheckpoint()
	pfc.Root = []byte("hello-world")
	require.NoError(t, beaconState.SetPreviousJustifiedCheckpoint(pfc))

	want := "source check point not equal to previous justified checkpoint"
	wsb, err := wrapper.WrappedSignedBeaconBlock(b)
	require.NoError(t, err)
	_, err = altair.ProcessAttestationsNoVerifySignature(context.Background(), beaconState, wsb)
	require.ErrorContains(t, want, err)
	b.Block.Body.Attestations[0].Data.Source.Epoch = time.PrevEpoch(beaconState)
	b.Block.Body.Attestations[0].Data.Target.Epoch = time.PrevEpoch(beaconState)
	b.Block.Body.Attestations[0].Data.Source.Root = []byte{}
	wsb, err = wrapper.WrappedSignedBeaconBlock(b)
	require.NoError(t, err)
	_, err = altair.ProcessAttestationsNoVerifySignature(context.Background(), beaconState, wsb)
	require.ErrorContains(t, want, err)
}

func TestProcessAttestations_InvalidAggregationBitsLength(t *testing.T) {
	beaconState, _ := util.DeterministicGenesisStateAltair(t, 100)

	aggBits := bitfield.NewBitlist(4)
	att := &ethpb.Attestation{
		Data: &ethpb.AttestationData{
			Source: &ethpb.Checkpoint{Epoch: 0, Root: []byte("hello-world")},
			Target: &ethpb.Checkpoint{Epoch: 0}},
		AggregationBits: aggBits,
	}

	b := util.NewBeaconBlockAltair()
	b.Block = &ethpb.BeaconBlockAltair{
		Body: &ethpb.BeaconBlockBodyAltair{
			Attestations: []*ethpb.Attestation{att},
		},
	}

	err := beaconState.SetSlot(beaconState.Slot() + params.BeaconConfig().MinAttestationInclusionDelay)
	require.NoError(t, err)

	cfc := beaconState.CurrentJustifiedCheckpoint()
	cfc.Root = []byte("hello-world")
	require.NoError(t, beaconState.SetCurrentJustifiedCheckpoint(cfc))

	expected := "failed to verify aggregation bitfield: wanted participants bitfield length 3, got: 4"
	wsb, err := wrapper.WrappedSignedBeaconBlock(b)
	require.NoError(t, err)
	_, err = altair.ProcessAttestationsNoVerifySignature(context.Background(), beaconState, wsb)
	require.ErrorContains(t, expected, err)
}

func TestProcessAttestations_OK(t *testing.T) {
	t.Skip()
	beaconState, privKeys := util.DeterministicGenesisStateAltair(t, 100)

	aggBits := bitfield.NewBitlist(3)
	aggBits.SetBitAt(0, true)
	var mockRoot [32]byte
	copy(mockRoot[:], "hello-world")
	att := util.HydrateAttestation(&ethpb.Attestation{
		Data: &ethpb.AttestationData{
			Source: &ethpb.Checkpoint{Root: mockRoot[:]},
			Target: &ethpb.Checkpoint{Root: mockRoot[:]},
		},
		AggregationBits: aggBits,
	})

	cfc := beaconState.CurrentJustifiedCheckpoint()
	cfc.Root = mockRoot[:]
	require.NoError(t, beaconState.SetCurrentJustifiedCheckpoint(cfc))

	committee, err := helpers.BeaconCommitteeFromState(context.Background(), beaconState, att.Data.Slot, att.Data.CommitteeIndex)
	require.NoError(t, err)
	attestingIndices, err := attestation.AttestingIndices(att.AggregationBits, committee)
	require.NoError(t, err)
	sigs := make([]bls.Signature, len(attestingIndices))
	for i, indice := range attestingIndices {
		sb, err := signing.ComputeDomainAndSign(beaconState, 0, att.Data, params.BeaconConfig().DomainBeaconAttester, privKeys[indice])
		require.NoError(t, err)
		sig, err := bls.SignatureFromBytes(sb)
		require.NoError(t, err)
		sigs[i] = sig
	}
	att.Signature = bls.AggregateSignatures(sigs).Marshal()

	block := util.NewBeaconBlockAltair()
	block.Block.Body.Attestations = []*ethpb.Attestation{att}

	err = beaconState.SetSlot(beaconState.Slot() + params.BeaconConfig().MinAttestationInclusionDelay)
	require.NoError(t, err)
	wsb, err := wrapper.WrappedSignedBeaconBlock(block)
	require.NoError(t, err)
	ctxBlockFetcher := params.CtxBlockFetcher(func(ctx context.Context, blockRoot [32]byte) (types.ValidatorIndex, types.Slot, uint64, error) {
		block := wsb
		votesIncluded := uint64(0)
		for _, att := range block.Block().Body().Attestations() {
			votesIncluded += att.AggregationBits.Count()
		}

		return block.Block().ProposerIndex() - 1, block.Block().Slot() - 1, votesIncluded, nil
	})

	ctxWithFetcher := context.WithValue(context.Background(),
		params.BeaconConfig().CtxBlockFetcherKey,
		ctxBlockFetcher)
	_, err = altair.ProcessAttestationsNoVerifySignature(ctxWithFetcher, beaconState, wsb)
	require.NoError(t, err)
}

func TestProcessAttestationNoVerify_SourceTargetHead(t *testing.T) {
	t.Skip()
	beaconState, _ := util.DeterministicGenesisStateAltair(t, 64)
	err := beaconState.SetSlot(beaconState.Slot() + params.BeaconConfig().MinAttestationInclusionDelay)
	require.NoError(t, err)

	aggBits := bitfield.NewBitlist(2)
	aggBits.SetBitAt(0, true)
	aggBits.SetBitAt(1, true)
	r, err := helpers.BlockRootAtSlot(beaconState, 0)
	require.NoError(t, err)
	att := &ethpb.Attestation{
		Data: &ethpb.AttestationData{
			BeaconBlockRoot: r,
			Source:          &ethpb.Checkpoint{Epoch: 0, Root: make([]byte, fieldparams.RootLength)},
			Target:          &ethpb.Checkpoint{Epoch: 0, Root: make([]byte, fieldparams.RootLength)},
		},
		AggregationBits: aggBits,
	}
	zeroSig := [96]byte{}
	att.Signature = zeroSig[:]

	ckp := beaconState.CurrentJustifiedCheckpoint()
	copy(ckp.Root, make([]byte, fieldparams.RootLength))
	require.NoError(t, beaconState.SetCurrentJustifiedCheckpoint(ckp))

	beaconState, err = altair.ProcessAttestationNoVerifySignature(context.Background(), beaconState, att, nil)
	require.NoError(t, err)

	p, err := beaconState.CurrentEpochParticipation()
	require.NoError(t, err)

	committee, err := helpers.BeaconCommitteeFromState(context.Background(), beaconState, att.Data.Slot, att.Data.CommitteeIndex)
	require.NoError(t, err)
	indices, err := attestation.AttestingIndices(att.AggregationBits, committee)
	require.NoError(t, err)
	for _, index := range indices {
		has, err := altair.HasValidatorFlag(p[index], params.BeaconConfig().TimelyHeadFlagIndex)
		require.NoError(t, err)
		require.Equal(t, true, has)
		has, err = altair.HasValidatorFlag(p[index], params.BeaconConfig().TimelySourceFlagIndex)
		require.NoError(t, err)
		require.Equal(t, true, has)
		has, err = altair.HasValidatorFlag(p[index], params.BeaconConfig().TimelyTargetFlagIndex)
		require.NoError(t, err)
		require.Equal(t, true, has)
	}
}

func TestValidatorFlag_Has(t *testing.T) {
	tests := []struct {
		name     string
		set      uint8
		expected []uint8
	}{
		{name: "none",
			set:      0,
			expected: []uint8{},
		},
		{
			name:     "source",
			set:      1,
			expected: []uint8{params.BeaconConfig().TimelySourceFlagIndex},
		},
		{
			name:     "target",
			set:      2,
			expected: []uint8{params.BeaconConfig().TimelyTargetFlagIndex},
		},
		{
			name:     "head",
			set:      4,
			expected: []uint8{params.BeaconConfig().TimelyHeadFlagIndex},
		},
		{
			name:     "voting",
			set:      8,
			expected: []uint8{params.BeaconConfig().DAGTimelyVotingFlagIndex},
		},
		{
			name:     "source, target",
			set:      3,
			expected: []uint8{params.BeaconConfig().TimelySourceFlagIndex, params.BeaconConfig().TimelyTargetFlagIndex},
		},
		{
			name:     "source, head",
			set:      5,
			expected: []uint8{params.BeaconConfig().TimelySourceFlagIndex, params.BeaconConfig().TimelyHeadFlagIndex},
		},
		{
			name:     "target, head",
			set:      6,
			expected: []uint8{params.BeaconConfig().TimelyTargetFlagIndex, params.BeaconConfig().TimelyTargetFlagIndex},
		},
		{
			name:     "source, target, head",
			set:      7,
			expected: []uint8{params.BeaconConfig().TimelySourceFlagIndex, params.BeaconConfig().TimelyTargetFlagIndex, params.BeaconConfig().TimelyHeadFlagIndex},
		},
		{
			name:     "source, target, head, voting",
			set:      15,
			expected: []uint8{params.BeaconConfig().TimelySourceFlagIndex, params.BeaconConfig().TimelyTargetFlagIndex, params.BeaconConfig().TimelyHeadFlagIndex, params.BeaconConfig().DAGTimelyVotingFlagIndex},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, f := range tt.expected {
				has, err := altair.HasValidatorFlag(tt.set, f)
				require.NoError(t, err)
				require.Equal(t, true, has)
			}
		})
	}
}

func TestValidatorFlag_Has_ExceedsLength(t *testing.T) {
	_, err := altair.HasValidatorFlag(0, 8)
	require.ErrorContains(t, "flag position exceeds length", err)
}

func TestValidatorFlag_Add(t *testing.T) {
	tests := []struct {
		name          string
		set           []uint8
		expectedTrue  []uint8
		expectedFalse []uint8
	}{
		{name: "none",
			set:           []uint8{},
			expectedTrue:  []uint8{},
			expectedFalse: []uint8{params.BeaconConfig().TimelySourceFlagIndex, params.BeaconConfig().TimelyTargetFlagIndex, params.BeaconConfig().TimelyHeadFlagIndex},
		},
		{
			name:          "source",
			set:           []uint8{params.BeaconConfig().TimelySourceFlagIndex},
			expectedTrue:  []uint8{params.BeaconConfig().TimelySourceFlagIndex},
			expectedFalse: []uint8{params.BeaconConfig().TimelyTargetFlagIndex, params.BeaconConfig().TimelyHeadFlagIndex},
		},
		{
			name:          "source, target",
			set:           []uint8{params.BeaconConfig().TimelySourceFlagIndex, params.BeaconConfig().TimelyTargetFlagIndex},
			expectedTrue:  []uint8{params.BeaconConfig().TimelySourceFlagIndex, params.BeaconConfig().TimelyTargetFlagIndex},
			expectedFalse: []uint8{params.BeaconConfig().TimelyHeadFlagIndex},
		},
		{
			name:          "source, target, head",
			set:           []uint8{params.BeaconConfig().TimelySourceFlagIndex, params.BeaconConfig().TimelyTargetFlagIndex, params.BeaconConfig().TimelyHeadFlagIndex},
			expectedTrue:  []uint8{params.BeaconConfig().TimelySourceFlagIndex, params.BeaconConfig().TimelyTargetFlagIndex, params.BeaconConfig().TimelyHeadFlagIndex},
			expectedFalse: []uint8{},
		},
	}
	var err error
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := uint8(0)
			for _, f := range tt.set {
				b, err = altair.AddValidatorFlag(b, f)
				require.NoError(t, err)
			}
			for _, f := range tt.expectedFalse {
				has, err := altair.HasValidatorFlag(b, f)
				require.NoError(t, err)
				require.Equal(t, false, has)
			}
			for _, f := range tt.expectedTrue {
				has, err := altair.HasValidatorFlag(b, f)
				require.NoError(t, err)
				require.Equal(t, true, has)
			}
		})
	}
}

func TestValidatorFlag_Add_ExceedsLength(t *testing.T) {
	_, err := altair.AddValidatorFlag(0, 8)
	require.ErrorContains(t, "flag position exceeds length", err)
}

func TestFuzzProcessAttestationsNoVerify_10000(t *testing.T) {
	t.Skip()
	fuzzer := fuzz.NewWithSeed(0)
	bState := &ethpb.BeaconStateAltair{}
	b := &ethpb.SignedBeaconBlockAltair{Block: &ethpb.BeaconBlockAltair{}}
	for i := 0; i < 10000; i++ {
		fuzzer.Fuzz(bState)
		fuzzer.Fuzz(b)
		if b.Block == nil {
			b.Block = &ethpb.BeaconBlockAltair{}
		}
		for j := 0; j < len(b.Block.Body.Attestations); j++ {
			if b.Block.Body.Attestations[j] == nil {
				b.Block.Body.Attestations = append(b.Block.Body.Attestations[:j], b.Block.Body.Attestations[j+1:]...)
				j = 0
				continue
			}
			if b.Block.Body.Attestations[j].Data == nil {
				b.Block.Body.Attestations[j].Data = &ethpb.AttestationData{
					Slot:            0,
					CommitteeIndex:  0,
					BeaconBlockRoot: make([]byte, 32),
					Source:          &ethpb.Checkpoint{},
					Target:          &ethpb.Checkpoint{},
				}
			}
		}

		s, err := stateAltair.InitializeFromProtoUnsafe(bState)
		require.NoError(t, err)
		wsb, err := wrapper.WrappedSignedBeaconBlock(b)
		require.NoError(t, err)
		r, err := altair.ProcessAttestationsNoVerifySignature(context.Background(), s, wsb)
		if err != nil && r != nil {
			t.Fatalf("return value should be nil on err. found: %v on error: %v for state: %v and block: %v", r, err, bState, b)
		}
	}
}

func TestSetParticipationAndRewardProposer(t *testing.T) {
	cfg := params.BeaconConfig()
	sourceFlagIndex := cfg.TimelySourceFlagIndex
	targetFlagIndex := cfg.TimelyTargetFlagIndex
	headFlagIndex := cfg.TimelyHeadFlagIndex
	voteFlagIndex := cfg.DAGTimelyVotingFlagIndex
	tests := []struct {
		name                string
		indices             []uint64
		epochParticipation  []byte
		participatedFlags   map[uint8]bool
		epoch               types.Epoch
		wantedBalance       uint64
		wantedParticipation []byte
	}{
		{name: "none participated",
			indices: []uint64{}, epochParticipation: []byte{0, 0, 0, 0, 0, 0, 0, 0}, participatedFlags: map[uint8]bool{
				sourceFlagIndex: false,
				targetFlagIndex: false,
				headFlagIndex:   false,
				voteFlagIndex:   false,
			},
			wantedParticipation: []byte{0, 0, 0, 0, 0, 0, 0, 0},
			wantedBalance:       32_000_000_000_000,
		},
		{name: "some participated without flags",
			indices: []uint64{0, 1, 2, 3}, epochParticipation: []byte{0, 0, 0, 0, 0, 0, 0, 0}, participatedFlags: map[uint8]bool{
				sourceFlagIndex: false,
				targetFlagIndex: false,
				headFlagIndex:   false,
				voteFlagIndex:   false,
			},
			wantedParticipation: []byte{0, 0, 0, 0, 0, 0, 0, 0},
			wantedBalance:       32_000_314_182_224,
		},
		{name: "some participated with some flags",
			indices: []uint64{0, 1, 2, 3}, epochParticipation: []byte{0, 0, 0, 0, 0, 0, 0, 0}, participatedFlags: map[uint8]bool{
				sourceFlagIndex: true,
				targetFlagIndex: true,
				headFlagIndex:   false,
				voteFlagIndex:   false,
			},
			wantedParticipation: []byte{3, 3, 3, 3, 0, 0, 0, 0},
			wantedBalance:       32_000_471_273_336,
		},
		{name: "all participated with some flags",
			indices: []uint64{0, 1, 2, 3, 4, 5, 6, 7}, epochParticipation: []byte{0, 0, 0, 0, 0, 0, 0, 0}, participatedFlags: map[uint8]bool{
				sourceFlagIndex: true,
				targetFlagIndex: false,
				headFlagIndex:   false,
				voteFlagIndex:   false,
			},
			wantedParticipation: []byte{1, 1, 1, 1, 1, 1, 1, 1},
			wantedBalance:       32_000_785_455_560,
		},
		{name: "all participated with all flags",
			indices: []uint64{0, 1, 2, 3, 4, 5, 6, 7}, epochParticipation: []byte{0, 0, 0, 0, 0, 0, 0, 0}, participatedFlags: map[uint8]bool{
				sourceFlagIndex: true,
				targetFlagIndex: true,
				headFlagIndex:   true,
				voteFlagIndex:   true,
			},
			wantedParticipation: []byte{15, 15, 15, 15, 15, 15, 15, 15},
			wantedBalance:       32_001_256_728_896,
		},
	}
	for _, test := range tests {
		t.Skip()
		t.Run(test.name, func(t *testing.T) {
			beaconState, _ := util.DeterministicGenesisStateAltair(t, params.BeaconConfig().MaxValidatorsPerCommittee)
			require.NoError(t, beaconState.SetSlot(params.BeaconConfig().SlotsPerEpoch))

			currentEpoch := time.CurrentEpoch(beaconState)
			if test.epoch == currentEpoch {
				require.NoError(t, beaconState.SetCurrentParticipationBits(test.epochParticipation))
			} else {
				require.NoError(t, beaconState.SetPreviousParticipationBits(test.epochParticipation))
			}

			st, err := altair.SetParticipationAndRewardProposer(context.Background(), beaconState, test.epoch, test.indices, test.participatedFlags, []byte{})
			require.NoError(t, err)

			i, err := helpers.BeaconProposerIndex(context.Background(), st)
			require.NoError(t, err)
			b, err := beaconState.BalanceAtIndex(i)
			require.NoError(t, err)
			require.Equal(t, test.wantedBalance, b)

			if test.epoch == currentEpoch {
				p, err := beaconState.CurrentEpochParticipation()
				require.NoError(t, err)
				require.DeepSSZEqual(t, test.wantedParticipation, p)
			} else {
				p, err := beaconState.PreviousEpochParticipation()
				require.NoError(t, err)
				require.DeepSSZEqual(t, test.wantedParticipation, p)
			}
		})
	}
}

func TestEpochParticipation(t *testing.T) {
	beaconState, _ := util.DeterministicGenesisStateAltair(t, params.BeaconConfig().MaxValidatorsPerCommittee)
	cfg := params.BeaconConfig()
	sourceFlagIndex := cfg.TimelySourceFlagIndex
	targetFlagIndex := cfg.TimelyTargetFlagIndex
	headFlagIndex := cfg.TimelyHeadFlagIndex
	votingFlagIndex := cfg.DAGTimelyVotingFlagIndex
	tests := []struct {
		name                     string
		indices                  []uint64
		epochParticipation       []byte
		participatedFlags        map[uint8]bool
		wantedNumerator          uint64
		wantedEpochParticipation []byte
	}{
		{name: "none participated",
			indices: []uint64{}, epochParticipation: []byte{0, 0, 0, 0, 0, 0, 0, 0}, participatedFlags: map[uint8]bool{
				sourceFlagIndex: false,
				targetFlagIndex: false,
				headFlagIndex:   false,
				votingFlagIndex: false,
			},
			wantedNumerator:          0,
			wantedEpochParticipation: []byte{0, 0, 0, 0, 0, 0, 0, 0},
		},
		{name: "some participated without flags",
			indices: []uint64{0, 1, 2, 3}, epochParticipation: []byte{0, 0, 0, 0, 0, 0, 0, 0}, participatedFlags: map[uint8]bool{
				sourceFlagIndex: false,
				targetFlagIndex: false,
				headFlagIndex:   false,
				votingFlagIndex: false,
			},
			wantedNumerator:          0,
			wantedEpochParticipation: []byte{0, 0, 0, 0, 0, 0, 0, 0},
		},
		{name: "some participated with some flags",
			indices: []uint64{0, 1, 2, 3}, epochParticipation: []byte{0, 0, 0, 0, 0, 0, 0, 0}, participatedFlags: map[uint8]bool{
				sourceFlagIndex: true,
				targetFlagIndex: true,
				headFlagIndex:   false,
				votingFlagIndex: false,
			},
			wantedNumerator:          49_676_568,
			wantedEpochParticipation: []byte{3, 3, 3, 3, 0, 0, 0, 0},
		},
		{name: "all participated with some flags",
			indices: []uint64{0, 1, 2, 3, 4, 5, 6, 7}, epochParticipation: []byte{0, 0, 0, 0, 0, 0, 0, 0}, participatedFlags: map[uint8]bool{
				sourceFlagIndex: true,
				targetFlagIndex: false,
				headFlagIndex:   false,
				votingFlagIndex: false,
			},
			wantedNumerator:          49_676_568,
			wantedEpochParticipation: []byte{1, 1, 1, 1, 1, 1, 1, 1},
		},
		{name: "all participated with all flags",
			indices: []uint64{0, 1, 2, 3, 4, 5, 6, 7}, epochParticipation: []byte{0, 0, 0, 0, 0, 0, 0, 0}, participatedFlags: map[uint8]bool{
				sourceFlagIndex: true,
				targetFlagIndex: true,
				headFlagIndex:   true,
				votingFlagIndex: true,
			},
			wantedNumerator:          198_706_272,
			wantedEpochParticipation: []byte{15, 15, 15, 15, 15, 15, 15, 15},
		},
	}
	for _, test := range tests {
		n, p, err := altair.EpochParticipation(beaconState, test.indices, test.epochParticipation, test.participatedFlags)
		require.NoError(t, err)
		require.Equal(t, test.wantedNumerator, n)
		require.DeepSSZEqual(t, test.wantedEpochParticipation, p)
	}
}

func TestAttestationParticipationFlagIndices(t *testing.T) {
	beaconState, _ := util.DeterministicGenesisStateAltair(t, params.BeaconConfig().MaxValidatorsPerCommittee)
	require.NoError(t, beaconState.SetSlot(1))
	cfg := params.BeaconConfig()
	sourceFlagIndex := cfg.TimelySourceFlagIndex
	targetFlagIndex := cfg.TimelyTargetFlagIndex
	headFlagIndex := cfg.TimelyHeadFlagIndex
	votingFlagIndex := cfg.DAGTimelyVotingFlagIndex

	tests := []struct {
		name                 string
		inputState           state.BeaconState
		inputData            *ethpb.AttestationData
		inputDelay           types.Slot
		participationIndices map[uint8]bool
	}{
		{
			name: "none",
			inputState: func() state.BeaconState {
				return beaconState
			}(),
			inputData: &ethpb.AttestationData{
				Source: &ethpb.Checkpoint{Root: params.BeaconConfig().ZeroHash[:]},
				Target: &ethpb.Checkpoint{},
			},
			inputDelay: params.BeaconConfig().SlotsPerEpoch,
			participationIndices: map[uint8]bool{
				votingFlagIndex: false,
			},
		},
		{
			name: "participated source",
			inputState: func() state.BeaconState {
				return beaconState
			}(),
			inputData: &ethpb.AttestationData{
				Source: &ethpb.Checkpoint{Root: params.BeaconConfig().ZeroHash[:]},
				Target: &ethpb.Checkpoint{},
			},
			inputDelay: types.Slot(math.IntegerSquareRoot(uint64(cfg.SlotsPerEpoch)) - 1),
			participationIndices: map[uint8]bool{
				sourceFlagIndex: true,
				votingFlagIndex: false,
			},
		},
		{
			name: "participated source and target",
			inputState: func() state.BeaconState {
				return beaconState
			}(),
			inputData: &ethpb.AttestationData{
				Source: &ethpb.Checkpoint{Root: params.BeaconConfig().ZeroHash[:]},
				Target: &ethpb.Checkpoint{Root: params.BeaconConfig().ZeroHash[:]},
			},
			inputDelay: types.Slot(math.IntegerSquareRoot(uint64(cfg.SlotsPerEpoch)) - 1),
			participationIndices: map[uint8]bool{
				sourceFlagIndex: true,
				targetFlagIndex: true,
				votingFlagIndex: false,
			},
		},
		{
			name: "participated source and target and head",
			inputState: func() state.BeaconState {
				return beaconState
			}(),
			inputData: &ethpb.AttestationData{
				BeaconBlockRoot: params.BeaconConfig().ZeroHash[:],
				Source:          &ethpb.Checkpoint{Root: params.BeaconConfig().ZeroHash[:]},
				Target:          &ethpb.Checkpoint{Root: params.BeaconConfig().ZeroHash[:]},
			},
			inputDelay: 1,
			participationIndices: map[uint8]bool{
				sourceFlagIndex: true,
				targetFlagIndex: true,
				headFlagIndex:   true,
				votingFlagIndex: true,
			},
		},
	}

	for _, test := range tests {
		flagIndices, err := altair.AttestationParticipationFlagIndices(test.inputState, test.inputData, test.inputDelay)
		require.NoError(t, err)
		require.DeepEqual(t, test.participationIndices, flagIndices)
	}
}

func TestMatchingStatus(t *testing.T) {
	beaconState, _ := util.DeterministicGenesisStateAltair(t, params.BeaconConfig().MaxValidatorsPerCommittee)
	require.NoError(t, beaconState.SetSlot(1))
	tests := []struct {
		name          string
		inputState    state.BeaconState
		inputData     *ethpb.AttestationData
		inputCheckpt  *ethpb.Checkpoint
		matchedSource bool
		matchedTarget bool
		matchedHead   bool
	}{
		{
			name:       "non matched",
			inputState: beaconState,
			inputData: &ethpb.AttestationData{
				Source: &ethpb.Checkpoint{Epoch: 1},
				Target: &ethpb.Checkpoint{},
			},
			inputCheckpt: &ethpb.Checkpoint{},
		},
		{
			name:       "source matched",
			inputState: beaconState,
			inputData: &ethpb.AttestationData{
				Source: &ethpb.Checkpoint{},
				Target: &ethpb.Checkpoint{},
			},
			inputCheckpt:  &ethpb.Checkpoint{},
			matchedSource: true,
		},
		{
			name:       "target matched",
			inputState: beaconState,
			inputData: &ethpb.AttestationData{
				Source: &ethpb.Checkpoint{Epoch: 1},
				Target: &ethpb.Checkpoint{Root: params.BeaconConfig().ZeroHash[:]},
			},
			inputCheckpt:  &ethpb.Checkpoint{},
			matchedTarget: true,
		},
		{
			name:       "head matched",
			inputState: beaconState,
			inputData: &ethpb.AttestationData{
				Source:          &ethpb.Checkpoint{Epoch: 1},
				Target:          &ethpb.Checkpoint{},
				BeaconBlockRoot: params.BeaconConfig().ZeroHash[:],
			},
			inputCheckpt: &ethpb.Checkpoint{},
			matchedHead:  true,
		},
		{
			name:       "everything matched",
			inputState: beaconState,
			inputData: &ethpb.AttestationData{
				Source:          &ethpb.Checkpoint{},
				Target:          &ethpb.Checkpoint{Root: params.BeaconConfig().ZeroHash[:]},
				BeaconBlockRoot: params.BeaconConfig().ZeroHash[:],
			},
			inputCheckpt:  &ethpb.Checkpoint{},
			matchedSource: true,
			matchedTarget: true,
			matchedHead:   true,
		},
	}

	for _, test := range tests {
		src, tgt, head, err := altair.MatchingStatus(test.inputState, test.inputData, test.inputCheckpt)
		require.NoError(t, err)
		require.Equal(t, test.matchedSource, src)
		require.Equal(t, test.matchedTarget, tgt)
		require.Equal(t, test.matchedHead, head)
	}
}

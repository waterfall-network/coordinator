package helpers_test

import (
	"context"
	"fmt"
	"testing"

	types "github.com/prysmaticlabs/eth2-types"
	"github.com/prysmaticlabs/go-bitfield"
	"github.com/waterfall-foundation/coordinator/beacon-chain/core/helpers"
	ethpb "github.com/waterfall-foundation/coordinator/proto/prysm/v1alpha1"
	"github.com/waterfall-foundation/coordinator/testing/assert"
	"github.com/waterfall-foundation/coordinator/testing/util"
	gwatCommon "github.com/waterfall-foundation/gwat/common"
)

func TestBlockVotingsCalcFinalization_finalization_OK(t *testing.T) {
	state, keys := util.DeterministicGenesisState(t, 128)

	sig := keys[0].Sign([]byte{'t', 'e', 's', 't'})

	list := bitfield.NewBitlist(4)
	list.SetBitAt(0, true)
	list.SetBitAt(1, true)
	list.SetBitAt(2, true)

	root_0 := gwatCommon.BytesToHash([]byte("root-0--------------------------"))
	var atts_0 []*ethpb.Attestation
	atts_0 = append(atts_0, &ethpb.Attestation{
		Data: &ethpb.AttestationData{
			CommitteeIndex:  0,
			Slot:            types.Slot(5),
			BeaconBlockRoot: root_0[:],
		},
		Signature:       sig.Marshal(),
		AggregationBits: list,
	})

	root_1 := gwatCommon.BytesToHash([]byte("root-1--------------------------"))
	var atts_1 []*ethpb.Attestation

	atts_1 = append(atts_1, &ethpb.Attestation{
		Data: &ethpb.AttestationData{
			CommitteeIndex:  0,
			Slot:            types.Slot(6),
			BeaconBlockRoot: root_1[:],
		},
		Signature:       sig.Marshal(),
		AggregationBits: list,
	})

	root_2 := gwatCommon.BytesToHash([]byte("root-2--------------------------"))
	var atts_2 []*ethpb.Attestation
	atts_2 = append(atts_2, &ethpb.Attestation{
		Data: &ethpb.AttestationData{
			CommitteeIndex:  0,
			Slot:            types.Slot(7),
			BeaconBlockRoot: root_2[:],
		},
		Signature:       sig.Marshal(),
		AggregationBits: list,
	})

	blobVotings := []*ethpb.BlockVoting{
		{
			Root:           root_0[:],
			TotalAttesters: 8,
			Candidates: gwatCommon.HashArray{
				gwatCommon.Hash{0xff, 0x02},
				gwatCommon.Hash{0xff, 0x01},
				gwatCommon.Hash{0xff, 0xff},

				gwatCommon.Hash{0x11, 0x11},
				gwatCommon.Hash{0x11, 0x22},
				gwatCommon.Hash{0x11, 0x33},

				gwatCommon.Hash{0xaa, 0x66},
			}.ToBytes(),
			Attestations: atts_0,
		},
		{
			Root:           root_1[:],
			TotalAttesters: 8,
			Candidates: gwatCommon.HashArray{
				gwatCommon.Hash{0xff, 0x03},
				gwatCommon.Hash{0xff, 0x02},
				gwatCommon.Hash{0xff, 0x01},
				gwatCommon.Hash{0xff, 0xff},

				gwatCommon.Hash{0x11, 0x11},
				gwatCommon.Hash{0x11, 0x22},
				gwatCommon.Hash{0x11, 0x33},

				gwatCommon.Hash{0xaa, 0x55},
			}.ToBytes(),
			Attestations: atts_1,
		},
		{
			Root:           root_2[:],
			TotalAttesters: 8,
			Candidates: gwatCommon.HashArray{

				gwatCommon.Hash{0x11, 0x11},
				gwatCommon.Hash{0x11, 0x22},

				gwatCommon.Hash{0x22, 0x33},
				gwatCommon.Hash{0xaa, 0x77},
			}.ToBytes(),
			Attestations: atts_2,
		},
	}

	want := gwatCommon.HashArray{
		gwatCommon.Hash{0x11, 0x11},
		gwatCommon.Hash{0x11, 0x22},
	}

	finalization, err := helpers.BlockVotingsCalcFinalization(context.Background(), state, blobVotings, gwatCommon.Hash{0xff, 0xff})

	assert.NoError(t, err)
	assert.DeepEqual(t, fmt.Sprintf("%v", want), fmt.Sprintf("%v", finalization))
}

func TestBlockVotingsCalcFinalization_fin_after_3_slots(t *testing.T) {
	state, keys := util.DeterministicGenesisState(t, 128)

	sig := keys[0].Sign([]byte{'t', 'e', 's', 't'})

	list := bitfield.NewBitlist(4)
	list.SetBitAt(0, true)
	list.SetBitAt(1, true)
	list.SetBitAt(2, true)

	root_0 := gwatCommon.BytesToHash([]byte("root-0--------------------------"))
	var atts_0 []*ethpb.Attestation
	atts_0 = append(atts_0, &ethpb.Attestation{
		Data: &ethpb.AttestationData{
			CommitteeIndex:  0,
			Slot:            types.Slot(5),
			BeaconBlockRoot: root_0[:],
		},
		Signature:       sig.Marshal(),
		AggregationBits: list,
	})

	atts_0 = append(atts_0, &ethpb.Attestation{
		Data: &ethpb.AttestationData{
			CommitteeIndex:  0,
			Slot:            types.Slot(6),
			BeaconBlockRoot: root_0[:],
		},
		Signature:       sig.Marshal(),
		AggregationBits: list,
	})

	atts_0 = append(atts_0, &ethpb.Attestation{
		Data: &ethpb.AttestationData{
			CommitteeIndex:  0,
			Slot:            types.Slot(7),
			BeaconBlockRoot: root_0[:],
		},
		Signature:       sig.Marshal(),
		AggregationBits: list,
	})

	blobVotings := []*ethpb.BlockVoting{
		{
			Root:           root_0[:],
			TotalAttesters: 5,
			Candidates: gwatCommon.HashArray{
				gwatCommon.Hash{0x11, 0x11},
			}.ToBytes(),
			Attestations: atts_0,
		},
	}

	want := gwatCommon.HashArray{
		gwatCommon.Hash{0x11, 0x11},
	}

	finalization, err := helpers.BlockVotingsCalcFinalization(context.Background(), state, blobVotings, gwatCommon.Hash{0xff, 0xff})

	assert.NoError(t, err)
	assert.DeepEqual(t, fmt.Sprintf("%v", want), fmt.Sprintf("%v", finalization))
}

func TestBlockVotingsCalcFinalization_fin_after_3_slots_v2(t *testing.T) {
	state, keys := util.DeterministicGenesisState(t, 128)

	sig := keys[0].Sign([]byte{'t', 'e', 's', 't'})

	list := bitfield.NewBitlist(4)
	list.SetBitAt(0, true)
	list.SetBitAt(1, true)
	list.SetBitAt(2, true)

	root_0 := gwatCommon.BytesToHash([]byte("root-0--------------------------"))
	var atts_0 []*ethpb.Attestation
	atts_0 = append(atts_0, &ethpb.Attestation{
		Data: &ethpb.AttestationData{
			CommitteeIndex:  0,
			Slot:            types.Slot(5),
			BeaconBlockRoot: root_0[:],
		},
		Signature:       sig.Marshal(),
		AggregationBits: list,
	})

	blobVotings := []*ethpb.BlockVoting{
		{
			Root:           root_0[:],
			TotalAttesters: 5,
			Candidates: gwatCommon.HashArray{
				gwatCommon.Hash{0x11, 0x11},
				gwatCommon.Hash{0x11, 0x22},
			}.ToBytes(),
			Attestations: atts_0,
		},
		{
			Root:           root_0[:],
			TotalAttesters: 6,
			Candidates: gwatCommon.HashArray{
				gwatCommon.Hash{0x11, 0x11},
				gwatCommon.Hash{0x11, 0x22},
			}.ToBytes(),
			Attestations: atts_0,
		},
		{
			Root:           root_0[:],
			TotalAttesters: 7,
			Candidates: gwatCommon.HashArray{
				gwatCommon.Hash{0x11, 0x11},
				gwatCommon.Hash{0x11, 0x22},
			}.ToBytes(),
			Attestations: atts_0,
		},
	}

	want := gwatCommon.HashArray{
		gwatCommon.Hash{0x11, 0x11},
		gwatCommon.Hash{0x11, 0x22},
	}

	finalization, err := helpers.BlockVotingsCalcFinalization(context.Background(), state, blobVotings, gwatCommon.Hash{0xff, 0xff})

	assert.NoError(t, err)
	assert.DeepEqual(t, fmt.Sprintf("%v", want), fmt.Sprintf("%v", finalization))
}

package voluntaryexits

import (
	"context"
	"reflect"
	"testing"

	types "github.com/prysmaticlabs/eth2-types"
	v1 "gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/state/v1"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/config/params"
	ethpb "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/require"
	"google.golang.org/protobuf/proto"
)

func TestPool_InsertVoluntaryExit(t *testing.T) {
	type fields struct {
		pending []*ethpb.VoluntaryExit
	}
	type args struct {
		exit *ethpb.VoluntaryExit
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []*ethpb.VoluntaryExit
	}{
		{
			name: "Prevent inserting nil exit",
			fields: fields{
				pending: make([]*ethpb.VoluntaryExit, 0),
			},
			args: args{
				exit: nil,
			},
			want: []*ethpb.VoluntaryExit{},
		},
		{
			name: "Prevent inserting malformed exit",
			fields: fields{
				pending: make([]*ethpb.VoluntaryExit, 0),
			},
			args: args{
				exit: &ethpb.VoluntaryExit{},
			},
			want: []*ethpb.VoluntaryExit{},
		},
		{
			name: "Empty list",
			fields: fields{
				pending: make([]*ethpb.VoluntaryExit, 0),
			},
			args: args{
				exit: &ethpb.VoluntaryExit{
					Epoch:          12,
					ValidatorIndex: 1,
				},
			},
			want: []*ethpb.VoluntaryExit{
				{
					Epoch:          12,
					ValidatorIndex: 1,
				},
			},
		},
		{
			name: "Duplicate identical exit",
			fields: fields{
				pending: []*ethpb.VoluntaryExit{
					{

						Epoch:          12,
						ValidatorIndex: 1,
					},
				},
			},
			args: args{
				exit: &ethpb.VoluntaryExit{
					Epoch:          12,
					ValidatorIndex: 1,
				},
			},
			want: []*ethpb.VoluntaryExit{
				{
					Epoch:          12,
					ValidatorIndex: 1,
				},
			},
		},
		{
			name: "Duplicate exit in pending list",
			fields: fields{
				pending: []*ethpb.VoluntaryExit{
					{
						Epoch:          12,
						ValidatorIndex: 1,
					},
				},
			},
			args: args{
				exit: &ethpb.VoluntaryExit{
					Epoch:          12,
					ValidatorIndex: 1,
				},
			},
			want: []*ethpb.VoluntaryExit{
				{
					Epoch:          12,
					ValidatorIndex: 1,
				},
			},
		},
		{
			name: "Duplicate validator index",
			fields: fields{
				pending: []*ethpb.VoluntaryExit{
					{
						Epoch:          12,
						ValidatorIndex: 1,
					},
				},
			},
			args: args{
				exit: &ethpb.VoluntaryExit{
					Epoch:          20,
					ValidatorIndex: 1,
				},
			},
			want: []*ethpb.VoluntaryExit{
				{
					Epoch:          12,
					ValidatorIndex: 1,
				},
			},
		},
		{
			name: "Duplicate received with more favorable exit epoch",
			fields: fields{
				pending: []*ethpb.VoluntaryExit{
					{
						Epoch:          12,
						ValidatorIndex: 1,
					},
				},
			},
			args: args{
				exit: &ethpb.VoluntaryExit{
					Epoch:          4,
					ValidatorIndex: 1,
				},
			},
			want: []*ethpb.VoluntaryExit{
				{
					Epoch:          4,
					ValidatorIndex: 1,
				},
			},
		},
		{
			name: "Exit for already exited validator",
			fields: fields{
				pending: []*ethpb.VoluntaryExit{},
			},
			args: args{
				exit: &ethpb.VoluntaryExit{
					Epoch:          12,
					ValidatorIndex: 2,
				},
			},
			want: []*ethpb.VoluntaryExit{},
		},
		{
			name: "Maintains sorted order",
			fields: fields{
				pending: []*ethpb.VoluntaryExit{
					{
						Epoch:          12,
						ValidatorIndex: 0,
					},
					{
						Epoch:          12,
						ValidatorIndex: 2,
					},
				},
			},
			args: args{
				exit: &ethpb.VoluntaryExit{
					Epoch:          10,
					ValidatorIndex: 1,
				},
			},
			want: []*ethpb.VoluntaryExit{
				{
					Epoch:          12,
					ValidatorIndex: 0,
				},
				{
					Epoch:          10,
					ValidatorIndex: 1,
				},
				{
					Epoch:          12,
					ValidatorIndex: 2,
				},
			},
		},
	}
	ctx := context.Background()
	validators := []*ethpb.Validator{
		{ // 0
			ExitEpoch: params.BeaconConfig().FarFutureEpoch,
		},
		{ // 1
			ExitEpoch: params.BeaconConfig().FarFutureEpoch,
		},
		{ // 2 - Already exited.
			ExitEpoch: 15,
		},
		{ // 3
			ExitEpoch: params.BeaconConfig().FarFutureEpoch,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Pool{
				pending: tt.fields.pending,
			}
			s, err := v1.InitializeFromProtoUnsafe(&ethpb.BeaconState{Validators: validators})
			require.NoError(t, err)
			p.InsertVoluntaryExit(ctx, s, tt.args.exit)
			if len(p.pending) != len(tt.want) {
				t.Fatalf("Mismatched lengths of pending list. Got %d, wanted %d.", len(p.pending), len(tt.want))
			}
			for i := range p.pending {
				if !proto.Equal(p.pending[i], tt.want[i]) {
					t.Errorf("Pending exit at index %d does not match expected. Got=%v wanted=%v", i, p.pending[i], tt.want[i])
				}
			}
		})
	}
}

func TestPool_MarkIncluded(t *testing.T) {
	type fields struct {
		pending []*ethpb.VoluntaryExit
	}
	type args struct {
		exit *ethpb.VoluntaryExit
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   fields
	}{
		{
			name: "Removes from pending list",
			fields: fields{
				pending: []*ethpb.VoluntaryExit{
					{ValidatorIndex: 1},
					{ValidatorIndex: 2},
					{ValidatorIndex: 3},
				},
			},
			args: args{
				exit: &ethpb.VoluntaryExit{ValidatorIndex: 2},
			},
			want: fields{
				pending: []*ethpb.VoluntaryExit{{ValidatorIndex: 1}, {ValidatorIndex: 3}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Pool{
				pending: tt.fields.pending,
			}
			p.MarkIncluded(tt.args.exit)
			if len(p.pending) != len(tt.want.pending) {
				t.Fatalf("Mismatched lengths of pending list. Got %d, wanted %d.", len(p.pending), len(tt.want.pending))
			}
			for i := range p.pending {
				if !proto.Equal(p.pending[i], tt.want.pending[i]) {
					t.Errorf("Pending exit at index %d does not match expected. Got=%v wanted=%v", i, p.pending[i], tt.want.pending[i])
				}
			}
		})
	}
}

func TestPool_PendingExits(t *testing.T) {
	type fields struct {
		pending []*ethpb.VoluntaryExit
		noLimit bool
	}
	type args struct {
		slot types.Slot
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []*ethpb.VoluntaryExit
	}{
		{
			name: "Empty list",
			fields: fields{
				pending: []*ethpb.VoluntaryExit{},
			},
			args: args{
				slot: 100000,
			},
			want: []*ethpb.VoluntaryExit{},
		},
		{
			name: "All eligible",
			fields: fields{
				pending: []*ethpb.VoluntaryExit{
					{Epoch: 0},
					{Epoch: 1},
					{Epoch: 2},
					{Epoch: 3},
					{Epoch: 4},
				},
			},
			args: args{
				slot: 1000000,
			},
			want: []*ethpb.VoluntaryExit{
				{Epoch: 0},
				{Epoch: 1},
				{Epoch: 2},
				{Epoch: 3},
				{Epoch: 4},
			},
		},
		{
			name: "All eligible, above max",
			fields: fields{
				noLimit: true,
				pending: []*ethpb.VoluntaryExit{
					{Epoch: 0},
					{Epoch: 1},
					{Epoch: 2},
					{Epoch: 3},
					{Epoch: 4},
					{Epoch: 5},
					{Epoch: 6},
					{Epoch: 7},
					{Epoch: 8},
					{Epoch: 9},
					{Epoch: 10},
					{Epoch: 11},
					{Epoch: 12},
					{Epoch: 13},
					{Epoch: 14},
					{Epoch: 15},
					{Epoch: 16},
					{Epoch: 17},
					{Epoch: 18},
					{Epoch: 19},
				},
			},
			args: args{
				slot: 1000000,
			},
			want: []*ethpb.VoluntaryExit{
				{Epoch: 0},
				{Epoch: 1},
				{Epoch: 2},
				{Epoch: 3},
				{Epoch: 4},
				{Epoch: 5},
				{Epoch: 6},
				{Epoch: 7},
				{Epoch: 8},
				{Epoch: 9},
				{Epoch: 10},
				{Epoch: 11},
				{Epoch: 12},
				{Epoch: 13},
				{Epoch: 14},
				{Epoch: 15},
				{Epoch: 16},
				{Epoch: 17},
				{Epoch: 18},
				{Epoch: 19},
			},
		},
		{
			name: "All eligible, block max",
			fields: fields{
				pending: []*ethpb.VoluntaryExit{
					{Epoch: 0},
					{Epoch: 1},
					{Epoch: 2},
					{Epoch: 3},
					{Epoch: 4},
					{Epoch: 5},
					{Epoch: 6},
					{Epoch: 7},
					{Epoch: 8},
					{Epoch: 9},
					{Epoch: 10},
					{Epoch: 11},
					{Epoch: 12},
					{Epoch: 13},
					{Epoch: 14},
					{Epoch: 15},
					{Epoch: 16},
					{Epoch: 17},
					{Epoch: 18},
					{Epoch: 19},
				},
			},
			args: args{
				slot: 1000000,
			},
			want: []*ethpb.VoluntaryExit{
				{Epoch: 0},
				{Epoch: 1},
				{Epoch: 2},
				{Epoch: 3},
				{Epoch: 4},
				{Epoch: 5},
				{Epoch: 6},
				{Epoch: 7},
				{Epoch: 8},
				{Epoch: 9},
				{Epoch: 10},
				{Epoch: 11},
				{Epoch: 12},
				{Epoch: 13},
				{Epoch: 14},
				{Epoch: 15},
			},
		},
		{
			name: "Some eligible",
			fields: fields{
				pending: []*ethpb.VoluntaryExit{
					{Epoch: 0},
					{Epoch: 3},
					{Epoch: 4},
					{Epoch: 2},
					{Epoch: 1},
				},
			},
			args: args{
				slot: 2 * params.BeaconConfig().SlotsPerEpoch,
			},
			want: []*ethpb.VoluntaryExit{
				{Epoch: 0},
				{Epoch: 2},
				{Epoch: 1},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Pool{
				pending: tt.fields.pending,
			}
			s, err := v1.InitializeFromProtoUnsafe(&ethpb.BeaconState{Validators: []*ethpb.Validator{{ExitEpoch: params.BeaconConfig().FarFutureEpoch}}})
			require.NoError(t, err)
			if got := p.PendingExits(s, tt.args.slot, tt.fields.noLimit); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PendingExits() = %v, want %v", got, tt.want)
			}
		})
	}
}

package altair_test

import (
	"fmt"
	"math"
	"testing"

	types "github.com/prysmaticlabs/eth2-types"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/core/altair"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/state"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/config/params"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/require"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/util"
)

func Test_BaseReward(t *testing.T) {
	genState := func(valCount uint64) state.ReadOnlyBeaconState {
		s, _ := util.DeterministicGenesisStateAltair(t, valCount)
		return s
	}
	tests := []struct {
		name      string
		valIdx    types.ValidatorIndex
		st        state.ReadOnlyBeaconState
		want      uint64
		errString string
	}{
		{
			name:      "unknown validator",
			valIdx:    2,
			st:        genState(1),
			want:      0,
			errString: "index 2 out of range",
		},
		{
			name:      "active balance is 32eth",
			valIdx:    0,
			st:        genState(1),
			want:      114486688,
			errString: "",
		},
		{
			name:      "active balance is 32eth * target committee size",
			valIdx:    0,
			st:        genState(params.BeaconConfig().TargetCommitteeSize),
			want:      10119264,
			errString: "",
		},
		{
			name:      "active balance is 32eth * max validator per  committee size",
			valIdx:    0,
			st:        genState(params.BeaconConfig().MaxValidatorsPerCommittee),
			want:      2529792,
			errString: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := altair.BaseReward(tt.st, tt.valIdx)
			if (err != nil) && (tt.errString != "") {
				require.ErrorContains(t, tt.errString, err)
				return
			}
			require.Equal(t, tt.want, got)
		})
	}
}

func Test_BaseRewardWithTotalBalance(t *testing.T) {
	s, _ := util.DeterministicGenesisStateAltair(t, 1)
	tests := []struct {
		name          string
		valIdx        types.ValidatorIndex
		activeBalance uint64
		want          uint64
		errString     string
	}{
		{
			name:          "active balance is 0",
			valIdx:        0,
			activeBalance: 0,
			want:          0,
			errString:     "active balance can't be 0",
		},
		{
			name:          "unknown validator",
			valIdx:        2,
			activeBalance: 1,
			want:          0,
			errString:     "index 2 out of range",
		},
		{
			name:          "active balance is 1",
			valIdx:        0,
			activeBalance: 1,
			want:          204800000000000,
			errString:     "",
		},
		{
			name:          "active balance is 1eth",
			valIdx:        0,
			activeBalance: params.BeaconConfig().EffectiveBalanceIncrement,
			want:          647636032,
			errString:     "",
		},
		{
			name:          "active balance is 32eth",
			valIdx:        0,
			activeBalance: params.BeaconConfig().MaxEffectiveBalance,
			want:          114486688,
			errString:     "",
		},
		{
			name:          "active balance is 32eth * 1m validators",
			valIdx:        0,
			activeBalance: params.BeaconConfig().MaxEffectiveBalance * 1e9,
			want:          69376,
			errString:     "",
		},
		{
			name:          "active balance is max uint64",
			valIdx:        0,
			activeBalance: math.MaxUint64,
			want:          47680,
			errString:     "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := altair.BaseRewardWithTotalBalance(s, tt.valIdx, tt.activeBalance)
			if (err != nil) && (tt.errString != "") {
				require.ErrorContains(t, tt.errString, err)
				return
			}
			require.Equal(t, tt.want, got)
		})
	}
}

func Test_BaseRewardPerIncrement(t *testing.T) {
	tests := []struct {
		name          string
		activeBalance uint64
		want          uint64
		errString     string
	}{
		{
			name:          "active balance is 0",
			activeBalance: 0,
			want:          0,
			errString:     "active balance can't be 0",
		},
		{
			name:          "active balance is 1",
			activeBalance: 1,
			want:          6400000000000,
			errString:     "",
		},
		{
			name:          "active balance is 1eth",
			activeBalance: params.BeaconConfig().EffectiveBalanceIncrement,
			want:          20238626,
			errString:     "",
		},
		{
			name:          "active balance is 32eth",
			activeBalance: params.BeaconConfig().MaxEffectiveBalance,
			want:          3577709,
			errString:     "",
		},
		{
			name:          "active balance is 32eth * 1m validators",
			activeBalance: params.BeaconConfig().MaxEffectiveBalance * 1e9,
			want:          2168,
			errString:     "",
		},
		{
			name:          "active balance is max uint64",
			activeBalance: math.MaxUint64,
			want:          1490,
			errString:     "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := altair.BaseRewardPerIncrement(tt.activeBalance)
			if (err != nil) && (tt.errString != "") {
				require.ErrorContains(t, tt.errString, err)
				return
			}
			require.Equal(t, tt.want, got)
		})
	}
}

func Test_CalculateBaseReward(t *testing.T) {
	cfg := params.MinimalSpecConfig()
	noCache := "no_cache"
	fmt.Print(noCache)

	tests := []struct {
		name                   string
		validatorsNum          int
		committeesNum          uint64
		membersPerCommitteeNum uint64
		rewardMultiplier       float64
		want                   uint64
	}{
		{
			name:                   "base reward when 2048 validators and 4 committees with 2048 members each",
			validatorsNum:          2048,
			committeesNum:          4,
			membersPerCommitteeNum: 2048,
			rewardMultiplier:       2.0,
			want:                   122_727,
		},
		{
			name:                   "base reward when 2048 validators and 4 committees with 128 members each",
			validatorsNum:          2048,
			committeesNum:          4,
			membersPerCommitteeNum: 128,
			rewardMultiplier:       2.0,
			want:                   1_963_638,
		},
		{
			name:                   "base reward when 2048 validators and 4 committees with 512 members each",
			validatorsNum:          2048,
			committeesNum:          4,
			membersPerCommitteeNum: 512,
			rewardMultiplier:       2.0,
			want:                   490_909,
		},
		{
			name:                   "base reward when 2048 validators and 64 committees with 2048 members each",
			validatorsNum:          2048,
			committeesNum:          64,
			membersPerCommitteeNum: 2048,
			rewardMultiplier:       2.0,
			want:                   7_670,
		},
		{
			name:                   "base reward when 2048 validators and 64 committees with 128 members each",
			validatorsNum:          2048,
			committeesNum:          64,
			membersPerCommitteeNum: 128,
			rewardMultiplier:       2.0,
			want:                   122_727,
		},
		{
			name:                   "base reward when 300000 validators and 4 committees with 2048 members each",
			validatorsNum:          300_000,
			committeesNum:          4,
			membersPerCommitteeNum: 2048,
			rewardMultiplier:       2.0,
			want:                   1_485_379,
		},
		{
			name:                   "base reward when 300000 validators and 64 committees with 2048 members each",
			validatorsNum:          300_000,
			committeesNum:          4,
			membersPerCommitteeNum: 2048,
			rewardMultiplier:       2.0,
			want:                   1_485_379,
		},
		{
			name:                   "base reward when 300000 validators and 64 committees with 128 members each",
			validatorsNum:          300_000,
			committeesNum:          64,
			membersPerCommitteeNum: 128,
			rewardMultiplier:       2.0,
			want:                   1_485_379,
		},
		{
			name:                   "base reward when 1 validator and 4 committees with 2048 members each",
			validatorsNum:          1,
			committeesNum:          4,
			membersPerCommitteeNum: 2048,
			rewardMultiplier:       2.0,
			want:                   2_711,
		},
		{
			name:                   "base reward when 1 validator and 64 committees with 2048 members each",
			validatorsNum:          1,
			committeesNum:          64,
			membersPerCommitteeNum: 2048,
			rewardMultiplier:       2.0,
			want:                   169,
		},
		{
			name:                   "base reward when 1 validator and 64 committees with 128 members each",
			validatorsNum:          1,
			committeesNum:          64,
			membersPerCommitteeNum: 128,
			rewardMultiplier:       2.0,
			want:                   2_711,
		},
		{
			name:                   "base reward when 2048 validators and 64 committees with 128 members each with reward multiplier 1.0",
			validatorsNum:          2048,
			committeesNum:          64,
			membersPerCommitteeNum: 128,
			rewardMultiplier:       1.0,
			want:                   245_454,
		},
		{
			name:                   "base reward when 2048 validators and 64 committees with 128 members each with reward multiplier 0.5",
			validatorsNum:          2048,
			committeesNum:          64,
			membersPerCommitteeNum: 128,
			rewardMultiplier:       0.5,
			want:                   490_909,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := altair.CalculateBaseReward(cfg, tt.validatorsNum, tt.committeesNum*tt.membersPerCommitteeNum, tt.rewardMultiplier)
			require.Equal(t, tt.want, got)
		})
	}
}

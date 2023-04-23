package v3

import (
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/state/types"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/config/params"
)

func init() {
	fieldMap = make(map[types.FieldIndex]types.DataType, params.BeaconConfig().BeaconStateBellatrixFieldCount)

	// Initialize the fixed sized arrays.
	fieldMap[types.BlockRoots] = types.BasicArray
	fieldMap[types.StateRoots] = types.BasicArray
	fieldMap[types.RandaoMixes] = types.BasicArray

	// Initialize the composite arrays.
	fieldMap[types.Eth1DataVotes] = types.CompositeArray
	fieldMap[types.BlockVoting] = types.CompositeArray
	fieldMap[types.Validators] = types.CompositeArray
	fieldMap[types.Balances] = types.CompressedArray
}

// Field Aliases for values from the types package.
const (
	genesisTime                    = types.GenesisTime
	genesisValidatorsRoot          = types.GenesisValidatorsRoot
	slot                           = types.Slot
	fork                           = types.Fork
	latestBlockHeader              = types.LatestBlockHeader
	blockRoots                     = types.BlockRoots
	stateRoots                     = types.StateRoots
	historicalRoots                = types.HistoricalRoots
	eth1Data                       = types.Eth1Data
	eth1DataVotes                  = types.Eth1DataVotes
	eth1DepositIndex               = types.Eth1DepositIndex
	validators                     = types.Validators
	spineData                      = types.SpineData
	balances                       = types.Balances
	randaoMixes                    = types.RandaoMixes
	slashings                      = types.Slashings
	previousEpochParticipationBits = types.PreviousEpochParticipationBits
	currentEpochParticipationBits  = types.CurrentEpochParticipationBits
	justificationBits              = types.JustificationBits
	previousJustifiedCheckpoint    = types.PreviousJustifiedCheckpoint
	currentJustifiedCheckpoint     = types.CurrentJustifiedCheckpoint
	finalizedCheckpoint            = types.FinalizedCheckpoint
	blockVoting                    = types.BlockVoting
	inactivityScores               = types.InactivityScores
	currentSyncCommittee           = types.CurrentSyncCommittee
	nextSyncCommittee              = types.NextSyncCommittee
	latestExecutionPayloadHeader   = types.LatestExecutionPayloadHeader
)

// fieldMap keeps track of each field
// to its corresponding data type.
var fieldMap map[types.FieldIndex]types.DataType

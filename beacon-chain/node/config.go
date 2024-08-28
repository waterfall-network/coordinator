package node

import (
	"fmt"

	types "github.com/prysmaticlabs/eth2-types"
	"github.com/urfave/cli/v2"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/cmd"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/cmd/beacon-chain/flags"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/config/params"
	tracing2 "gitlab.waterfall.network/waterfall/protocol/coordinator/monitoring/tracing"
	"gitlab.waterfall.network/waterfall/protocol/gwat/common"
)

func configureTracing(cliCtx *cli.Context) error {
	return tracing2.Setup(
		"beacon-chain", // service name
		cliCtx.String(cmd.TracingProcessNameFlag.Name),
		cliCtx.String(cmd.TracingEndpointFlag.Name),
		cliCtx.Float64(cmd.TraceSampleFractionFlag.Name),
		cliCtx.Bool(cmd.EnableTracingFlag.Name),
	)
}

func configureChainConfig(cliCtx *cli.Context) {
	if cliCtx.IsSet(cmd.ChainConfigFileFlag.Name) {
		chainConfigFileName := cliCtx.String(cmd.ChainConfigFileFlag.Name)
		params.LoadChainConfigFile(chainConfigFileName, nil)
	}
}

func configureHistoricalSlasher(cliCtx *cli.Context) {
	if cliCtx.Bool(flags.HistoricalSlasherNode.Name) {
		c := params.BeaconConfig()
		// Save a state every 4 epochs.
		c.SlotsPerArchivedPoint = params.BeaconConfig().SlotsPerEpoch * 4
		params.OverrideBeaconConfig(c)
		cmdConfig := cmd.Get()
		// Allow up to 4096 attestations at a time to be requested from the beacon nde.
		cmdConfig.MaxRPCPageSize = int(params.BeaconConfig().SlotsPerEpoch.Mul(params.BeaconConfig().MaxAttestations)) // lint:ignore uintcast -- Page size should not exceed int64 with these constants.
		cmd.Init(cmdConfig)
		log.Warnf(
			"Setting %d slots per archive point and %d max RPC page size for historical slasher usage. This requires additional storage",
			c.SlotsPerArchivedPoint,
			cmdConfig.MaxRPCPageSize,
		)
	}
}

func configureSafeSlotsToImportOptimistically(cliCtx *cli.Context) {
	if cliCtx.IsSet(flags.SafeSlotsToImportOptimistically.Name) {
		c := params.BeaconConfig()
		c.SafeSlotsToImportOptimistically = types.Slot(cliCtx.Int(flags.SafeSlotsToImportOptimistically.Name))
		params.OverrideBeaconConfig(c)
	}
}

func configureSlotsPerArchivedPoint(cliCtx *cli.Context) {
	if cliCtx.IsSet(flags.SlotsPerArchivedPoint.Name) {
		c := params.BeaconConfig()
		c.SlotsPerArchivedPoint = types.Slot(cliCtx.Int(flags.SlotsPerArchivedPoint.Name))
		params.OverrideBeaconConfig(c)
	}
}

func configureEth1Config(cliCtx *cli.Context) {
	if cliCtx.IsSet(flags.ChainID.Name) {
		c := params.BeaconConfig()
		c.DepositChainID = cliCtx.Uint64(flags.ChainID.Name)
		params.OverrideBeaconConfig(c)
	}
	if cliCtx.IsSet(flags.NetworkID.Name) {
		c := params.BeaconConfig()
		c.DepositNetworkID = cliCtx.Uint64(flags.NetworkID.Name)
		params.OverrideBeaconConfig(c)
	}
	if cliCtx.IsSet(flags.DepositContractFlag.Name) {
		c := params.BeaconConfig()
		c.DepositContractAddress = cliCtx.String(flags.DepositContractFlag.Name)
		params.OverrideBeaconConfig(c)
	}
	if cliCtx.IsSet(flags.VotingRequiredSlots.Name) {
		c := params.BeaconConfig()
		c.VotingRequiredSlots = cliCtx.Int(flags.VotingRequiredSlots.Name)
		params.OverrideBeaconConfig(c)
	}
	if cliCtx.IsSet(flags.DelegatingStakeForkSlot.Name) {
		c := params.BeaconConfig()
		c.DelegateForkSlot = types.Slot(cliCtx.Uint64(flags.DelegatingStakeForkSlot.Name))
		params.OverrideBeaconConfig(c)
	}
	if cliCtx.IsSet(flags.PrefixFinForkSlot.Name) {
		c := params.BeaconConfig()
		c.PrefixFinForkSlot = types.Slot(cliCtx.Uint64(flags.PrefixFinForkSlot.Name))
		params.OverrideBeaconConfig(c)
	}
	if cliCtx.IsSet(flags.FinEth1ForkSlot.Name) {
		c := params.BeaconConfig()
		c.FinEth1ForkSlot = types.Slot(cliCtx.Uint64(flags.FinEth1ForkSlot.Name))
		params.OverrideBeaconConfig(c)
	}
	if cliCtx.IsSet(flags.BlockVotingForkSlot.Name) {
		c := params.BeaconConfig()
		c.BlockVotingForkSlot = types.Slot(cliCtx.Uint64(flags.BlockVotingForkSlot.Name))
		params.OverrideBeaconConfig(c)
	}
}

func configurePrevoting(cliCtx *cli.Context) {
	if cliCtx.IsSet(cmd.PrevotingDisableFlag.Name) {
		c := params.BeaconConfig()
		c.PrevotingDisabled = cliCtx.Bool(cmd.PrevotingDisableFlag.Name)
		params.OverrideBeaconConfig(c)
	}
}

func configureNetwork(cliCtx *cli.Context) {
	if cliCtx.IsSet(cmd.BootstrapNode.Name) {
		c := params.BeaconNetworkConfig()
		c.BootstrapNodes = cliCtx.StringSlice(cmd.BootstrapNode.Name)
		params.OverrideBeaconNetworkConfig(c)
	}
}

func configureInteropConfig(cliCtx *cli.Context) {
	genStateIsSet := cliCtx.IsSet(flags.InteropGenesisStateFlag.Name)
	genTimeIsSet := cliCtx.IsSet(flags.InteropGenesisTimeFlag.Name)
	numValsIsSet := cliCtx.IsSet(flags.InteropNumValidatorsFlag.Name)
	votesIsSet := cliCtx.IsSet(flags.InteropMockEth1DataVotesFlag.Name)

	if genStateIsSet || genTimeIsSet || numValsIsSet || votesIsSet {
		bCfg := params.BeaconConfig()
		bCfg.ConfigName = "interop"
		params.OverrideBeaconConfig(bCfg)
	}
}

func configureDataConfig(cliCtx *cli.Context) {
	bCfg := params.BeaconConfig()
	bCfg.DataDir = cliCtx.String(cmd.DataDirFlag.Name)
	params.OverrideBeaconConfig(bCfg)
}

func configureRewardLogConfig(cliCtx *cli.Context) {
	bCfg := params.BeaconConfig()
	bCfg.WriteRewardLogFlag = cliCtx.Bool(cmd.WriteRewardLogFlag.Name)
	params.OverrideBeaconConfig(bCfg)
}

func configureExecutionSetting(cliCtx *cli.Context) error {
	if !cliCtx.IsSet(flags.SuggestedFeeRecipient.Name) {
		return nil
	}

	c := params.BeaconConfig()
	ha := cliCtx.String(flags.SuggestedFeeRecipient.Name)
	if !common.IsHexAddress(ha) {
		return fmt.Errorf("%s is not a valid fee recipient address", ha)
	}
	c.DefaultFeeRecipient = common.HexToAddress(ha)
	params.OverrideBeaconConfig(c)
	return nil
}

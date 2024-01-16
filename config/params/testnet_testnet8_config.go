package params

import (
	types "github.com/prysmaticlabs/eth2-types"
	gwatParams "gitlab.waterfall.network/waterfall/protocol/gwat/params"
)

// UseTestnet8NetworkConfig uses the Testnet8 specific network config.
func UseTestnet8NetworkConfig() {
	cfg := BeaconNetworkConfig().Copy()
	cfg.ContractDeploymentBlock = 0
	cfg.BootstrapNodes = []string{
		// Prysm's bootnode
		"enr:-LG4QC0DoIv8bWBuE_ZVx9zcrDaE1HbBPuNWVpl74GoStnSPXO0B73WF5VlfDJQSqTetQ775V9PWi7Yg3Ua7igL1ucOGAYvKzQeKh2F0" +
			"dG5ldHOIAAAAAAAAAACEZXRoMpATfFTTAAAgCf__________gmlkgnY0gmlwhICMLZGJc2VjcDI1NmsxoQKi0xTSOgGw6UO9URJjAM1T" +
			"PqPfadDeuORaJ027WIjLYIN1ZHCCD6A",
	}
	OverrideBeaconNetworkConfig(cfg)
}

// UseTestnet8Config sets the main beacon chain config for Testnet8.
func UseTestnet8Config() {
	beaconConfig = Testnet8Config()
}

// Testnet8Config defines the config for the Testnet8.
func Testnet8Config() *BeaconChainConfig {
	cfg := MainnetConfig().Copy()
	cfg.ConfigName = ConfigNames[Testnet8]
	cfg.DepositContractAddress = "0x6671Ed1732b6b5AF82724A1d1A94732D1AA37aa6"
	cfg.DepositChainID = gwatParams.Testnet8ChainConfig.ChainID.Uint64()
	cfg.DepositNetworkID = gwatParams.Testnet8ChainConfig.ChainID.Uint64()
	cfg.DelegateForkSlot = types.Slot(gwatParams.Testnet8ChainConfig.ForkSlotDelegate)
	cfg.SlotsPerArchivedPoint = 2048
	cfg.InitializeForkSchedule()
	return cfg
}

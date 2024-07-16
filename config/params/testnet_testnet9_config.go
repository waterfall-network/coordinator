package params

import "math"

// UseTestnet9NetworkConfig uses the Testnet9 specific network config.
func UseTestnet9NetworkConfig() {
	cfg := BeaconNetworkConfig().Copy()
	cfg.BootstrapNodes = []string{}
	OverrideBeaconNetworkConfig(cfg)
}

// UseTestnet9Config sets the main beacon chain config for Testnet9.
func UseTestnet9Config() {
	beaconConfig = Testnet8Config()
}

// Testnet9Config defines the config for the Testnet9.
func Testnet9Config() *BeaconChainConfig {
	cfg := MainnetConfig().Copy()

	cfg.ConfigName = ConfigNames[Testnet9]
	//cfg.DepositContractAddress = "0x6671Ed1732b6b5AF82724A1d1A94732D1AA37aa6"
	cfg.DepositChainID = 1501869
	cfg.DepositNetworkID = 1501869
	cfg.DelegateForkSlot = 0
	cfg.DelegateForkSlot = 0
	cfg.PrefixFinForkSlot = 0
	cfg.FinEth1ForkSlot = 0
	cfg.BlockVotingForkSlot = math.MaxUint64
	//cfg.SlotsPerArchivedPoint = 2048

	cfg.SlotsPerEpoch = 32
	cfg.SecondsPerSlot = 6

	//cfg.MinDepositAmount = 1000 * 1e9
	//cfg.MaxEffectiveBalance = 32000 * 1e9
	//cfg.EjectionBalance = 16000 * 1e9
	//cfg.EffectiveBalanceIncrement = 1000 * 1e9
	//cfg.OptValidatorsNum = 300_000

	cfg.InitializeForkSchedule()
	return cfg
}

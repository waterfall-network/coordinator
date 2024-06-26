package params

// UseTestConfig sets the main beacon chain config for tests.
func UseTestConfig() {
	beaconConfig = TestConfig()
}

// TestConfig defines the config for the tests.
func TestConfig() *BeaconChainConfig {
	cfg := MainnetConfig().Copy()
	cfg.ConfigName = ConfigNames[Testnet8]
	cfg.PresetBase = "mainnet"

	cfg.MinDepositAmount = 100 * 1e9
	cfg.MaxEffectiveBalance = 3200 * 1e9
	cfg.EjectionBalance = 1600 * 1e9
	cfg.EffectiveBalanceIncrement = 100 * 1e9
	cfg.OptValidatorsNum = 3_000_000

	cfg.SecondsPerSlot = 4

	cfg.InitializeForkSchedule()

	return cfg
}

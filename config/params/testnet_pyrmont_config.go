package params

import "math"

// UsePyrmontNetworkConfig uses the Pyrmont specific
// network config.
func UsePyrmontNetworkConfig() {
	cfg := BeaconNetworkConfig().Copy()
	cfg.ContractDeploymentBlock = 0
	cfg.BootstrapNodes = []string{
		//"enr:-Ku4QOA5OGWObY8ep_x35NlGBEj7IuQULTjkgxC_0G1AszqGEA0Wn2RNlyLFx9zGTNB1gdFBA6ZDYxCgIza1uJUUOj4Dh2F0dG5ldHOIAAAAAAAAAACEZXRoMpDVTPWXAAAgCf__________gmlkgnY0gmlwhDQPSjiJc2VjcDI1NmsxoQM6yTQB6XGWYJbI7NZFBjp4Yb9AYKQPBhVrfUclQUobb4N1ZHCCIyg",
		//"enr:-Ku4QOksdA2tabOGrfOOr6NynThMoio6Ggka2oDPqUuFeWCqcRM2alNb8778O_5bK95p3EFt0cngTUXm2H7o1jkSJ_8Dh2F0dG5ldHOIAAAAAAAAAACEZXRoMpDVTPWXAAAgCf__________gmlkgnY0gmlwhDaa13aJc2VjcDI1NmsxoQKdNQJvnohpf0VO0ZYCAJxGjT0uwJoAHbAiBMujGjK0SoN1ZHCCIyg",
	}
	OverrideBeaconNetworkConfig(cfg)
}

// UsePyrmontConfig sets the main beacon chain
// config for Pyrmont.
func UsePyrmontConfig() {
	beaconConfig = PyrmontConfig()
}

// PyrmontConfig defines the config for the
// Pyrmont testnet.
func PyrmontConfig() *BeaconChainConfig {
	cfg := MainnetConfig().Copy()
	cfg.MinGenesisTime = 1605700800
	cfg.GenesisDelay = 432000
	cfg.ConfigName = ConfigNames[Pyrmont]
	cfg.GenesisForkVersion = []byte{0x00, 0x00, 0x20, 0x09}
	cfg.AltairForkVersion = []byte{0x01, 0x00, 0x20, 0x09}
	cfg.AltairForkEpoch = 1
	cfg.BellatrixForkVersion = []byte{0x02, 0x00, 0x20, 0x09}
	cfg.BellatrixForkEpoch = math.MaxUint64
	cfg.ShardingForkVersion = []byte{0x03, 0x00, 0x20, 0x09}
	cfg.ShardingForkEpoch = math.MaxUint64
	cfg.SecondsPerETH1Block = 14
	cfg.DepositChainID = 337733
	cfg.DepositNetworkID = 337733
	cfg.DepositContractAddress = "0xf30097f8c858c1f6b0c6efe72240319efa65b825"
	cfg.InitializeForkSchedule()
	return cfg
}

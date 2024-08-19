package params

import (
	"math"
	"time"

	"gitlab.waterfall.network/waterfall/protocol/coordinator/encoding/bytesutil"
)

// MainnetConfig returns the configuration to be used in the main network.
func MainnetConfig() *BeaconChainConfig {
	if mainnetBeaconConfig.ForkVersionSchedule == nil {
		mainnetBeaconConfig.InitializeForkSchedule()
	}
	return mainnetBeaconConfig
}

// UseMainnetConfig for beacon chain services.
func UseMainnetConfig() {
	beaconConfig = MainnetConfig()
}

const (
	// Genesis Fork Epoch for the mainnet config.
	genesisForkEpoch = 0
	// Altair Fork Epoch for mainnet config.
	mainnetAltairForkEpoch = 1
	// Placeholder for the merge epoch until it is decided
	mainnetBellatrixForkEpoch = math.MaxUint64
)

var mainnetNetworkConfig = &NetworkConfig{
	GossipMaxSize:                   1 << 20,      // 1 MiB
	GossipMaxSizeBellatrix:          10 * 1 << 20, // 10 MiB
	MaxChunkSize:                    1 << 20,      // 1 MiB
	MaxChunkSizeBellatrix:           10 * 1 << 20, // 10 MiB
	AttestationSubnetCount:          64,
	AttestationPropagationSlotRange: 32,
	MaxRequestBlocks:                1 << 10, // 1024
	TtfbTimeout:                     5 * time.Second,
	RespTimeout:                     10 * time.Second,
	MaximumGossipClockDisparity:     500 * time.Millisecond,
	MessageDomainInvalidSnappy:      [4]byte{00, 00, 00, 00},
	MessageDomainValidSnappy:        [4]byte{01, 00, 00, 00},
	ETH2Key:                         "eth2",
	AttSubnetKey:                    "attnets",
	SyncCommsSubnetKey:              "syncnets",
	MinimumPeersInSubnetSearch:      20,
	BootstrapNodes: []string{
		"enr:-LG4QAGJyiJYWVjQnJ2ANfWE_AtbYnYEVcYS3k5iyUaALWKBI7OL30dc_-Nxigt7FpiB4b0cfmq62iGXB76C8BmRVGeGAZBd0Ihph2F0dG5ldHOIAAAAAAAAAACEZXRoMpB5pkc4AAAgCf__________gmlkgnY0gmlwhCImhDKJc2VjcDI1NmsxoQND0I1D6IIk-kwev1LftepaWrPOyN3pgkTbDrHfJN0bGIN1ZHCCD6A",
		"enr:-LG4QLMIXca_9nOcCfvGrO_214_l5lSWAchwHZbfFwLDGY-dc0JXVqRb2-mSjlQ13A8K5ztYYyhoyk6Iu1AIBwMYPLCGAZCwyG9xh2F0dG5ldHOIAAAAAAAAAACEZXRoMpB5pkc4AAAgCf__________gmlkgnY0gmlwhKUWFmeJc2VjcDI1NmsxoQMjY_D90oGB2dLQXEFQo3-ait2innKYEQdS4cU2D5wY4YN1ZHCCD6E",
		"enr:-LG4QKPw-TJftwQSVV1pvbD-tUegL1rdK09fLWy5a6a1KZenDJhStaH1Y1Qsb_nMEs-cu6WBBqVe58d5g86biTYXvW-GAZCwyGqgh2F0dG5ldHOIAAAAAAAAAACEZXRoMpB5pkc4AAAgCf__________gmlkgnY0gmlwhES3xj6Jc2VjcDI1NmsxoQJ-s2l20ud2p830t0hOrNkGoTPrqeZwOBoxcjtBE9sihoN1ZHCCD6E",
		"enr:-LG4QLRe_wvUmKSV8u1TSXzZazhX9NPxtQNkt0Av9-lKervWPbrkirBxdd7-TT3Mex4_dpW--yIBnskXslsQ3_3xJzyGAZCwyE8rh2F0dG5ldHOIAAAAAAAAAACEZXRoMpB5pkc4AAAgCf__________gmlkgnY0gmlwhI9uuGWJc2VjcDI1NmsxoQP92cFXNZPEy-X5R492RnmxwPWCQaoqBJ3TTR7hvkXZyYN1ZHCCD6E",
		"enr:-LG4QNKJi6AAiz9ZZ1N5YWkoUWPydtHd8hVq6JSgtdAcvKBgaUg9zny_2QWG3-mv9f6xMg8fa4zr4ggX2SSOE3EI67WGAZCwx2aPh2F0dG5ldHOIAAAAAAAAAACEZXRoMpB5pkc4AAAgCf__________gmlkgnY0gmlwhE4vZPKJc2VjcDI1NmsxoQJdpDQ7ZzJhnrxqMyy9atE9SyOv8kyogXHHqdpt1FJ49YN1ZHCCD6E",
		"enr:-LG4QH_6bcF-WqgU8CE8g-o0JRzQtom6IWiYOk-dMi471oPjMtaJUwyt3ReSjfYMoOh6Npl9GxsyCXfuOhMxxiehofmGAZCwx20Vh2F0dG5ldHOIAAAAAAAAAACEZXRoMpB5pkc4AAAgCf__________gmlkgnY0gmlwhEFs2WOJc2VjcDI1NmsxoQL7ZxxxxRu-HSCwA-C-aeCDGiPD5qZfoipv8fEi_4QmhoN1ZHCCD6E",
		"enr:-LG4QGuEEWi3jTOcPUUwQiNsPSt6u-zJO455nI7nIpwl4KMUTV_pOnsxbLRgU2uI6fdgrjTpgNb6EbAjjTtS4-rmbaGGAZCwx82xh2F0dG5ldHOIAAAAAAAAAACEZXRoMpB5pkc4AAAgCf__________gmlkgnY0gmlwhCIm6eWJc2VjcDI1NmsxoQLdlMpIS6ndVia90J8V0MNAzQcvEHHTvuYvgtGC-x55pYN1ZHCCD6E",
		"enr:-LG4QNbopI8yniBEKA88rQnM_jch8eML3CwIedYCcGj_ATIiWWl1B6ecncmv-UsX4FgJVav55O_vFgfU95jAHLNq75OGAZCwyAfwh2F0dG5ldHOIAAAAAAAAAACEZXRoMpB5pkc4AAAgCf__________gmlkgnY0gmlwhCIvBAKJc2VjcDI1NmsxoQOLDdXK1aerAL7oM2NFFkXWATPiwh8LJVygJL1jvoKw9YN1ZHCCD6E",
		"enr:-LG4QHdSweICNDUEsTXxDmjhc9ZZolXyJzcv6F547G_3u5uhCYikm2xpWay1lnva_ifMZSWuioB_UJta8w9PUVHz6HCGAZCwyFHQh2F0dG5ldHOIAAAAAAAAAACEZXRoMpB5pkc4AAAgCf__________gmlkgnY0gmlwhCKXxmqJc2VjcDI1NmsxoQIpKtnJ2gkRjdWljBijoBSKtCFCu1LqjLWuNkU3GMtK9YN1ZHCCD6E",
		"enr:-LG4QKMfgK7jhO4yQ7-CFUfP-yJbspkHRe1mibEvTkURHneeOmC7LZTHih8IoIgUpljxv6Y5yE6B66wA4wK4I1VfOZ2GAZCwyBr-h2F0dG5ldHOIAAAAAAAAAACEZXRoMpB5pkc4AAAgCf__________gmlkgnY0gmlwhCISE0uJc2VjcDI1NmsxoQKYswnQXlZp_1qAk7oioBgIeBppLF1L4ZLvEglS69aUbIN1ZHCCD6E",
		"enr:-LG4QK419449i2AXNVHLYqQff4yUToPIDUpWG8cLFZvk8hqVSGxdPCfv8YlXzfHn21mWAF1Z5kPVFMqrDuD1WufEsCOGAZCwx_p0h2F0dG5ldHOIAAAAAAAAAACEZXRoMpB5pkc4AAAgCf__________gmlkgnY0gmlwhCKlMHKJc2VjcDI1NmsxoQPNclITf31-j4j9zvCJZ-trVoIh4rzbLAs1-fsryt1XuoN1ZHCCD6E",
		"enr:-LG4QMwbcNwYQKAsdjHpGeJiEZxhCNukA6O1SO0MyITl1OgPN-lkI2hsoYtcKe0vBi0V-rI7xvzuC3Fp9lGBfyBRfuCGAZCwyA1th2F0dG5ldHOIAAAAAAAAAACEZXRoMpB5pkc4AAAgCf__________gmlkgnY0gmlwhCIjOSyJc2VjcDI1NmsxoQNFm80FFjGtkgdfVIzYFwW0Ihybl-LQ9c2eC2ZBXEZWtIN1ZHCCD6E",
	},
}

var mainnetBeaconConfig = &BeaconChainConfig{
	// Constants (Non-configurable)
	FarFutureEpoch:           math.MaxUint64,
	FarFutureSlot:            math.MaxUint64,
	BaseRewardsPerEpoch:      4,
	DepositContractTreeDepth: 32,
	GenesisDelay:             300, // 5 min.

	// Misc constant.
	TargetCommitteeSize:            128,
	MaxValidatorsPerCommittee:      2048,
	MaxCommitteesPerSlot:           64,
	MinPerEpochChurnLimit:          4,
	ChurnLimitQuotient:             1 << 16,
	ShuffleRoundCount:              90,
	MinGenesisActiveValidatorCount: 16384,
	MinGenesisTime:                 1606824000, // Dec 1, 2020, 12pm UTC.
	TargetAggregatorsPerCommittee:  16,
	HysteresisQuotient:             4,
	HysteresisDownwardMultiplier:   1,
	HysteresisUpwardMultiplier:     5,

	// Gwei value constants.
	MinDepositAmount:          1000 * 1e9,
	MaxEffectiveBalance:       32000 * 1e9,
	EjectionBalance:           16000 * 1e9,
	EffectiveBalanceIncrement: 1000 * 1e9,

	// Initial value constants.
	BLSWithdrawalPrefixByte: byte(0),
	ZeroHash:                [32]byte{},

	// Time parameter constants.
	MinAttestationInclusionDelay: 1,
	//SecondsPerSlot:               12,
	SecondsPerSlot:       6,
	SlotsPerEpoch:        32,
	SqrRootSlotsPerEpoch: 5,
	MinSeedLookahead:     1,
	MaxSeedLookahead:     4,
	//EpochsPerEth1VotingPeriod:        64,
	EpochsPerEth1VotingPeriod:        4,
	SlotsPerHistoricalRoot:           8192,
	WithdrawalBalanceLockPeriod:      4,
	MinValidatorWithdrawabilityDelay: 4, // orig val: 256
	ShardCommitteePeriod:             4, // orig val: 256:epochs a validator must participate before exiting.
	MinEpochsToInactivityPenalty:     4,
	//Eth1FollowDistance:               2048,
	Eth1FollowDistance: 16,
	//Eth1FollowDistance:         64,
	SafeSlotsToUpdateJustified: 8,

	//CleanWithdrawalsAftEpochs: 100,
	CleanWithdrawalsAftEpochs: 100,

	//Optimistic consensus constants.
	VotingRequiredSlots:           3,
	BlockVotingMinSupportPrc:      50,
	SpinePublicationsPefixSupport: 2,

	// Fork choice algorithm constants.
	ProposerScoreBoost: 70,
	IntervalsPerSlot:   3,

	// Shard node mainnet settings.
	DepositChainID:         181, // Chain ID of eth1 mainnet.
	DepositNetworkID:       181, // Network ID of eth1 mainnet.
	DepositContractAddress: "0x329c3A3d65Ab0bE08c6eff6695933391Cfc02cCA",

	// Validator params.
	RandomSubnetsPerValidator:         1 << 0,
	EpochsPerRandomSubnetSubscription: 1 << 8,

	// While eth1 mainnet block times are closer to 13s, we must conform with other clients in
	// order to vote on the correct eth1 blocks.
	//
	// Additional context: https://github.com/ethereum/consensus-specs/issues/2132
	// Bug prompting this change: https://github.com/prysmaticlabs/prysm/issues/7856
	// Future optimization: https://github.com/prysmaticlabs/prysm/issues/7739
	//SecondsPerETH1Block: 14,
	SecondsPerETH1Block: 4,

	GwatSyncIntervalMs: 1000,

	// State list length constants.
	EpochsPerHistoricalVector: 65536,
	EpochsPerSlashingsVector:  8192,
	HistoricalRootsLimit:      16777216,
	ValidatorRegistryLimit:    1099511627776,
	WithdrawalOpsLimit:        1024,
	AllSpinesLimit:            128,

	// Reward and penalty quotients constants.
	BaseRewardFactor:               64,
	WhistleBlowerRewardQuotient:    512,
	ProposerRewardQuotient:         8,
	InactivityPenaltyQuotient:      67108864,
	MinSlashingPenaltyQuotient:     128,
	ProportionalSlashingMultiplier: 1,
	BaseRewardMultiplier:           2.0,
	MaxAnnualizedReturnRate:        0.2,
	OptValidatorsNum:               300_000,

	// Max operations per block constants.
	MaxProposerSlashings: 16,
	MaxAttesterSlashings: 2,
	MaxAttestations:      128,
	MaxDeposits:          16,
	MaxVoluntaryExits:    16,
	MaxWithdrawals:       1024,

	// BLS domain values.
	DomainBeaconProposer:              bytesutil.ToBytes4(bytesutil.Bytes4(0)),
	DomainBeaconAttester:              bytesutil.ToBytes4(bytesutil.Bytes4(1)),
	DomainRandao:                      bytesutil.ToBytes4(bytesutil.Bytes4(2)),
	DomainDeposit:                     bytesutil.ToBytes4(bytesutil.Bytes4(3)),
	DomainVoluntaryExit:               bytesutil.ToBytes4(bytesutil.Bytes4(4)),
	DomainSelectionProof:              bytesutil.ToBytes4(bytesutil.Bytes4(5)),
	DomainAggregateAndProof:           bytesutil.ToBytes4(bytesutil.Bytes4(6)),
	DomainSyncCommittee:               bytesutil.ToBytes4(bytesutil.Bytes4(7)),
	DomainSyncCommitteeSelectionProof: bytesutil.ToBytes4(bytesutil.Bytes4(8)),
	DomainContributionAndProof:        bytesutil.ToBytes4(bytesutil.Bytes4(9)),

	// Prysm constants.
	GweiPerEth:                     1000000000,
	BLSSecretKeyLength:             32,
	BLSPubkeyLength:                48,
	DefaultBufferSize:              10000,
	WithdrawalPrivkeyFileName:      "/shardwithdrawalkey",
	ValidatorPrivkeyFileName:       "/validatorprivatekey",
	RPCSyncCheck:                   1,
	EmptySignature:                 [96]byte{},
	DefaultPageSize:                250,
	MaxPeersToSync:                 15,
	SlotsPerArchivedPoint:          2048,
	GenesisCountdownInterval:       time.Minute,
	ConfigName:                     ConfigNames[Mainnet],
	PresetBase:                     "mainnet",
	BeaconStateFieldCount:          21 + 2,
	BeaconStateAltairFieldCount:    24 + 2,
	BeaconStateBellatrixFieldCount: 25 + 2,
	CtxBlockFetcherKey:             CtxFnKey("CtxBlockFetcher"),

	// Slasher related values.
	WeakSubjectivityPeriod:          54000,
	PruneSlasherStoragePeriod:       10,
	SlashingProtectionPruningEpochs: 512,

	// Weak subjectivity values.
	SafetyDecay: 10,

	// Fork related values.
	GenesisEpoch:         genesisForkEpoch,
	GenesisForkVersion:   []byte{0, 0, 0, 0},
	AltairForkVersion:    []byte{1, 0, 0, 0},
	AltairForkEpoch:      mainnetAltairForkEpoch,
	DelegateForkSlot:     0,
	PrefixFinForkSlot:    0,
	FinEth1ForkSlot:      0,
	BlockVotingForkSlot:  216000,
	BellatrixForkVersion: []byte{2, 0, 0, 0},
	BellatrixForkEpoch:   mainnetBellatrixForkEpoch,
	ShardingForkVersion:  []byte{3, 0, 0, 0},
	ShardingForkEpoch:    math.MaxUint64,

	// New values introduced in Altair hard fork 1.
	// Participation flag indices.
	TimelySourceFlagIndex: 0,
	TimelyTargetFlagIndex: 1,
	TimelyHeadFlagIndex:   2,

	// DAG Participation flag indices.
	DAGTimelyVotingFlagIndex: 3,

	// Incentivization weight values.
	TimelySourceWeight: 14,
	TimelyTargetWeight: 26,
	TimelyHeadWeight:   14,
	SyncRewardWeight:   2,
	ProposerWeight:     8,
	WeightDenominator:  64,

	// DAG Incentivization weight values.
	DAGTimelySourceWeight: 0.25,
	DAGTimelyTargetWeight: 0.25,
	DAGTimelyHeadWeight:   0.25,
	DAGTimelyVotingWeight: 0.25,

	// Validator related values.
	TargetAggregatorsPerSyncSubcommittee: 16,
	SyncCommitteeSubnetCount:             4,

	// Misc values.
	SyncCommitteeSize:            512,
	InactivityScoreBias:          4,
	InactivityScoreRecoveryRate:  16,
	EpochsPerSyncCommitteePeriod: 256,

	// Updated penalty values.
	InactivityPenaltyQuotientAltair:         3 * 1 << 24, //50331648
	MinSlashingPenaltyQuotientAltair:        64,
	ProportionalSlashingMultiplierAltair:    2,
	MinSlashingPenaltyQuotientBellatrix:     32,
	ProportionalSlashingMultiplierBellatrix: 3,

	// Light client
	MinSyncCommitteeParticipants: 1,

	// Bellatrix
	TerminalBlockHashActivationEpoch: 18446744073709551615,
	TerminalBlockHash:                [32]byte{},
	TerminalTotalDifficulty:          "115792089237316195423570985008687907853269984665640564039457584007913129638912",
}

package features

import (
	"time"

	"github.com/urfave/cli/v2"
)

var (
	// PyrmontTestnet flag for the multiclient waterfall consensus testnet.
	PyrmontTestnet = &cli.BoolFlag{
		Name:  "pyrmont",
		Usage: "This defines the flag through which we can run on the Pyrmont Multiclient Testnet",
	}
	// Testnet5 flag to override node configuration.
	Testnet5 = &cli.BoolFlag{
		Name:  "testnet5",
		Usage: "Override node configuration for the testnet5 network",
	}
	// Testnet9 flag to override node configuration.
	Testnet9 = &cli.BoolFlag{
		Name:  "testnet9",
		Usage: "Override node configuration for the testnet9 network",
	}
	// Testnet8 flag for the multiclient waterfall consensus testnet.
	Testnet8 = &cli.BoolFlag{
		Name:  "testnet8",
		Usage: "Run node configured for the Testnet8 test network",
	}
	// Mainnet flag for easier tooling, no-op
	Mainnet = &cli.BoolFlag{
		Value: true,
		Name:  "mainnet",
		Usage: "Run on Waterfall Beacon Chain Main Net. This is the default and can be omitted.",
	}
	devModeFlag = &cli.BoolFlag{
		Name:  "dev",
		Usage: "Enable experimental features still in development. These features may not be stable.",
	}
	writeSSZStateTransitionsFlag = &cli.BoolFlag{
		Name:  "interop-write-ssz-state-transitions",
		Usage: "Write ssz states to disk after attempted state transition",
	}
	enableExternalSlasherProtectionFlag = &cli.BoolFlag{
		Name: "enable-external-slasher-protection",
		Usage: "Enables the validator to connect to a beacon node using the --slasher flag" +
			"for remote slashing protection",
	}
	disableGRPCConnectionLogging = &cli.BoolFlag{
		Name:  "disable-grpc-connection-logging",
		Usage: "Disables displaying logs for newly connected grpc clients",
	}
	enablePeerScorer = &cli.BoolFlag{
		Name:  "enable-peer-scorer",
		Usage: "Enable experimental P2P peer scorer",
	}
	enablePassSlotInfoToGwat = &cli.BoolFlag{
		Name:  "enable-pass-slot-info-to-gwat",
		Usage: "Enables passing slot info to GWAT during sync process",
	}
	checkPtInfoCache = &cli.BoolFlag{
		Name:  "use-check-point-cache",
		Usage: "Enables check point info caching",
	}
	enableLargerGossipHistory = &cli.BoolFlag{
		Name:  "enable-larger-gossip-history",
		Usage: "Enables the node to store a larger amount of gossip messages in its cache.",
	}
	disablePeerScorer = &cli.BoolFlag{
		Name:  "disable-peer-scorer",
		Usage: "(Danger): Disables P2P peer scorer. Do NOT use this in production!",
	}
	writeWalletPasswordOnWebOnboarding = &cli.BoolFlag{
		Name: "write-wallet-password-on-web-onboarding",
		Usage: "(Danger): Writes the wallet password to the wallet directory on completing Prysm web onboarding. " +
			"We recommend against this flag unless you are an advanced user.",
	}
	disableAttestingHistoryDBCache = &cli.BoolFlag{
		Name: "disable-attesting-history-db-cache",
		Usage: "(Danger): Disables the cache for attesting history in the validator DB, greatly increasing " +
			"disk reads and writes as well as increasing time required for attestations to be produced",
	}
	dynamicKeyReloadDebounceInterval = &cli.DurationFlag{
		Name: "dynamic-key-reload-debounce-interval",
		Usage: "(Advanced): Specifies the time duration the validator waits to reload new keys if they have " +
			"changed on disk. Default 1s, can be any type of duration such as 1.5s, 1000ms, 1m.",
		Value: time.Second,
	}
	disableBroadcastSlashingFlag = &cli.BoolFlag{
		Name:  "disable-broadcast-slashings",
		Usage: "Disables broadcasting slashings submitted to the beacon node.",
	}
	attestTimely = &cli.BoolFlag{
		Name:  "attest-timely",
		Usage: "Fixes validator can attest timely after current block processes. See #8185 for more details",
	}
	enableSlasherFlag = &cli.BoolFlag{
		Name:  "slasher",
		Usage: "Enables a slasher in the beacon node for detecting slashable offenses",
	}
	disableProposerAttsSelectionUsingMaxCover = &cli.BoolFlag{
		Name:  "disable-proposer-atts-selection-using-max-cover",
		Usage: "Disable max-cover algorithm when selecting attestations for proposer",
	}
	enableSlashingProtectionPruning = &cli.BoolFlag{
		Name:  "enable-slashing-protection-history-pruning",
		Usage: "Enables the pruning of the validator client's slashing protection database",
	}
	disableOptimizedBalanceUpdate = &cli.BoolFlag{
		Name:  "disable-optimized-balance-update",
		Usage: "Disable the optimized method of updating validator balances.",
	}
	enableDoppelGangerProtection = &cli.BoolFlag{
		Name: "enable-doppelganger",
		Usage: "Enables the validator to perform a doppelganger check. (Warning): This is not " +
			"a foolproof method to find duplicate instances in the network. Your validator will still be" +
			" vulnerable if it is being run in unsafe configurations.",
	}
	enableHistoricalSpaceRepresentation = &cli.BoolFlag{
		Name: "enable-historical-state-representation",
		Usage: "Enables the beacon chain to save historical states in a space efficient manner." +
			" (Warning): Once enabled, this feature migrates your database in to a new schema and " +
			"there is no going back. At worst, your entire database might get corrupted.",
	}
	disableCorrectlyInsertOrphanedAtts = &cli.BoolFlag{
		Name: "disable-correctly-insert-orphaned-atts",
		Usage: "Disable the fix for bug where orphaned attestations don't get reinserted back to mem pool. Which is an improves validator profitability and overall network health," +
			"see issue #9441 for further detail",
	}
	disableCorrectlyPruneCanonicalAtts = &cli.BoolFlag{
		Name: "disable-correctly-prune-canonical-atts",
		Usage: "Disable the fix for bug where any block attestations can get incorrectly pruned, which improves validator profitability and overall network health," +
			"see issue #9443 for further detail",
	}
	disableActiveBalanceCache = &cli.BoolFlag{
		Name:  "disable-active-balance-cache",
		Usage: "This disables active balance cache, which improves node performance during block processing",
	}
	disableGetBlockOptimizations = &cli.BoolFlag{
		Name:  "disable-get-block-optimizations",
		Usage: "This disables some optimizations on the GetBlock() function.",
	}
	disableBatchGossipVerification = &cli.BoolFlag{
		Name:  "disable-batch-gossip-verification",
		Usage: "This enables batch verification of signatures received over gossip.",
	}
	disableBalanceTrieComputation = &cli.BoolFlag{
		Name:  "disable-balance-trie-computation",
		Usage: "This disables optimized hash tree root operations for our balance field.",
	}
	enableNativeState = &cli.BoolFlag{
		Name:  "enable-native-state",
		Usage: "Enables representing the beacon state as a pure Go struct.",
	}
	enableForkChoiceDoublyLinkedTree = &cli.BoolFlag{
		Name:  "enable-forkchoice-doubly-linked-tree",
		Usage: "Enables new forkchoice store structure that uses doubly linked trees. (Warning): This feature is still in development and temporary disabled by default.",
	}
)

// devModeFlags holds list of flags that are set when development mode is on.
var devModeFlags = []cli.Flag{
	enablePeerScorer,
	enableForkChoiceDoublyLinkedTree,
}

// ValidatorFlags contains a list of all the feature flags that apply to the validator client.
var ValidatorFlags = append(deprecatedFlags, []cli.Flag{
	writeWalletPasswordOnWebOnboarding,
	enableExternalSlasherProtectionFlag,
	disableAttestingHistoryDBCache,
	PyrmontTestnet,
	//Testnet5,
	//Testnet9,
	Testnet8,
	Mainnet,
	dynamicKeyReloadDebounceInterval,
	attestTimely,
	enableSlashingProtectionPruning,
	enableDoppelGangerProtection,
}...)

// E2EValidatorFlags contains a list of the validator feature flags to be tested in E2E.
var E2EValidatorFlags = []string{
	"--enable-doppelganger",
}

// BeaconChainFlags contains a list of all the feature flags that apply to the beacon-chain client.
var BeaconChainFlags = append(deprecatedFlags, []cli.Flag{
	devModeFlag,
	writeSSZStateTransitionsFlag,
	disableGRPCConnectionLogging,
	PyrmontTestnet,
	Testnet8,
	Testnet5,
	Testnet9,
	Mainnet,
	disablePeerScorer,
	enablePeerScorer,
	enableLargerGossipHistory,
	checkPtInfoCache,
	disableBroadcastSlashingFlag,
	enableSlasherFlag,
	disableProposerAttsSelectionUsingMaxCover,
	disableOptimizedBalanceUpdate,
	enableHistoricalSpaceRepresentation,
	disableCorrectlyInsertOrphanedAtts,
	disableGetBlockOptimizations,
	disableCorrectlyPruneCanonicalAtts,
	disableActiveBalanceCache,
	disableBatchGossipVerification,
	disableBalanceTrieComputation,
	enableNativeState,
	enableForkChoiceDoublyLinkedTree,
	enablePassSlotInfoToGwat,
}...)

// E2EBeaconChainFlags contains a list of the beacon chain feature flags to be tested in E2E.
var E2EBeaconChainFlags = []string{
	"--dev",
	"--use-check-point-cache",
	"--enable-active-balance-cache",
}

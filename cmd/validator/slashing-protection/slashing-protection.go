package historycmd

import (
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/cmd"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/cmd/validator/flags"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/config/features"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/runtime/tos"
)

// Commands for slashing protection.
var Commands = &cli.Command{
	Name:     "slashing-protection-history",
	Category: "slashing-protection-history",
	Usage:    "defines commands for interacting your validator's slashing protection history",
	Subcommands: []*cli.Command{
		{
			Name:        "export",
			Description: `exports your validator slashing protection history into an EIP-3076 compliant JSON`,
			Flags: cmd.WrapFlags([]cli.Flag{
				cmd.DataDirFlag,
				flags.SlashingProtectionExportDirFlag,
				features.Mainnet,
				features.PyrmontTestnet,
				features.Testnet8,
				features.Testnet5,
				features.Testnet9,
				cmd.AcceptTosFlag,
			}),
			Before: func(cliCtx *cli.Context) error {
				if err := cmd.LoadFlagsFromConfig(cliCtx, cliCtx.Command.Flags); err != nil {
					return err
				}
				return tos.VerifyTosAcceptedOrPrompt(cliCtx)
			},
			Action: func(cliCtx *cli.Context) error {
				features.ConfigureValidator(cliCtx)
				if err := exportSlashingProtectionJSON(cliCtx); err != nil {
					logrus.Fatalf("Could not export slashing protection file: %v", err)
				}
				return nil
			},
		},
		{
			Name:        "import",
			Description: `imports a selected EIP-3076 compliant slashing protection JSON to the validator database`,
			Flags: cmd.WrapFlags([]cli.Flag{
				cmd.DataDirFlag,
				flags.SlashingProtectionJSONFileFlag,
				features.Mainnet,
				features.PyrmontTestnet,
				features.Testnet8,
				features.Testnet5,
				features.Testnet9,
				cmd.AcceptTosFlag,
			}),
			Before: func(cliCtx *cli.Context) error {
				if err := cmd.LoadFlagsFromConfig(cliCtx, cliCtx.Command.Flags); err != nil {
					return err
				}
				return tos.VerifyTosAcceptedOrPrompt(cliCtx)
			},
			Action: func(cliCtx *cli.Context) error {
				features.ConfigureValidator(cliCtx)
				err := importSlashingProtectionJSON(cliCtx)
				if err != nil {
					logrus.Fatalf("Could not import slashing protection cli: %v", err)
				}
				return nil
			},
		},
	},
}

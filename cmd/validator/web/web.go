package web

import (
	"fmt"

	"github.com/urfave/cli/v2"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/cmd"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/cmd/validator/flags"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/config/features"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/runtime/tos"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/validator/rpc"
)

// Commands for managing Prysm validator accounts.
var Commands = &cli.Command{
	Name:     "web",
	Category: "web",
	Usage:    "defines commands for interacting with the Prysm web interface",
	Subcommands: []*cli.Command{
		{
			Name:        "generate-auth-token",
			Description: `Generate an authentication token for the Prysm web interface`,
			Flags: cmd.WrapFlags([]cli.Flag{
				flags.WalletDirFlag,
				flags.GRPCGatewayHost,
				flags.GRPCGatewayPort,
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
				walletDirPath := cliCtx.String(flags.WalletDirFlag.Name)
				if walletDirPath == "" {
					log.Fatal("--wallet-dir not specified")
				}
				gatewayHost := cliCtx.String(flags.GRPCGatewayHost.Name)
				gatewayPort := cliCtx.Int(flags.GRPCGatewayPort.Name)
				validatorWebAddr := fmt.Sprintf("%s:%d", gatewayHost, gatewayPort)
				if err := rpc.CreateAuthToken(walletDirPath, validatorWebAddr); err != nil {
					log.Fatalf("Could not create web auth token: %v", err)
				}
				return nil
			},
		},
	},
}

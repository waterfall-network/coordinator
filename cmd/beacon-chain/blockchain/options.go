package blockchaincmd

import (
	"github.com/urfave/cli/v2"
	"github.com/waterfall-foundation/coordinator/beacon-chain/blockchain"
	"github.com/waterfall-foundation/coordinator/beacon-chain/core/helpers"
	"github.com/waterfall-foundation/coordinator/cmd"
	"github.com/waterfall-foundation/coordinator/cmd/beacon-chain/flags"
)

// FlagOptions for blockchain service flag configurations.
func FlagOptions(c *cli.Context) ([]blockchain.Option, error) {
	wsp := c.String(flags.WeakSubjectivityCheckpoint.Name)
	wsCheckpt, err := helpers.ParseWeakSubjectivityInputString(wsp)
	if err != nil {
		return nil, err
	}
	maxRoutines := c.Int(cmd.MaxGoroutines.Name)
	opts := []blockchain.Option{
		blockchain.WithMaxGoroutines(maxRoutines),
		blockchain.WithWeakSubjectivityCheckpoint(wsCheckpt),
	}
	return opts, nil
}

package db

import (
	"os"
	"path"
	"strings"

	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/cmd"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/io/file"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/io/prompt"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/validator/db/kv"
)

const dbExistsYesNoPrompt = "A database file already exists in the target directory. " +
	"Are you sure that you want to overwrite it? [y/n]"

// Restore a Prysm validator database.
func Restore(cliCtx *cli.Context) error {
	sourceFile := cliCtx.String(cmd.RestoreSourceFileFlag.Name)
	targetDir := cliCtx.String(cmd.RestoreTargetDirFlag.Name)

	if file.FileExists(path.Join(targetDir, kv.ProtectionDbFileName)) {
		resp, err := prompt.ValidatePrompt(
			os.Stdin, dbExistsYesNoPrompt, prompt.ValidateYesOrNo,
		)
		if err != nil {
			return errors.Wrap(err, "could not validate choice")
		}
		if strings.EqualFold(resp, "n") {
			log.Info("Restore aborted")
			return nil
		}
	}
	if err := file.MkdirAll(targetDir); err != nil {
		return err
	}
	if err := file.CopyFile(sourceFile, path.Join(targetDir, kv.ProtectionDbFileName)); err != nil {
		return err
	}

	log.Info("Restore completed successfully")
	return nil
}

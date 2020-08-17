package cmd

import (
	"github.com/innobead/kubefire/internal/config"
	"github.com/innobead/kubefire/pkg/script"
	"github.com/spf13/cobra"
	"os/exec"
)

var InstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install prerequisites",
	RunE: func(cmd *cobra.Command, args []string) error {
		// download install script
		scripts := []script.Type{
			script.InstallPrerequisites,
		}

		for _, s := range scripts {
			if err := script.Download(s, config.TagVersion, true); err != nil {
				return err
			}

			if err := script.Run(s, config.TagVersion, createSetupInstallCommandEnvsFunc()); err != nil {
				return err
			}
		}

		return nil
	},
}

func createSetupInstallCommandEnvsFunc() func(cmd *exec.Cmd) error {
	return func(cmd *exec.Cmd) error {
		cmd.Env = append(
			cmd.Env,
			config.ExpectedPrerequisiteVersionsEnvVars()...,
		)

		return nil
	}
}

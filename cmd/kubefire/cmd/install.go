package cmd

import (
	"github.com/innobead/kubefire/internal/config"
	"github.com/innobead/kubefire/internal/info"
	"github.com/innobead/kubefire/pkg/script"
	"github.com/spf13/cobra"
	"os/exec"
)

var forceDownload bool

var InstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install prerequisites",
	RunE: func(cmd *cobra.Command, args []string) error {
		scripts := []script.Type{
			script.InstallPrerequisites,
		}

		for _, s := range scripts {
			if err := script.Download(s, config.TagVersion, forceDownload); err != nil {
				return err
			}

			if err := script.Run(s, config.TagVersion, CreateSetupInstallCommandEnvsFunc()); err != nil {
				return err
			}
		}

		return nil
	},
}

func init() {
	flags := InstallCmd.Flags()
	flags.BoolVar(&forceDownload, "force", false, "force to download")
}

func CreateSetupInstallCommandEnvsFunc() func(cmd *exec.Cmd) error {
	return func(cmd *exec.Cmd) error {
		cmd.Env = append(
			cmd.Env,
			info.CurrentRuntimeVersionInfo().ExpectedEnvVars()...,
		)

		return nil
	}
}

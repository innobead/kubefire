package cmd

import (
	"github.com/innobead/kubefire/internal/config"
	"github.com/innobead/kubefire/pkg/script"
	"github.com/spf13/cobra"
)

var UninstallCmd = &cobra.Command{
	Use:     "uninstall",
	Aliases: []string{"un"},
	Short:   "Uninstall prerequisites",
	PreRun: func(cmd *cobra.Command, args []string) {
		if !forceDownload {
			forceDownload = !config.IsReleasedTagVersion(config.TagVersion)
		}
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := script.Download(script.UninstallPrerequisites, config.TagVersion, forceDownload); err != nil {
			return err
		}

		if err := script.Run(script.UninstallPrerequisites, config.TagVersion, createSetupInstallCommandEnvsFunc()); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	flags := UninstallCmd.Flags()
	flags.BoolVar(&forceDownload, "force", false, "Force to uninstall")
}

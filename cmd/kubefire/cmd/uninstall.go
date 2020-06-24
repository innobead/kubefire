package cmd

import (
	"github.com/innobead/kubefire/internal/config"
	"github.com/innobead/kubefire/pkg/script"
	"github.com/spf13/cobra"
)

func init() {
	flags := UninstallCmd.Flags()
	flags.BoolVar(&forceDownload, "force", false, "force to download")
}

var UninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall prerequisites",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := script.Download(script.UninstallPrerequisites, config.TagVersion, forceDownload); err != nil {
			return err
		}

		if err := script.Run(script.UninstallPrerequisites, config.TagVersion); err != nil {
			return err
		}

		return nil
	},
}
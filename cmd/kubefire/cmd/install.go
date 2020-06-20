package cmd

import (
	"github.com/innobead/kubefire/internal/config"
	"github.com/innobead/kubefire/pkg/script"
	"github.com/spf13/cobra"
)

var InstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install prerequisites",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := script.Download(script.InstallPrerequisites, config.TagVersion, false); err != nil {
			return err
		}

		if err := script.Run(script.InstallPrerequisites, config.TagVersion); err != nil {
			return err
		}

		return nil
	},
}

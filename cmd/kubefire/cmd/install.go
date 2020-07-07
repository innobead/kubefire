package cmd

import (
	"github.com/innobead/kubefire/internal/config"
	"github.com/innobead/kubefire/pkg/script"
	"github.com/spf13/cobra"
)

var forceDownload bool

func init() {
	flags := InstallCmd.Flags()
	flags.BoolVar(&forceDownload, "force", false, "force to download")
}

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

			if err := script.Run(s, config.TagVersion); err != nil {
				return err
			}
		}

		return nil
	},
}

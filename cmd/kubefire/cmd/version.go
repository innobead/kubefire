package cmd

import (
	"fmt"
	"github.com/innobead/kubefire/internal/config"
	"github.com/spf13/cobra"
)

var VersionCmd = &cobra.Command{
	Use:     "version",
	Aliases: []string{"v"},
	Short:   "Show version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Version: %s\nBuild: %s\n", config.TagVersion, config.BuildVersion)
	},
}

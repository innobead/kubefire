package cmd

import (
	"fmt"
	"github.com/innobead/kubefire/internal/config"
	"github.com/spf13/cobra"
)

var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Version: %s\n", config.BuildVersion)
	},
}

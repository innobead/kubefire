package main

import (
	"github.com/innobead/kubefire/cmd/kubefire/cmd"
	"github.com/innobead/kubefire/cmd/kubefire/cmd/cluster"
	"github.com/innobead/kubefire/internal/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.PersistentFlags().StringVar(&config.LogLevel, "log-level", string(logrus.InfoLevel), "Log level")
	rootCmd.PersistentFlags().StringVar(&config.Output, "output", "", "Output format (ex: json)")
}

var rootCmd = &cobra.Command{
	Use:   "kubefire",
	Short: "KubeFire, manage Kubernetes clusters on FireCracker microVMs",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func main() {
	cmds := []*cobra.Command{
		cmd.InstallCmd,
		cmd.VersionCmd,
		cluster.Cmd,
	}

	for _, c := range cmds {
		rootCmd.AddCommand(c)
	}

	if err := rootCmd.Execute(); err != nil {
		logrus.WithError(err).Fatal("Failed to run command")
	}
}

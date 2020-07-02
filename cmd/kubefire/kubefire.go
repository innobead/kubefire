package main

import (
	"github.com/innobead/kubefire/cmd/kubefire/cmd"
	"github.com/innobead/kubefire/cmd/kubefire/cmd/cluster"
	"github.com/innobead/kubefire/cmd/kubefire/cmd/node"
	"github.com/innobead/kubefire/internal/config"
	"github.com/innobead/kubefire/pkg/output"
	"github.com/innobead/kubefire/pkg/util"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:           "kubefire",
	Short:         "KubeFire, manage Kubernetes clusters on FireCracker microVMs",
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&config.LogLevel, "log-level", logrus.InfoLevel.String(), util.FlagsValuesUsage("log level", logrus.AllLevels))
	rootCmd.PersistentFlags().StringVar(&config.Output, "output", string(output.DEFAULT), util.FlagsValuesUsage("output format", output.BuiltinTypes))
}

func initConfig() {
	level, _ := logrus.ParseLevel(config.LogLevel)
	logrus.SetLevel(level)
}

func main() {
	cmds := []*cobra.Command{
		cmd.VersionCmd,
		cmd.InstallCmd,
		cmd.UninstallCmd,
		cluster.Cmd,
		node.Cmd,
	}

	for _, c := range cmds {
		rootCmd.AddCommand(c)
	}

	if err := rootCmd.Execute(); err != nil {
		logrus.Tracef("%+v", err)

		logrus.WithError(err).Fatalf("failed to run kubefire")
	}
}

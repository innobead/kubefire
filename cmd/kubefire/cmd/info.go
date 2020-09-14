package cmd

import (
	intcmd "github.com/innobead/kubefire/internal/cmd"
	"github.com/innobead/kubefire/internal/di"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var showBootstrapperInfo bool

var InfoCmd = &cobra.Command{
	Use:     "info",
	Aliases: []string{"i"},
	Short:   "Show info of prerequisites, supported K8s/K3s versions",
	PreRun: func(cmd *cobra.Command, args []string) {
		logrus.SetLevel(logrus.ErrorLevel)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if showBootstrapperInfo {
			if err := di.Output().Print(intcmd.BootstrapperVersionInfos(), nil, ""); err != nil {
				return errors.WithMessage(err, "failed to print output of bootstrapper info")
			}

			return nil
		}

		versionsInfo := intcmd.CurrentPrerequisitesInfos()
		if err := di.Output().Print(versionsInfo, nil, ""); err != nil {
			return errors.WithMessage(err, "failed to print output of prerequisites info")
		}

		return nil
	},
}

func init() {
	intcmd.AddOutputFlag(InfoCmd)

	flags := InfoCmd.Flags()
	flags.BoolVarP(&showBootstrapperInfo, "bootstrapper", "b", false, "Show K8s/K3s supported versions in builtin bootstrappers")
}

package cmd

import (
	intcmd "github.com/innobead/kubefire/internal/cmd"
	"github.com/innobead/kubefire/internal/di"
	"github.com/innobead/kubefire/pkg/bootstrap"
	pkgconfig "github.com/innobead/kubefire/pkg/config"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	showBootstrapperInfo bool
	noCache              bool
)

var InfoCmd = &cobra.Command{
	Use:     "info",
	Aliases: []string{"i"},
	Short:   "Shows info of prerequisites, supported K8s/K3s versions",
	PreRun: func(cmd *cobra.Command, args []string) {
		logrus.SetLevel(logrus.ErrorLevel)

		if noCache {
			for _, t := range bootstrap.BuiltinTypes {
				_ = di.ConfigManager().DeleteBootstrapperVersions(pkgconfig.NewBootstrapperVersion(t, ""))
			}
		}
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
	flags.BoolVar(&noCache, "no-cache", false, "Forget caches")
}

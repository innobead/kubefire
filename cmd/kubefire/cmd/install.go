package cmd

import (
	"github.com/innobead/kubefire/pkg/bootstrap"
	"github.com/innobead/kubefire/pkg/util"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var check bool
var bootstrapper string

func init() {
	InstallCmd.Flags().BoolVar(&check, "check", false, "check the prerequisites ready")
	InstallCmd.Flags().StringVar(&bootstrapper, "bootstrapper", string(bootstrap.KUBEADM), util.FlagsValuesUsage("bootstrapper type", bootstrap.BuiltinTypes))
}

var InstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install the requirements",
	RunE: func(cmd *cobra.Command, args []string) error {
		installers := []bootstrap.Bootstrapper{
			bootstrap.NewKubeadmBootstrapper(),
			bootstrap.NewSkubaBootstrapper(),
		}

		switch {
		case check:
			for _, i := range installers {
				if err := i.(bootstrap.BootstrapperInstaller).CheckRequirements(); err != nil {
					return errors.WithMessagef(err, "failed to check requirements")
				}
			}

		default:
			for _, i := range installers {
				if err := i.(bootstrap.BootstrapperInstaller).InstallRequirements(); err != nil {
					return errors.WithMessagef(err, "failed to install requirements")
				}
			}
		}

		return nil
	},
}

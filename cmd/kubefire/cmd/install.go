package cmd

import (
	"fmt"
	"github.com/innobead/kubefire/pkg/bootstrap"
	"github.com/spf13/cobra"
)

var check bool
var bootstrapper string

func init() {
	InstallCmd.Flags().BoolVar(&check, "check", false, "CheckRequirements if the prerequisites are ready")
	InstallCmd.Flags().StringVar(&bootstrapper, "bootstrapper", string(bootstrap.KUBEADM), fmt.Sprintf("bootstrapper type. ex: %v", bootstrap.BuiltinTypes))
}

var InstallCmd = &cobra.Command{
	Use:   "install",
	Short: "InstallRequirements the prerequisites of kubefire",
	RunE: func(cmd *cobra.Command, args []string) error {
		installers := []bootstrap.Bootstrapper{
			bootstrap.NewKubeadmBootstrapper(),
			bootstrap.NewSkubaBootstrapper(),
		}

		switch {
		case check:
			for _, i := range installers {
				if err := i.(bootstrap.BootstrapperInstaller).CheckRequirements(); err != nil {
					return err
				}
			}

		default:
			for _, i := range installers {
				if err := i.(bootstrap.BootstrapperInstaller).InstallRequirements(); err != nil {
					return err
				}
			}
		}

		return nil
	},
}

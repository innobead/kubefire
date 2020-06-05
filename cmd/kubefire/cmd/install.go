package cmd

import (
	"fmt"
	"github.com/innobead/kubefire/pkg/bootstrap"
	"github.com/spf13/cobra"
)

var check bool
var bootstrapper string

func init() {
	InstallCmd.Flags().BoolVar(&check, "check", false, "Check if the prerequisites are ready")
	InstallCmd.Flags().StringVar(&bootstrapper, "bootstrapper", string(bootstrap.KUBEADM), fmt.Sprintf("Bootstrapper type. ex: %s", bootstrap.BuiltinTypes))
}

var InstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install the prerequisites of kubefire",
	RunE: func(cmd *cobra.Command, args []string) error {
		installers := []bootstrap.Bootstrapper{bootstrap.NewKubeadmBootstrapper(), bootstrap.NewSkubaBootstrapper()}

		switch {
		case check:
			for _, i := range installers {
				if err := i.(bootstrap.BootstrapperInstaller).Check(); err != nil {
					return err
				}
			}
		default:
			for _, i := range installers {
				if err := i.(bootstrap.BootstrapperInstaller).Install(); err != nil {
					return err
				}
			}
		}

		return nil
	},
}

package kubeconfig

import (
	"github.com/innobead/kubefire/internal/config"
	"github.com/innobead/kubefire/internal/di"
	"github.com/innobead/kubefire/internal/validate"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"os"
)

var downloadCmd = &cobra.Command{
	Use:     "download [cluster-name]",
	Aliases: []string{"d"},
	Short:   "Downloads the kubeconfig of cluster",
	Args:    validate.OneArg("cluster name"),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return validate.CheckClusterExist(args[0])
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		cluster, err := di.ClusterManager().Get(name)
		if err != nil {
			return errors.WithMessagef(err, "failed to get cluster (%s) info", name)
		}

		config.Bootstrapper = cluster.Spec.Bootstrapper
		di.DelayInit(true)

		wd, _ := os.Getwd()
		if _, err := di.Bootstrapper().DownloadKubeConfig(cluster, wd); err != nil {
			return errors.WithMessagef(err, "failed to download kubeconfig of cluster (%s)", name)
		}

		return nil
	},
}

package cluster

import (
	"github.com/innobead/kubefire/internal/di"
	"github.com/innobead/kubefire/internal/validate"
	"github.com/innobead/kubefire/pkg/bootstrap"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"os"
)

var downloadCmd = &cobra.Command{
	Use:   "download [name]",
	Short: "Download the kubeconfig of cluster",
	Args:  validate.OneArg("name"),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return validate.CheckClusterExist(args[0])
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		cluster, err := di.ClusterManager().Get(name)
		if err != nil {
			return errors.WithMessagef(err, "failed to get cluster (%s) info", name)
		}

		wd, _ := os.Getwd()
		if _, err := bootstrap.New(cluster.Spec.Bootstrapper, di.NodeManager()).DownloadKubeConfig(cluster, wd); err != nil {
			return errors.WithMessagef(err, "failed to download kubeconfig of cluster (%s)", name)
		}

		return nil
	},
}

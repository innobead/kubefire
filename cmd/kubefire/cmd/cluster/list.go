package cluster

import (
	"github.com/innobead/kubefire/pkg/config"
	"github.com/innobead/kubefire/pkg/util"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List clusters",
	RunE: func(cmd *cobra.Command, args []string) error {
		clusters, err := util.ClusterManager().List()
		if err != nil {
			return errors.WithMessagef(err, "failed to list clusters info")
		}

		var configClusters []*config.Cluster
		for _, c := range clusters {
			configClusters = append(configClusters, &c.Spec)
		}

		if err := util.Output().Print(configClusters, []string{"Name", "Bootstrapper"}, ""); err != nil {
			return errors.WithMessagef(err, "failed to print output of clusters")
		}

		return nil
	},
}

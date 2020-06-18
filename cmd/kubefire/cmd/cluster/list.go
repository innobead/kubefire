package cluster

import (
	"github.com/innobead/kubefire/internal/di"
	"github.com/innobead/kubefire/pkg/config"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List clusters",
	RunE: func(cmd *cobra.Command, args []string) error {
		clusters, err := di.ClusterManager().List()
		if err != nil {
			return errors.WithMessagef(err, "failed to list clusters info")
		}

		var configClusters []*config.Cluster
		for _, c := range clusters {
			configClusters = append(configClusters, &c.Spec)
		}

		if err := di.Output().Print(configClusters, []string{"Name", "Bootstrapper"}, ""); err != nil {
			return errors.WithMessagef(err, "failed to print output of clusters")
		}

		return nil
	},
}

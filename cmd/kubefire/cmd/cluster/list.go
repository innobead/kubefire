package cluster

import (
	"github.com/innobead/kubefire/internal/util"
	"github.com/innobead/kubefire/pkg/config"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List clusters",
	RunE: func(cmd *cobra.Command, args []string) error {
		clusters, err := util.ClusterManager().List()
		if err != nil {
			return err
		}

		var configClusters []*config.Cluster
		for _, c := range clusters {
			configClusters = append(configClusters, &c.Spec)
		}

		if err := util.Output().Print(configClusters, []string{"name", "bootstrapper"}, ""); err != nil {
			return err
		}

		return nil
	},
}

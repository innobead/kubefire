package cluster

import (
	intcmd "github.com/innobead/kubefire/internal/cmd"
	"github.com/innobead/kubefire/internal/di"
	"github.com/innobead/kubefire/pkg/config"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List clusters",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		_, err := di.ClusterManager().List()
		if err != nil {
			return errors.WithMessagef(err, "failed to list clusters info")
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		clusters, _ := di.ClusterManager().List()

		var configClusters []*config.Cluster
		for _, c := range clusters {
			configClusters = append(configClusters, &c.Spec)
		}

		if err := di.Output().Print(configClusters, nil, ""); err != nil {
			return errors.WithMessagef(err, "failed to print output of clusters")
		}

		return nil
	},
}

func init() {
	intcmd.AddOutputFlag(listCmd)
}

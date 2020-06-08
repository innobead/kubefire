package cluster

import (
	"github.com/innobead/kubefire/internal/util"
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

		if err := util.Output().Print(clusters); err != nil {
			return err
		}

		return nil
	},
}

package cluster

import (
	"github.com/innobead/kubefire/internal/util"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get [name]",
	Short: "GetCluster cluster",
	Args:  util.Validate1thArg("name"),
	RunE: func(cmd *cobra.Command, args []string) error {
		cluster, err := util.ClusterManager().Get(args[0])
		if err != nil {
			return err
		}

		// print the cluster config
		if err := util.Output().Print(cluster, nil, ""); err != nil {
			return err
		}

		return nil
	},
}

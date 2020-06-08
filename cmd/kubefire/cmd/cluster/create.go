package cluster

import (
	"github.com/innobead/kubefire/internal/util"
	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Create an initialized cluster",
	Args:  util.Validate1thArg("name"),
	RunE: func(cmd *cobra.Command, args []string) error {
		cluster.Name = args[0]
		if err := util.ClusterManager().Create(cluster.Name); err != nil {
			return err
		}

		return nil
	},
}

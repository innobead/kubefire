package cluster

import (
	"github.com/innobead/kubefire/internal/util"
	"github.com/spf13/cobra"
)

var force bool

func init() {
	deleteCmd.Flags().BoolVar(&force, "force", false, "Force to delete cluster")
}

var deleteCmd = &cobra.Command{
	Use:   "delete [name]",
	Short: "DeleteCluster cluster",
	Args:  util.Validate1thArg("name"),
	RunE: func(cmd *cobra.Command, args []string) error {
		return util.ClusterManager().Delete(args[0], force)
	},
}

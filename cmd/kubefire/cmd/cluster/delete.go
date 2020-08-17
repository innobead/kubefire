package cluster

import (
	"github.com/innobead/kubefire/internal/di"
	"github.com/innobead/kubefire/internal/validate"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var forceDeleteCluster bool

var deleteCmd = &cobra.Command{
	Use:   "delete [name, ...]",
	Short: "Delete clusters",
	Args:  validate.MinimumArgs("name"),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return validate.ClusterExist(args[0])
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		for _, n := range args {
			if err := di.ClusterManager().Delete(n, forceDeleteCluster); err != nil {
				return errors.WithMessagef(err, "failed to delete cluster (%s)", n)
			}
		}

		return nil
	},
}

func init() {
	deleteCmd.Flags().BoolVar(&forceDeleteCluster, "force", false, "force to delete")
}

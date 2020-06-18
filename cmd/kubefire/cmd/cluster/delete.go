package cluster

import (
	"github.com/innobead/kubefire/internal/di"
	"github.com/innobead/kubefire/pkg/util"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var forceDelete bool

func init() {
	deleteCmd.Flags().BoolVar(&forceDelete, "force", false, "force to delete")
}

var deleteCmd = &cobra.Command{
	Use:   "delete [name]",
	Short: "Delete cluster",
	Args:  util.Validate1thArg("name"),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		if err := di.ClusterManager().Delete(name, forceDelete); err != nil {
			return errors.WithMessagef(err, "failed to delete cluster (%s)", name)
		}

		return nil
	},
}

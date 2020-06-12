package cluster

import (
	"github.com/innobead/kubefire/internal/util"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get [name]",
	Short: "Get cluster",
	Args:  util.Validate1thArg("name"),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		cluster, err := util.ClusterManager().Get(name)
		if err != nil {
			return errors.WithMessagef(err, "failed to get cluster (%s) info", name)
		}

		if err := util.Output().Print(cluster, nil, ""); err != nil {
			return errors.WithMessagef(err, "failed to print output of cluster (%s)", name)
		}

		return nil
	},
}

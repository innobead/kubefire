package cluster

import (
	"github.com/innobead/kubefire/internal/di"
	"github.com/innobead/kubefire/internal/validate"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop [name]",
	Short: "Stop cluster",
	Args:  validate.OneArg("name"),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return validate.CheckClusterExist(args[0])
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return stopCluster(args[0])
	},
}

func stopCluster(name string) error {
	if err := di.NodeManager().StopNodes(name); err != nil {
		return errors.WithMessagef(err, "failed to stop all nodes cluster (%s)", name)
	}

	return nil
}

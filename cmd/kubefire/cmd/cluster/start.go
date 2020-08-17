package cluster

import (
	"github.com/innobead/kubefire/internal/di"
	"github.com/innobead/kubefire/internal/validate"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start [name]",
	Short: "Start cluster",
	Args:  validate.OneArg("name"),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return validate.ClusterExist(args[0])
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		if err := di.NodeManager().StartNodes(name); err != nil {
			err := errors.WithMessagef(err, "failed to stop all nodes cluster (%s)", name)

			if !forceDeleteCluster {
				return err
			}

			logrus.WithError(err).WithField("node", name).Println()
		}

		cluster, err := di.ConfigManager().GetCluster(name)
		if err != nil {
			return err
		}

		if cluster.Deployed {
			return nil
		}

		if err := deployCluster(cluster.Name); err != nil {
			return err
		}

		return nil
	},
}

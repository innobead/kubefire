package node

import (
	"github.com/innobead/kubefire/internal/di"
	"github.com/innobead/kubefire/internal/validate"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop [name]",
	Short: "Stop node",
	Args:  validate.OneArg("node name"),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return validate.CheckNodeExist(args[0])
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return stopNode(args[0])
	},
}

func stopNode(name string) error {
	node, _ := di.NodeManager().GetNode(name)

	if !node.Status.Running {
		logrus.WithField("node", node.Name).Infoln("node is already stopped")
		return nil
	}

	if err := di.NodeManager().StopNode(name); err != nil {
		return errors.WithMessagef(err, "failed to stop node (%s)", name)
	}

	return nil
}

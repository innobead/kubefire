package node

import (
	"github.com/innobead/kubefire/internal/di"
	"github.com/innobead/kubefire/internal/validate"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get [name]",
	Short: "Get node",
	Args:  validate.OneArg("name"),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return validate.NodeExist(args[0])
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		node, err := di.NodeManager().GetNode(name)
		if err != nil {
			return errors.WithMessagef(err, "failed to get node (%s) info", name)
		}

		if err := di.Output().Print(node, nil, ""); err != nil {
			return errors.WithMessagef(err, "failed to print output of node (%s)", name)
		}

		return nil
	},
}

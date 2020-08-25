package node

import (
	"github.com/innobead/kubefire/internal/validate"
	"github.com/spf13/cobra"
)

var restartCmd = &cobra.Command{
	Use:   "restart [name]",
	Short: "Restart node",
	Args:  validate.OneArg("node name"),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return validate.CheckNodeExist(args[0])
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		if err := stopNode(name); err != nil {
			return err
		}

		if err := startNode(name); err != nil {
			return err
		}

		return nil
	},
}

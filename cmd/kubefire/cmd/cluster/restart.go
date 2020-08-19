package cluster

import (
	"github.com/innobead/kubefire/internal/validate"
	"github.com/spf13/cobra"
)

var restartCmd = &cobra.Command{
	Use:   "restart [name]",
	Short: "Restart cluster",
	Args:  validate.OneArg("name"),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return validate.ClusterExist(args[0])
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		if err := stopCluster(name); err != nil {
			return err
		}

		if _, err := startCluster(name); err != nil {
			return err
		}

		return nil
	},
}

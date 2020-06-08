package cluster

import (
	"github.com/spf13/cobra"
)

func init() {
	cmds := []*cobra.Command{
		initCmd,
		createCmd,
		deleteCmd,
		getCmd,
		listCmd,
	}

	for _, c := range cmds {
		Cmd.AddCommand(c)
	}
}

var Cmd = &cobra.Command{
	Use:   "cluster",
	Short: "Manage cluster",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

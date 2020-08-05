package node

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "node",
	Short: "Manage node",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func init() {
	cmds := []*cobra.Command{
		sshCmd,
	}

	for _, c := range cmds {
		Cmd.AddCommand(c)
	}
}

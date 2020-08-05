package cluster

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "cluster",
	Short: "Manage cluster",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func init() {
	cmds := []*cobra.Command{
		createCmd,
		deleteCmd,
		getCmd,
		listCmd,
		downloadCmd,
		envCmd,
	}

	for _, c := range cmds {
		Cmd.AddCommand(c)
	}
}

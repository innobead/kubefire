package node

import (
	"github.com/innobead/kubefire/internal/validate"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:     "node",
	Aliases: []string{"n"},
	Short:   "Manage node",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return validate.CheckPrerequisites()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func init() {
	cmds := []*cobra.Command{
		sshCmd,
		getCmd,
		startCmd,
		stopCmd,
		restartCmd,
	}

	for _, c := range cmds {
		Cmd.AddCommand(c)
	}
}

package cluster

import (
	"github.com/innobead/kubefire/internal/validate"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:     "cluster",
	Aliases: []string{"c"},
	Short:   "Manage cluster",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return validate.CheckPrerequisites()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func init() {
	cmds := []*cobra.Command{
		createCmd,
		startCmd,
		stopCmd,
		restartCmd,
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

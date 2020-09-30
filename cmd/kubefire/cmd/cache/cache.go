package cache

import (
	"github.com/innobead/kubefire/internal/di"
	"github.com/innobead/kubefire/internal/validate"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:     "cache",
	Aliases: []string{"c"},
	Short:   "Manage caches",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		di.DelayInit(false)
		return validate.CheckPrerequisites()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func init() {
	cmds := []*cobra.Command{
		showCmd,
		deleteCmd,
	}

	for _, c := range cmds {
		Cmd.AddCommand(c)
	}
}

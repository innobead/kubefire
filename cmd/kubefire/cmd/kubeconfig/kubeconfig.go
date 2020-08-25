package kubeconfig

import (
	"github.com/innobead/kubefire/internal/di"
	"github.com/innobead/kubefire/internal/validate"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:     "kubeconfig",
	Aliases: []string{"k"},
	Short:   "Manage kubeconfig of clusters",
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
		getCmd,
		downloadCmd,
	}

	for _, c := range cmds {
		Cmd.AddCommand(c)
	}
}

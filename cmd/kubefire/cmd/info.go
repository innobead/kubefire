package cmd

import (
	intcmd "github.com/innobead/kubefire/internal/cmd"
	"github.com/innobead/kubefire/internal/di"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var InfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show runtime info",
	RunE: func(cmd *cobra.Command, args []string) error {
		versionsInfo := intcmd.CurrentPrerequisitesInfos()
		if err := di.Output().Print(versionsInfo, nil, ""); err != nil {
			return errors.WithMessage(err, "failed to print output of runtime info")
		}

		return nil
	},
}

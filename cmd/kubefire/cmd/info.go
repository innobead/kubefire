package cmd

import (
	"github.com/innobead/kubefire/internal/di"
	"github.com/innobead/kubefire/internal/info"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var InfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show runtime info",
	RunE: func(cmd *cobra.Command, args []string) error {
		versionInfo := info.CurrentRuntimeVersionInfo()
		if err := di.Output().Print(versionInfo, nil, ""); err != nil {
			return errors.WithMessage(err, "failed to print output of runtime info")
		}

		return nil
	},
}

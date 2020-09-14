package cmd

import (
	"github.com/innobead/kubefire/internal/config"
	"github.com/innobead/kubefire/pkg/output"
	"github.com/innobead/kubefire/pkg/util"
	"github.com/spf13/cobra"
)

func AddOutputFlag(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&config.Output, "output", "o", string(output.DEFAULT), util.FlagsValuesUsage("output format", output.BuiltinTypes))
}

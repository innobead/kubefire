package cluster

import "github.com/spf13/cobra"

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate cluster default config",
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

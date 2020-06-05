package cluster

import "github.com/spf13/cobra"

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List clusters",
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}
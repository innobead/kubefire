package node

import (
	"github.com/innobead/kubefire/internal/di"
	"github.com/innobead/kubefire/internal/validate"
	"github.com/spf13/cobra"
)

var sshCmd = &cobra.Command{
	Use:   "ssh [name]",
	Short: "SSH into node",
	Args:  validate.OneArg("node name"),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return validate.CheckNodeExist(args[0])
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return di.ClusterManager().GetNodeManager().LoginBySSH(
			args[0],
			di.ClusterManager().GetConfigManager(),
		)
	},
}

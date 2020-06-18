package node

import (
	"github.com/innobead/kubefire/internal/di"
	"github.com/innobead/kubefire/pkg/util"
	"github.com/spf13/cobra"
)

var sshCmd = &cobra.Command{
	Use:   "ssh [name]",
	Short: "SSH into node",
	Args:  util.Validate1thArg("name"),
	RunE: func(cmd *cobra.Command, args []string) error {
		return di.ClusterManager().GetNodeManager().LoginBySSH(
			args[0],
			di.ClusterManager().GetConfigManager(),
		)
	},
}

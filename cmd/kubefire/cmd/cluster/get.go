package cluster

import (
	"fmt"
	"github.com/innobead/kubefire/internal/util"
	"github.com/innobead/kubefire/pkg/config"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get [name]",
	Short: "Get cluster",
	Args:  util.Validate1thArg("name"),
	RunE: func(cmd *cobra.Command, args []string) error {
		cluster, err := util.ClusterManager().Get(args[0])
		if err != nil {
			return err
		}

		if err := util.Output().Print(cluster.Config); err != nil {
			return err
		}
		println("")

		data := map[string]*config.Node{
			"Admin Node":  &cluster.Config.Admin,
			"Master Node": &cluster.Config.Master,
			"Worker Node": &cluster.Config.Worker,
		}

		for k, v := range data {
			fmt.Printf("# %s\n", k)

			if err := util.Output().Print(v); err != nil {
				return err
			}
		}

		return nil
	},
}

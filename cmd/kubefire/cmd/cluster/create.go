package cluster

import (
	"github.com/innobead/kubefire/internal/util"
	config2 "github.com/innobead/kubefire/pkg/config"
	"github.com/spf13/cobra"
)

var cluster = config2.Cluster{
	Master: config2.Node{},
	Worker: config2.Node{},
}

func init() {
	flags := createCmd.Flags()

	flags.StringVar((*string)(&cluster.Bootstrapper), "bootstrapper", "", "")
	flags.StringVar(&cluster.Pubkey, "pubkey", "", "")
	flags.StringVar(&cluster.Image, "image", "innobead/kubefire-sle15sp1", "")
	flags.StringVar(&cluster.KernelImage, "kernel-image", "innobead/ignite-kernel:4.19.125-amd64", "")
	flags.StringVar(&cluster.KernelArgs, "kernel-args", "console=ttyS0 reboot=k panic=1 pci=off ip=dhcp security=apparmor apparmor=1", "")

	flags.IntVar(&cluster.Master.Count, "master-count", 1, "")
	flags.IntVar(&cluster.Master.Cpus, "master-cpu", 2, "")
	flags.StringVar(&cluster.Master.Memory, "master-memory", "2GB", "")
	flags.StringVar(&cluster.Master.Size, "master-size", "10GB", "")

	flags.IntVar(&cluster.Worker.Count, "worker-count", 1, "")
	flags.IntVar(&cluster.Worker.Cpus, "worker-cpu", 2, "")
	flags.StringVar(&cluster.Worker.Memory, "worker-memory", "2GB", "")
	flags.StringVar(&cluster.Worker.Size, "worker-size", "10GB", "")
}

var createCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Create cluster",
	Args:  util.Validate1thArg("name"),
	RunE: func(cmd *cobra.Command, args []string) error {
		//TODO
		// - create VMs
		// - deploy the cluster via the bootstrapper

		return nil
	},
}

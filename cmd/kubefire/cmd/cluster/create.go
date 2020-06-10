package cluster

import (
	"fmt"
	"github.com/innobead/kubefire/internal/util"
	"github.com/innobead/kubefire/pkg/bootstrap"
	pkgconfig "github.com/innobead/kubefire/pkg/config"
	"github.com/spf13/cobra"
)

var cluster = pkgconfig.Cluster{
	Master: pkgconfig.Node{},
	Worker: pkgconfig.Node{},
}

func init() {
	flags := createCmd.Flags()

	flags.StringVar(&cluster.Bootstrapper, "bootstrapper", bootstrap.KUBEADM, fmt.Sprintf("bootstrapper type. ex: %v", bootstrap.BuiltinTypes))
	flags.StringVar(&cluster.Pubkey, "pubkey", "", "")
	flags.StringVar(&cluster.Image, "image", "innobead/kubefire:sle15sp1-latest", "")
	flags.StringVar(&cluster.KernelImage, "kernel-image", "innobead/ignite-kernel:4.19.125-amd64", "")
	flags.StringVar(&cluster.KernelArgs, "kernel-args", "console=ttyS0 reboot=k panic=1 pci=off ip=dhcp security=apparmor apparmor=1", "")

	flags.IntVar(&cluster.Admin.Count, "admin-count", 0, "")
	flags.IntVar(&cluster.Admin.Cpus, "admin-cpu", 1, "")
	flags.StringVar(&cluster.Admin.Memory, "admin-memory", "512MB", "")
	flags.StringVar(&cluster.Admin.DiskSize, "admin-size", "2GB", "")

	flags.IntVar(&cluster.Master.Count, "master-count", 1, "")
	flags.IntVar(&cluster.Master.Cpus, "master-cpu", 2, "")
	flags.StringVar(&cluster.Master.Memory, "master-memory", "2GB", "")
	flags.StringVar(&cluster.Master.DiskSize, "master-size", "10GB", "")

	flags.IntVar(&cluster.Worker.Count, "worker-count", 1, "")
	flags.IntVar(&cluster.Worker.Cpus, "worker-cpu", 2, "")
	flags.StringVar(&cluster.Worker.Memory, "worker-memory", "2GB", "")
	flags.StringVar(&cluster.Worker.DiskSize, "worker-size", "10GB", "")
}

var createCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Create an initialized cluster",
	Args:  util.Validate1thArg("name"),
	RunE: func(cmd *cobra.Command, args []string) error {
		cluster.Name = args[0]

		if err := util.ClusterManager().Init(&cluster); err != nil {
			return err
		}

		if err := util.ClusterManager().Create(cluster.Name); err != nil {
			return err
		}

		return nil
	},
}

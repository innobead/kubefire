package cluster

import (
	"github.com/innobead/kubefire/internal/config"
	"github.com/innobead/kubefire/internal/di"
	"github.com/innobead/kubefire/pkg/bootstrap"
	pkgconfig "github.com/innobead/kubefire/pkg/config"
	"github.com/innobead/kubefire/pkg/util"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var cluster = pkgconfig.Cluster{
	Master: pkgconfig.Node{},
	Worker: pkgconfig.Node{},
}

var forceCreate bool

func init() {
	flags := createCmd.Flags()

	flags.StringVar(&cluster.Bootstrapper, "bootstrapper", bootstrap.KUBEADM, util.FlagsValuesUsage("bootstrapper type", bootstrap.BuiltinTypes))
	flags.StringVar(&cluster.Pubkey, "pubkey", "", "Public key")
	flags.StringVar(&cluster.Image, "image", "innobead/kubefire:sle15sp1-latest", "rootfs container image")
	flags.StringVar(&cluster.KernelImage, "kernel-image", "innobead/ignite-kernel:4.19.125-amd64", "kernel container image")
	flags.StringVar(&cluster.KernelArgs, "kernel-args", "console=ttyS0 reboot=k panic=1 pci=off ip=dhcp security=apparmor apparmor=1", "kernel arguments")
	flags.StringVar(&cluster.ExtraOptions, "extra-opts", "", "extra options (ex: key=value,...) for bootstrapper")

	flags.IntVar(&cluster.Admin.Count, "admin-count", 0, "count of admin node")
	flags.IntVar(&cluster.Admin.Cpus, "admin-cpu", 1, "CPUs of admin node")
	flags.StringVar(&cluster.Admin.Memory, "admin-memory", "512MB", "memory of admin node")
	flags.StringVar(&cluster.Admin.DiskSize, "admin-size", "2GB", "disk size of admin node")

	flags.IntVar(&cluster.Master.Count, "master-count", 1, "count of master node")
	flags.IntVar(&cluster.Master.Cpus, "master-cpu", 2, "CPUs of master node")
	flags.StringVar(&cluster.Master.Memory, "master-memory", "2GB", "memory of master node")
	flags.StringVar(&cluster.Master.DiskSize, "master-size", "10GB", "disk size of master node")

	flags.IntVar(&cluster.Worker.Count, "worker-count", 0, "count of worker node")
	flags.IntVar(&cluster.Worker.Cpus, "worker-cpu", 2, "CPUs of worker node")
	flags.StringVar(&cluster.Worker.Memory, "worker-memory", "2GB", "memory of worker node")
	flags.StringVar(&cluster.Worker.DiskSize, "worker-size", "10GB", "disk size of worker node")

	flags.BoolVar(&forceCreate, "force", false, "force to recreate")

	config.Bootstrap = cluster.Bootstrapper
}

var createCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Create cluster",
	Args:  util.Validate1thArg("name"),
	RunE: func(cmd *cobra.Command, args []string) error {
		cluster.Name = args[0]

		if forceCreate {
			_ = di.ClusterManager().Delete(cluster.Name, true)
		}

		if err := di.ClusterManager().Init(&cluster); err != nil {
			return errors.WithMessagef(err, "failed to init cluster (%s)", cluster.Name)
		}

		if err := di.ClusterManager().Create(cluster.Name); err != nil {
			return errors.WithMessagef(err, "failed to create cluster (%s)", cluster.Name)
		}

		c, err := di.ClusterManager().Get(cluster.Name)
		if err != nil {
			return errors.WithMessagef(err, "failed to get cluster (%s) before bootstrapping", cluster.Name)
		}

		if err := di.Bootstrapper().Deploy(c); err != nil {
			return errors.WithMessagef(err, "failed to deploy cluster (%s)", c.Spec.Bootstrapper)
		}

		return nil
	},
}

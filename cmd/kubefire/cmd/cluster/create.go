package cluster

import (
	"github.com/innobead/kubefire/internal/config"
	"github.com/innobead/kubefire/internal/di"
	"github.com/innobead/kubefire/internal/validate"
	"github.com/innobead/kubefire/pkg/bootstrap"
	pkgconfig "github.com/innobead/kubefire/pkg/config"
	"github.com/innobead/kubefire/pkg/util"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	cluster = pkgconfig.NewCluster()
	started bool
)

var createCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Create cluster",
	Args:  validate.OneArg("name"),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if !bootstrap.IsValid(cluster.Bootstrapper) {
			return errors.Errorf("%s unsupported bootstrapper", cluster.Bootstrapper)
		}
		config.Bootstrapper = cluster.Bootstrapper

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		cluster.Name = args[0]

		if forceDeleteCluster {
			_ = di.ClusterManager().Delete(cluster.Name, true)
		}

		if err := di.ClusterManager().Init(cluster); err != nil {
			return errors.WithMessagef(err, "failed to init cluster (%s)", cluster.Name)
		}

		if err := di.ClusterManager().Create(cluster.Name, started); err != nil {
			return errors.WithMessagef(err, "failed to create cluster (%s)", cluster.Name)
		}

		if started {
			if err := deployCluster(cluster.Name); err != nil {
				return err
			}
		}

		return nil
	},
}

func init() {
	flags := createCmd.Flags()

	flags.StringVar(&cluster.Bootstrapper, "bootstrapper", bootstrap.KUBEADM, util.FlagsValuesUsage("bootstrapper type", bootstrap.BuiltinTypes))
	flags.StringVar(&cluster.Pubkey, "pubkey", "", "Public key")
	flags.StringVar(&cluster.Image, "image", "innobead/kubefire-opensuse-leap:15.2", "rootfs container image")
	flags.StringVar(&cluster.KernelImage, "kernel-image", "innobead/kubefire-kernel-4.19.125-amd64:latest", "kernel container image")
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

	flags.BoolVar(&forceDeleteCluster, "force", false, "force to recreate if the cluster exists")
	flags.BoolVar(&started, "start", true, "start nodes")
}

func deployCluster(name string) error {
	cluster, err := di.ClusterManager().Get(name)
	if err != nil {
		return errors.WithMessagef(err, "failed to get cluster (%s) before bootstrapping", cluster.Name)
	}

	err = di.Bootstrapper().Deploy(
		cluster,
		func() error {
			return di.Bootstrapper().Prepare(forceDeleteCluster)
		},
	)
	if err != nil {
		return errors.WithMessagef(err, "failed to deploy cluster (%s)", cluster.Name)
	}

	cluster.Spec.Deployed = true
	if err := di.ConfigManager().SaveCluster(&cluster.Spec); err != nil {
		return errors.WithMessagef(err, "failed to mark the cluster (%s) as deployed", cluster.Name)
	}

	if _, err := di.Bootstrapper().DownloadKubeConfig(cluster, ""); err != nil {
		return errors.WithMessagef(err, "failed to download the kubeconfig of cluster (%s)", cluster.Name)
	}

	return nil
}

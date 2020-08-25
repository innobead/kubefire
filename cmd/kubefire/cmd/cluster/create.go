package cluster

import (
	"github.com/innobead/kubefire/internal/config"
	"github.com/innobead/kubefire/internal/di"
	"github.com/innobead/kubefire/internal/validate"
	"github.com/innobead/kubefire/pkg/bootstrap"
	pkgconfig "github.com/innobead/kubefire/pkg/config"
	"github.com/innobead/kubefire/pkg/constants"
	"github.com/innobead/kubefire/pkg/util"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"regexp"
	"strings"
)

var (
	cluster = pkgconfig.NewCluster()
	started bool
	cached  bool
)

var createCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Create cluster",
	Args:  validate.OneArg("cluster name"),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if err := validate.CheckBootstrapperType(cluster.Bootstrapper); err != nil {
			return err
		}
		config.Bootstrapper = cluster.Bootstrapper
		di.DelayInit(true)

		if err := validate.CheckClusterVersion(cluster.Version); err != nil {
			return err
		}

		if !cached {
			_ = di.ConfigManager().DeleteBootstrapperVersions(pkgconfig.NewBootstrapperVersion(cluster.Bootstrapper, ""))
		}

		if _, _, err := bootstrap.GenerateSaveBootstrapperVersions(config.Bootstrapper, di.ConfigManager()); err != nil {
			return err
		}

		version, err := correctClusterVersion(cluster.Version)
		if err != nil {
			return err
		}
		cluster.Version = version

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

	flags.StringVar(&cluster.Bootstrapper, "bootstrapper", constants.KUBEADM, util.FlagsValuesUsage("Bootstrapper type", bootstrap.BuiltinTypes))
	flags.StringVar(&cluster.Pubkey, "pubkey", "", "Public key")
	flags.StringVar(&cluster.Version, "version", "", "Version of Kubernetes supported by bootstrapper (ex: v1.18, v1.18.8, empty)")
	flags.StringVar(&cluster.Image, "image", "innobead/kubefire-opensuse-leap:15.2", "Rootfs container image")
	flags.StringVar(&cluster.KernelImage, "kernel-image", "innobead/kubefire-kernel-4.19.125-amd64:latest", "Kernel container image")
	flags.StringVar(&cluster.KernelArgs, "kernel-args", "console=ttyS0 reboot=k panic=1 pci=off ip=dhcp security=apparmor apparmor=1", "Kernel arguments")
	flags.StringVar(&cluster.ExtraOptions, "extra-opts", "", "Extra options (ex: key=value,...) for bootstrapper")

	flags.IntVar(&cluster.Admin.Count, "admin-count", 0, "Count of admin node")
	flags.IntVar(&cluster.Admin.Cpus, "admin-cpu", 1, "CPUs of admin node")
	flags.StringVar(&cluster.Admin.Memory, "admin-memory", "512MB", "Memory of admin node")
	flags.StringVar(&cluster.Admin.DiskSize, "admin-size", "2GB", "Disk size of admin node")

	flags.IntVar(&cluster.Master.Count, "master-count", 1, "Count of master node")
	flags.IntVar(&cluster.Master.Cpus, "master-cpu", 2, "CPUs of master node")
	flags.StringVar(&cluster.Master.Memory, "master-memory", "2GB", "Memory of master node")
	flags.StringVar(&cluster.Master.DiskSize, "master-size", "10GB", "Disk size of master node")

	flags.IntVar(&cluster.Worker.Count, "worker-count", 0, "Count of worker node")
	flags.IntVar(&cluster.Worker.Cpus, "worker-cpu", 2, "CPUs of worker node")
	flags.StringVar(&cluster.Worker.Memory, "worker-memory", "2GB", "Memory of worker node")
	flags.StringVar(&cluster.Worker.DiskSize, "worker-size", "10GB", "Disk size of worker node")

	flags.BoolVar(&forceDeleteCluster, "force", false, "Force to recreate if the cluster exists")
	flags.BoolVar(&cached, "cache", true, "Use caches")
	flags.BoolVar(&started, "start", true, "Start nodes")
}

func deployCluster(name string) error {
	cluster, err := di.ClusterManager().Get(name)
	if err != nil {
		return errors.WithMessagef(err, "failed to get cluster (%s) before bootstrapping", cluster.Name)
	}

	err = di.Bootstrapper().Deploy(
		cluster,
		func() error {
			return di.Bootstrapper().Prepare(cluster, forceDeleteCluster)
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

func correctClusterVersion(version string) (string, error) {
	latestVersion, err := di.VersionFinder().GetLatestVersion()
	if err != nil {
		return "", err
	}

	if version == "" {
		return latestVersion.String(), nil
	}

	bootstrapperVersion := pkgconfig.NewBootstrapperVersion(di.Bootstrapper().Type(), latestVersion.String())
	versions, err := di.ConfigManager().GetBootstrapperVersions(bootstrapperVersion)
	if err != nil {
		return "", err
	}

	patternCheckMajorMinorVersion := regexp.MustCompile(`^v\d+\.\d+$`)
	patternCheckMajorMinorPatchVersion := regexp.MustCompile(`^v\d+\.\d+\.\d+$`)

	for _, v := range versions {
		if version == v.Version() {
			return v.Version(), nil
		}

		if strings.HasPrefix(v.Version(), version+".") {
			return v.Version(), nil
		}

		switch {
		case version == v.Version():
			return v.Version(), nil

		case patternCheckMajorMinorVersion.MatchString(version) && strings.HasPrefix(v.Version(), version+"."):
			return v.Version(), nil
		}
	}

	if patternCheckMajorMinorPatchVersion.MatchString(version) && di.VersionFinder().HasPatchVersion(version) {
		return version, nil
	}

	return "", errors.New("version not found")
}

package bootstrap

import (
	"fmt"
	interr "github.com/innobead/kubefire/internal/error"
	"github.com/innobead/kubefire/pkg/bootstrap/versionfinder"
	pkgconfig "github.com/innobead/kubefire/pkg/config"
	"github.com/innobead/kubefire/pkg/constants"
	"github.com/innobead/kubefire/pkg/data"
	"github.com/innobead/kubefire/pkg/node"
	utilssh "github.com/innobead/kubefire/pkg/util/ssh"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"os"
	"path"
	"strings"
)

var BuiltinTypes = []string{
	constants.KUBEADM,
	constants.SKUBA,
	constants.K3S,
}

type Bootstrapper interface {
	Deploy(cluster *data.Cluster, before func() error) error
	DownloadKubeConfig(cluster *data.Cluster, destDir string) (string, error)
	Prepare(cluster *data.Cluster, force bool) error
	Type() string
}

func New(bootstrapper string) Bootstrapper {
	switch bootstrapper {
	case constants.SKUBA:
		return NewSkubaBootstrapper()

	case constants.KUBEADM, "":
		return NewKubeadmBootstrapper()

	case constants.K3S:
		return NewK3sBootstrapper()

	default:
		panic("no supported bootstrapper")
	}
}

func IsValid(bootstrapper string) bool {
	switch bootstrapper {
	case constants.KUBEADM, constants.SKUBA, constants.K3S:
		return true
	default:
		return false
	}
}

func GenerateSaveBootstrapperVersions(bootstrapperType string, configManager pkgconfig.Manager) (bootstrapperLatestVersion pkgconfig.BootstrapperVersioner, bootstrapperVersions []pkgconfig.BootstrapperVersioner, err error) {
	versionFinder := versionfinder.New(bootstrapperType)

	latestVersion, err := versionFinder.GetLatestVersion()
	if err != nil {
		return
	}

	bootstrapperVersion := pkgconfig.NewBootstrapperVersion(bootstrapperType, latestVersion.String())
	if _, err = os.Stat(bootstrapperVersion.LocalVersionFile()); !os.IsNotExist(err) {
		bootstrapperVersions, err = configManager.GetBootstrapperVersions(bootstrapperVersion)
		if err != nil {
			return
		}

		return
	}

	var versions []*data.Version
	versions, err = versionFinder.GetVersionsAfterVersion(*latestVersion)
	if err != nil {
		err = errors.WithMessagef(err, "failed to get the supported versions after/include %s from bootstrapper %s", latestVersion.String(), bootstrapperType)
		return
	}

	switch versionFinder := versionFinder.(type) {
	case *versionfinder.KubeadmVersionFinder:

		var critoolVersions []*data.Version
		critoolVersions, err = versionFinder.GetCritoolVersionsAfterVersion(*latestVersion)
		if err != nil {
			err = errors.WithMessagef(err, "failed to get the CriTools versions after/include %s from bootstrapper %s", latestVersion.String(), bootstrapperType)
			return
		}

		var kubeReleaseToolLatestVersion *data.Version
		kubeReleaseToolLatestVersion, err = versionFinder.GetKubeReleaseToolLatestVersion(*latestVersion)
		if err != nil {
			err = errors.WithMessagef(err, "failed to get the kubernetes release tool version from bootstrapper %s", bootstrapperType)
			return
		}

		for i, v := range versions {
			bv := pkgconfig.NewKubeadmBootstrapperVersion(
				v.String(),
				critoolVersions[i].String(),
				kubeReleaseToolLatestVersion.String(),
			)
			bootstrapperVersions = append(bootstrapperVersions, bv)

			if bv.Version() == latestVersion.String() {
				bootstrapperLatestVersion = bv
			}
		}

	case *versionfinder.K3sVersionFinder:

		for _, v := range versions {
			bv := pkgconfig.NewK3sBootstrapperVersion(v.String())
			bootstrapperVersions = append(bootstrapperVersions, bv)

			if bv.Version() == latestVersion.String() {
				bootstrapperLatestVersion = bv
			}
		}

	case *versionfinder.SkubaVersionFinder:

		for _, v := range versions {
			bv := pkgconfig.NewSkubaBootstrapperVersion(v.String())
			bootstrapperVersions = append(bootstrapperVersions, bv)

			if bv.Version() == latestVersion.String() {
				bootstrapperLatestVersion = bv
			}
		}
	}

	if err = configManager.SaveBootstrapperVersions(bootstrapperLatestVersion, bootstrapperVersions); err != nil {
		return
	}

	return
}

func downloadKubeConfig(nodeManager node.Manager, cluster *data.Cluster, remoteKubeConfigPath string, destDir string) (string, error) {
	logrus.Infof("downloading the kubeconfig of cluster (%s)", cluster.Name)

	firstMaster, err := nodeManager.GetNode(node.Name(cluster.Name, node.Master, 1))
	if err != nil {
		return "", err
	}

	sshClient, err := utilssh.NewClient(
		firstMaster.Name,
		cluster.Spec.Prikey,
		"root",
		firstMaster.Status.IPAddresses,
		nil,
	)
	if err != nil {
		return "", err
	}
	defer sshClient.Close()

	destPath := cluster.Spec.LocalKubeConfig()

	if destDir != "" {
		destPath = path.Join(destDir, "admin.conf")
	}

	logrus.Infof("saved the kubeconfig of cluster (%s) to %s", cluster.Name, destPath)

	if remoteKubeConfigPath == "" {
		remoteKubeConfigPath = "/etc/kubernetes/admin.conf"
	}

	if err := sshClient.Download(remoteKubeConfigPath, destPath); err != nil {
		return "", err
	}

	return destPath, nil
}

func getSupportedBootstrapperVersion(versionFinder versionfinder.Finder, configManager pkgconfig.Manager, bootstrapper Bootstrapper, version string) (pkgconfig.BootstrapperVersioner, error) {
	latestVersion, err := versionFinder.GetLatestVersion()
	if err != nil {
		return nil, err
	}

	bootstrapperVersion := pkgconfig.NewBootstrapperVersion(bootstrapper.Type(), latestVersion.String())
	versions, err := configManager.GetBootstrapperVersions(bootstrapperVersion)
	if err != nil {
		return nil, err
	}

	for _, v := range versions {
		if v.Version() == version {
			return v, nil
		}

		if strings.HasPrefix(version, data.ParseVersion(v.Version()).MajorMinorString()) {
			switch v := v.(type) {
			case *pkgconfig.KubeadmBootstrapperVersion:
				v.BootstrapperVersion = version
			}

			return v, nil
		}
	}

	return nil, errors.WithMessagef(
		interr.NotFoundError,
		fmt.Sprintf("bootstrapper=%s, version=%s", bootstrapper.Type(), version),
	)
}

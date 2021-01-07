package config

import (
	"fmt"
	"github.com/innobead/kubefire/internal/config"
	"github.com/innobead/kubefire/pkg/constants"
	"path"
	"strings"
)

type BootstrapperVersioner interface {
	Type() string
	LocalVersionFile() string
	Version() string
	Display() string
}

type BaseBootstrapperVersion struct {
	BootstrapperVersion string `json:"version"`
	BootstrapperType    string `json:"type"`
}

type KubeadmBootstrapperVersion struct {
	BaseBootstrapperVersion

	CrictlVersion      string `json:"crictl_version"`
	KubeReleaseVersion string `json:"kube_release_version"`
}

type K3sBootstrapperVersion struct {
	BaseBootstrapperVersion
}

type RKEBootstrapperVersion struct {
	BaseBootstrapperVersion
	KubernetesVersions []string `json:"kubernetes_versions"`
}

type RKE2BootstrapperVersion struct {
	BaseBootstrapperVersion
}

type K0sBootstrapperVersion struct {
	BaseBootstrapperVersion
}

var _ BootstrapperVersioner = (*KubeadmBootstrapperVersion)(nil)
var _ BootstrapperVersioner = (*K3sBootstrapperVersion)(nil)
var _ BootstrapperVersioner = (*RKEBootstrapperVersion)(nil)
var _ BootstrapperVersioner = (*RKE2BootstrapperVersion)(nil)
var _ BootstrapperVersioner = (*K0sBootstrapperVersion)(nil)

func NewBootstrapperVersion(bootstrapperType string, version string) BootstrapperVersioner {
	bootstrapperVersion := BaseBootstrapperVersion{
		BootstrapperVersion: version,
		BootstrapperType:    bootstrapperType,
	}

	switch bootstrapperType {
	case constants.KUBEADM:
		return &KubeadmBootstrapperVersion{
			BaseBootstrapperVersion: bootstrapperVersion,
			CrictlVersion:           "",
			KubeReleaseVersion:      "",
		}
	case constants.K3S:
		return &K3sBootstrapperVersion{BaseBootstrapperVersion: bootstrapperVersion}
	case constants.RKE:
		return &RKEBootstrapperVersion{BaseBootstrapperVersion: bootstrapperVersion}
	case constants.RKE2:
		return &RKE2BootstrapperVersion{BaseBootstrapperVersion: bootstrapperVersion}
	case constants.K0s:
		return &K0sBootstrapperVersion{BaseBootstrapperVersion: bootstrapperVersion}
	}

	return nil
}

func NewKubeadmBootstrapperVersion(bootstrapperVersion string, crictlVersion string, kubeReleaseVersion string) *KubeadmBootstrapperVersion {
	return &KubeadmBootstrapperVersion{
		BaseBootstrapperVersion: BaseBootstrapperVersion{
			BootstrapperVersion: bootstrapperVersion,
			BootstrapperType:    constants.KUBEADM,
		},
		CrictlVersion:      crictlVersion,
		KubeReleaseVersion: kubeReleaseVersion,
	}
}

func NewK3sBootstrapperVersion(bootstrapperVersion string) *K3sBootstrapperVersion {
	return &K3sBootstrapperVersion{
		BaseBootstrapperVersion: BaseBootstrapperVersion{
			BootstrapperVersion: bootstrapperVersion,
			BootstrapperType:    constants.K3S,
		},
	}
}

func NewRKEBootstrapperVersion(bootstrapperVersion string, kubernetesVersions []string) *RKEBootstrapperVersion {
	return &RKEBootstrapperVersion{
		BaseBootstrapperVersion: BaseBootstrapperVersion{
			BootstrapperVersion: bootstrapperVersion,
			BootstrapperType:    constants.RKE,
		},
		KubernetesVersions: kubernetesVersions,
	}
}

func NewRKE2BootstrapperVersion(bootstrapperVersion string) *RKE2BootstrapperVersion {
	return &RKE2BootstrapperVersion{
		BaseBootstrapperVersion: BaseBootstrapperVersion{
			BootstrapperVersion: bootstrapperVersion,
			BootstrapperType:    constants.RKE2,
		},
	}
}

func NewK0sBootstrapperVersion(bootstrapperVersion string) *K0sBootstrapperVersion {
	return &K0sBootstrapperVersion{
		BaseBootstrapperVersion: BaseBootstrapperVersion{
			BootstrapperVersion: bootstrapperVersion,
			BootstrapperType:    constants.K0s,
		},
	}
}

func (b *BaseBootstrapperVersion) Type() string {
	return b.BootstrapperType
}

func (b *BaseBootstrapperVersion) LocalVersionFile() string {
	return path.Join(
		BootstrapperRootDir,
		config.TagVersion,
		b.BootstrapperType,
		b.BootstrapperVersion+".yaml",
	)
}

func (b *BaseBootstrapperVersion) Version() string {
	return b.BootstrapperVersion
}

func (b *BaseBootstrapperVersion) Display() string {
	return b.Version()
}

func (s *RKEBootstrapperVersion) Display() string {
	return fmt.Sprintf(
		"%s (%s)",
		s.Version(),
		strings.Join(s.KubernetesVersions, ", "),
	)
}

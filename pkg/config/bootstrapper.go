package config

import (
	"github.com/innobead/kubefire/internal/config"
	"github.com/innobead/kubefire/pkg/constants"
	"path"
)

type BootstrapperVersioner interface {
	Type() string
	LocalVersionFile() string
	Version() string
}

type KubeadmBootstrapperVersion struct {
	BootstrapperVersion string `json:"version"`
	BootstrapperType    string `json:"type"`
	CrictlVersion       string `json:"crictl_version"`
	KubeReleaseVersion  string `json:"kube_release_version"`
}

func NewKubeadmBootstrapperVersion(bootstrapperVersion string, crictlVersion string, kubeReleaseVersion string) *KubeadmBootstrapperVersion {
	return &KubeadmBootstrapperVersion{BootstrapperVersion: bootstrapperVersion, BootstrapperType: constants.KUBEADM, CrictlVersion: crictlVersion, KubeReleaseVersion: kubeReleaseVersion}
}

type K3sBootstrapperVersion struct {
	BootstrapperVersion string `json:"version"`
	BootstrapperType    string `json:"type"`
}

func NewK3sBootstrapperVersion(bootstrapperVersion string) *K3sBootstrapperVersion {
	return &K3sBootstrapperVersion{BootstrapperVersion: bootstrapperVersion, BootstrapperType: constants.K3S}
}

type SkubaBootstrapperVersion struct {
	BootstrapperVersion string `json:"version"`
	BootstrapperType    string `json:"type"`
}

func NewSkubaBootstrapperVersion(bootstrapperVersion string) *SkubaBootstrapperVersion {
	return &SkubaBootstrapperVersion{BootstrapperVersion: bootstrapperVersion, BootstrapperType: constants.SKUBA}
}

func NewBootstrapperVersion(bootstrapperType string, version string) BootstrapperVersioner {
	switch bootstrapperType {
	case constants.KUBEADM:
		return &KubeadmBootstrapperVersion{BootstrapperVersion: version, BootstrapperType: bootstrapperType}

	case constants.K3S:
		return &K3sBootstrapperVersion{BootstrapperVersion: version, BootstrapperType: bootstrapperType}

	case constants.SKUBA:
		return &SkubaBootstrapperVersion{BootstrapperVersion: version, BootstrapperType: bootstrapperType}
	}

	return nil
}

func (k *KubeadmBootstrapperVersion) Type() string {
	return k.BootstrapperType
}

func (k *KubeadmBootstrapperVersion) LocalVersionFile() string {
	return path.Join(BootstrapperRootDir, config.TagVersion, k.BootstrapperType, k.BootstrapperVersion+".yaml")
}

func (k *KubeadmBootstrapperVersion) Version() string {
	return k.BootstrapperVersion
}

func (k *K3sBootstrapperVersion) Type() string {
	return k.BootstrapperType
}

func (k *K3sBootstrapperVersion) LocalVersionFile() string {
	return path.Join(BootstrapperRootDir, config.TagVersion, k.BootstrapperType, k.BootstrapperVersion+".yaml")
}

func (k *K3sBootstrapperVersion) Version() string {
	return k.BootstrapperVersion
}

func (s *SkubaBootstrapperVersion) Type() string {
	return s.BootstrapperType
}

func (s *SkubaBootstrapperVersion) LocalVersionFile() string {
	return path.Join(BootstrapperRootDir, config.TagVersion, s.BootstrapperType, s.BootstrapperVersion+".yaml")
}

func (s *SkubaBootstrapperVersion) Version() string {
	return s.BootstrapperVersion
}

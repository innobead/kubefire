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

type KubeadmBootstrapperVersion struct {
	BootstrapperVersion string `json:"version"`
	BootstrapperType    string `json:"type"`
	CrictlVersion       string `json:"crictl_version"`
	KubeReleaseVersion  string `json:"kube_release_version"`
}

type K3sBootstrapperVersion struct {
	BootstrapperVersion string `json:"version"`
	BootstrapperType    string `json:"type"`
}

type RKEBootstrapperVersion struct {
	BootstrapperVersion string   `json:"version"`
	BootstrapperType    string   `json:"type"`
	KubernetesVersions  []string `json:"kubernetes_versions"`
}

type RKE2BootstrapperVersion struct {
	BootstrapperVersion string `json:"version"`
	BootstrapperType    string `json:"type"`
}

func NewBootstrapperVersion(bootstrapperType string, version string) BootstrapperVersioner {
	switch bootstrapperType {
	case constants.KUBEADM:
		return &KubeadmBootstrapperVersion{BootstrapperVersion: version, BootstrapperType: bootstrapperType}
	case constants.K3S:
		return &K3sBootstrapperVersion{BootstrapperVersion: version, BootstrapperType: bootstrapperType}
	case constants.RKE:
		return &RKEBootstrapperVersion{BootstrapperVersion: version, BootstrapperType: bootstrapperType}
	case constants.RKE2:
		return &RKE2BootstrapperVersion{BootstrapperVersion: version, BootstrapperType: bootstrapperType}
	}

	return nil
}

func NewKubeadmBootstrapperVersion(bootstrapperVersion string, crictlVersion string, kubeReleaseVersion string) *KubeadmBootstrapperVersion {
	return &KubeadmBootstrapperVersion{
		BootstrapperVersion: bootstrapperVersion,
		BootstrapperType:    constants.KUBEADM,
		CrictlVersion:       crictlVersion,
		KubeReleaseVersion:  kubeReleaseVersion,
	}
}

func NewK3sBootstrapperVersion(bootstrapperVersion string) *K3sBootstrapperVersion {
	return &K3sBootstrapperVersion{
		BootstrapperVersion: bootstrapperVersion,
		BootstrapperType:    constants.K3S,
	}
}

func NewRKEBootstrapperVersion(bootstrapperVersion string, kubernetesVersions []string) *RKEBootstrapperVersion {
	return &RKEBootstrapperVersion{
		BootstrapperVersion: bootstrapperVersion,
		BootstrapperType:    constants.RKE,
		KubernetesVersions:  kubernetesVersions,
	}
}

func NewRKE2BootstrapperVersion(bootstrapperVersion string) *RKE2BootstrapperVersion {
	return &RKE2BootstrapperVersion{
		BootstrapperVersion: bootstrapperVersion,
		BootstrapperType:    constants.RKE2,
	}
}

func (k *KubeadmBootstrapperVersion) Display() string {
	return k.Version()
}

func (k *KubeadmBootstrapperVersion) Type() string {
	return k.BootstrapperType
}

func (k *KubeadmBootstrapperVersion) LocalVersionFile() string {
	return path.Join(
		BootstrapperRootDir,
		config.TagVersion,
		k.BootstrapperType,
		k.BootstrapperVersion+".yaml",
	)
}

func (k *KubeadmBootstrapperVersion) Version() string {
	return k.BootstrapperVersion
}

func (k *K3sBootstrapperVersion) Display() string {
	return k.Version()
}

func (k *K3sBootstrapperVersion) Type() string {
	return k.BootstrapperType
}

func (k *K3sBootstrapperVersion) LocalVersionFile() string {
	return path.Join(
		BootstrapperRootDir,
		config.TagVersion,
		k.BootstrapperType,
		k.BootstrapperVersion+".yaml",
	)
}

func (k *K3sBootstrapperVersion) Version() string {
	return k.BootstrapperVersion
}

func (s *RKEBootstrapperVersion) Display() string {
	return fmt.Sprintf(
		"%s (%s)",
		s.Version(),
		strings.Join(s.KubernetesVersions, ", "),
	)
}

func (s *RKEBootstrapperVersion) Type() string {
	return s.BootstrapperType
}

func (s *RKEBootstrapperVersion) LocalVersionFile() string {
	return path.Join(
		BootstrapperRootDir,
		config.TagVersion,
		s.BootstrapperType,
		s.BootstrapperVersion+".yaml",
	)
}

func (s *RKEBootstrapperVersion) Version() string {
	return s.BootstrapperVersion
}

func (r *RKE2BootstrapperVersion) Type() string {
	return r.BootstrapperType
}

func (r *RKE2BootstrapperVersion) LocalVersionFile() string {
	return path.Join(
		BootstrapperRootDir,
		config.TagVersion,
		r.BootstrapperType,
		r.BootstrapperVersion+".yaml",
	)
}

func (r *RKE2BootstrapperVersion) Version() string {
	return r.BootstrapperVersion
}

func (r *RKE2BootstrapperVersion) Display() string {
	return r.Version()
}

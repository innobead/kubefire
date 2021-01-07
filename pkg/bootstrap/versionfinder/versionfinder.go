package versionfinder

import (
	"github.com/innobead/kubefire/pkg/constants"
	"github.com/innobead/kubefire/pkg/data"
)

type Finder interface {
	GetVersionsAfterVersion(afterVersion data.Version) ([]*data.Version, error)
	GetLatestVersion() (*data.Version, error)
}

type BaseVersionFinder struct {
	bootstrapperType string
}

func New(bootstrapperType string) Finder {
	switch bootstrapperType {
	case constants.KUBEADM:
		return NewKubeadmVersionFinder()
	case constants.K3S:
		return NewK3sVersionFinder()
	case constants.RKE:
		return NewRKEVersionFinder()
	case constants.RKE2:
		return NewRKE2VersionFinder()
	case constants.K0s:
		return NewK0sVersionFinder()
	}

	return nil
}

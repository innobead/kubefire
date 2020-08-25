package versionfinder

import (
	"github.com/innobead/kubefire/pkg/constants"
	"github.com/innobead/kubefire/pkg/data"
	"github.com/sirupsen/logrus"
)

type SkubaVersionFinder struct {
	BaseVersionFinder
}

func NewSkubaVersionFinder() *SkubaVersionFinder {
	return &SkubaVersionFinder{
		BaseVersionFinder: BaseVersionFinder{
			constants.SKUBA,
		},
	}
}

func (s *SkubaVersionFinder) GetVersionsAfterVersion(afterVersion data.Version) ([]*data.Version, error) {
	logrus.WithField("bootstrapper", s.bootstrapperType).Debugln("getting the released versions info")

	return []*data.Version{
		data.ParseVersion("v1.4.1"), // CaaSP 4.2.2
		data.ParseVersion("v1.3.5"), // CaaSP 4.2.1
	}, nil
}

func (s *SkubaVersionFinder) GetLatestVersion() (*data.Version, error) {
	logrus.WithField("bootstrapper", s.bootstrapperType).Debugln("getting the latest released version info")

	versions, _ := s.GetVersionsAfterVersion(data.Version{})
	return versions[0], nil
}

func (s *SkubaVersionFinder) HasPatchVersion(version string) bool {
	versions, _ := s.GetVersionsAfterVersion(data.Version{})

	for _, v := range versions {
		if v.String() == version {
			return true
		}
	}

	return false
}

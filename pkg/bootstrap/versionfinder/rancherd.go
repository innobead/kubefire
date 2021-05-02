package versionfinder

import (
	"github.com/innobead/kubefire/internal/config"
	"github.com/innobead/kubefire/pkg/constants"
	"github.com/innobead/kubefire/pkg/data"
	"github.com/innobead/kubefire/pkg/util"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type RancherdVersionFinder struct {
	BaseVersionFinder
	githubInfoer *util.GithubInfoer
	owner        string
	repo         string
}

func NewRancherdVersionFinder() *RancherdVersionFinder {
	return &RancherdVersionFinder{
		BaseVersionFinder: BaseVersionFinder{
			constants.RANCHERD,
		},
		githubInfoer: util.NewGithubInfoer(config.GithubToken),
		owner:        "rancher",
		repo:         "rancher",
	}
}

func (k *RancherdVersionFinder) GetVersionsAfterVersion(afterVersion data.Version) ([]*data.Version, error) {
	afterVersion = *data.ParseVersion("v2.5.100") // rancherd supported from Rancher 2.5
	logrus.WithField("bootstrapper", k.bootstrapperType).Debugln("getting the released versions info")

	return k.githubInfoer.GetVersionsAfterVersion(afterVersion, k.owner, k.repo, data.SupportedMinorVersionCount)
}

func (k *RancherdVersionFinder) GetLatestVersion() (*data.Version, error) {
	logrus.WithField("bootstrapper", k.bootstrapperType).Debugln("getting the latest released version info")

	versions, err := k.GetVersionsAfterVersion(data.Version{})
	if err != nil {
		return nil, err
	}

	if len(versions) == 0 {
		return nil, errors.New("no latest version found")
	}

	return versions[0], nil
}

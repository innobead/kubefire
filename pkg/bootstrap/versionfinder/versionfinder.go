package versionfinder

import (
	"fmt"
	"github.com/innobead/kubefire/pkg/constants"
	"github.com/innobead/kubefire/pkg/data"
	"github.com/innobead/kubefire/pkg/util"
	"github.com/sirupsen/logrus"
	"net/http"
)

const GithubReleaseUrlTemplate = "https://github.com/%s/%s/releases/tag/%s"

type Finder interface {
	GetVersionsAfterVersion(afterVersion data.Version) ([]*data.Version, error)
	GetLatestVersion() (*data.Version, error)
	HasPatchVersion(version string) bool
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

	case constants.SKUBA:
		return NewSkubaVersionFinder()
	}

	return nil
}

func hasPatchVersionGithub(owner, repo, tag, bootstrapperType string) bool {
	url := fmt.Sprintf(GithubReleaseUrlTemplate, owner, repo, tag)
	_, resp, err := util.HttpGet(url)
	if err != nil {
		logrus.WithField("bootstrapper", bootstrapperType).WithError(err).Errorf("failed to check %s", url)
		return false
	}

	return resp.StatusCode == http.StatusOK
}

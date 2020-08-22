package versionfinder

import (
	"encoding/json"
	interr "github.com/innobead/kubefire/internal/error"
	"github.com/innobead/kubefire/pkg/constants"
	"github.com/innobead/kubefire/pkg/data"
	"github.com/innobead/kubefire/pkg/util"
	"github.com/sirupsen/logrus"
	"regexp"
	"strings"
)

const K3sChannelInfoUrl = "https://update.k3s.io/v1-release/channels"

type K3sVersionFinder struct {
	BaseVersionFinder
}

func NewK3sVersionFinder() *K3sVersionFinder {
	return &K3sVersionFinder{BaseVersionFinder{
		constants.K3S,
	}}
}

func (k *K3sVersionFinder) GetVersionsAfterVersion(afterVersion data.Version) ([]*data.Version, error) {
	logrus.WithField("bootstrapper", k.bootstrapperType).Infoln("getting the released versions info")

	var versions []*data.Version

	body, err := util.HttpGet(K3sChannelInfoUrl)
	if err != nil {
		return nil, err
	}

	versionsInfoMap := map[string]interface{}{}
	decoder := json.NewDecoder(strings.NewReader(body))
	if err := decoder.Decode(&versionsInfoMap); err != nil {
		return nil, err
	}

	if _, ok := versionsInfoMap["data"]; ok {
		re := regexp.MustCompile(`^v\d+\.\d+$`)

		for _, versionInfo := range versionsInfoMap["data"].([]interface{}) {
			info := versionInfo.(map[string]interface{})

			if id, ok := info["id"]; ok {
				if ok := re.MatchString(id.(string)); ok {
					version := strings.ReplaceAll(info["latest"].(string), "+k3s1", "")
					versions = append(
						versions,
						data.ParseVersion(version),
					)
				}
			}
		}
	}

	if len(versions) > 0 {
		return versions, nil
	}

	return nil, interr.NodeNotFoundError
}

func (k *K3sVersionFinder) GetLatestVersion() (*data.Version, error) {
	logrus.WithField("bootstrapper", k.bootstrapperType).Infof("getting the latest released version info")

	body, err := util.HttpGet(K3sChannelInfoUrl)
	if err != nil {
		return nil, err
	}

	versionsInfoMap := map[string]interface{}{}
	decoder := json.NewDecoder(strings.NewReader(body))
	if err := decoder.Decode(&versionsInfoMap); err != nil {
		return nil, err
	}

	if _, ok := versionsInfoMap["data"]; ok {
		for _, versionInfo := range versionsInfoMap["data"].([]interface{}) {
			info := versionInfo.(map[string]interface{})

			if id, ok := info["id"]; ok && id == "latest" {
				version := strings.ReplaceAll(info["latest"].(string), "+k3s1", "")
				return data.ParseVersion(version), nil
			}
		}
	}

	return nil, interr.NotFoundError
}

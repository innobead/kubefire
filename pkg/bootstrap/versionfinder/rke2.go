package versionfinder

import (
	"encoding/json"
	interr "github.com/innobead/kubefire/internal/error"
	"github.com/innobead/kubefire/pkg/constants"
	"github.com/innobead/kubefire/pkg/data"
	"github.com/innobead/kubefire/pkg/util"
	"github.com/sirupsen/logrus"
	"regexp"
	"sort"
	"strings"
)

const RKE2ChannelInfoUrl = "https://update.rke2.io/v1-release/channels"

type RKE2VersionFinder struct {
	BaseVersionFinder
}

func NewRKE2VersionFinder() *RKE2VersionFinder {
	return &RKE2VersionFinder{BaseVersionFinder{
		constants.RKE2,
	}}
}

func (k *RKE2VersionFinder) GetVersionsAfterVersion(afterVersion data.Version) ([]*data.Version, error) {
	logrus.WithField("bootstrapper", k.bootstrapperType).Debugln("getting the released versions info")

	versionsInfoMap, err := getRKE2VersionsInfo()
	if err != nil {
		return nil, err
	}

	var versions []*data.Version

	if _, ok := versionsInfoMap["data"]; ok {
		re := regexp.MustCompile(`^v\d+\.\d+$`)

		for _, versionInfo := range versionsInfoMap["data"].([]interface{}) {
			info := versionInfo.(map[string]interface{})

			if id, ok := info["id"]; ok {
				if id == "latest" || re.MatchString(id.(string)) {
					versions = append(
						versions,
						data.ParseVersion(info["latest"].(string)),
					)
				}
			}
		}
	}

	sort.Slice(versions, func(i, j int) bool {
		v1 := versions[i]
		v2 := versions[j]

		return v1.Compare(v2) >= 0
	})

	if len(versions) > 0 {
		return versions, nil
	}

	return nil, interr.NodeNotFoundError
}

func (k *RKE2VersionFinder) GetLatestVersion() (*data.Version, error) {
	logrus.WithField("bootstrapper", k.bootstrapperType).Debugln("getting the latest released version info")

	versionsInfoMap, err := getRKE2VersionsInfo()
	if err != nil {
		return nil, err
	}

	if _, ok := versionsInfoMap["data"]; ok {
		for _, versionInfo := range versionsInfoMap["data"].([]interface{}) {
			info := versionInfo.(map[string]interface{})

			if id, ok := info["id"]; ok && id == "latest" {
				return data.ParseVersion(info["latest"].(string)), nil
			}
		}
	}

	return nil, interr.NotFoundError
}

func getRKE2VersionsInfo() (map[string]interface{}, error) {
	body, _, err := util.HttpGet(RKE2ChannelInfoUrl)
	if err != nil {
		return nil, err
	}

	versionsInfoMap := map[string]interface{}{}
	decoder := json.NewDecoder(strings.NewReader(body))
	if err := decoder.Decode(&versionsInfoMap); err != nil {
		return nil, err
	}

	return versionsInfoMap, nil
}

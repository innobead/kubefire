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
	logrus.WithField("bootstrapper", k.bootstrapperType).Debugln("getting the released versions info")

	versionsInfoMap, err := getK3sVersionsInfo()
	if err != nil {
		return nil, err
	}

	var versions []*data.Version

	if _, ok := versionsInfoMap["data"]; ok {
		re := regexp.MustCompile(`^v\d+\.\d+$`)

		for _, versionInfo := range versionsInfoMap["data"].([]interface{}) {
			info := versionInfo.(map[string]interface{})

			if id, ok := info["id"]; ok {
				if ok := re.MatchString(id.(string)); ok {
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

func (k *K3sVersionFinder) GetLatestVersion() (*data.Version, error) {
	logrus.WithField("bootstrapper", k.bootstrapperType).Debugln("getting the latest released version info")

	versionsInfoMap, err := getK3sVersionsInfo()
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

func (k *K3sVersionFinder) HasPatchVersion(version string) bool {

	return hasPatchVersionGithub("rancher", "k3s", version+"+k3s1", k.bootstrapperType)
}

func getK3sVersionsInfo() (map[string]interface{}, error) {
	body, _, err := util.HttpGet(K3sChannelInfoUrl)
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

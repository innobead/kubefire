package versionfinder

import (
	"encoding/json"
	"fmt"
	interr "github.com/innobead/kubefire/internal/error"
	"github.com/innobead/kubefire/pkg/constants"
	"github.com/innobead/kubefire/pkg/data"
	"github.com/innobead/kubefire/pkg/util"
	"github.com/sirupsen/logrus"
	"github.com/thoas/go-funk"
	"sort"
	"strings"
)

const (
	RKEDefaultBranchUrl = "https://api.github.com/repos/rancher/kontainer-driver-metadata"
	RKEVersionInfoUrl   = "https://raw.githubusercontent.com/rancher/kontainer-driver-metadata/%s/data/data.json"
)

type RKEVersionFinder struct {
	BaseVersionFinder
}

func NewRKEVersionFinder() *RKEVersionFinder {
	return &RKEVersionFinder{BaseVersionFinder{
		constants.RKE,
	}}
}

func (k *RKEVersionFinder) GetVersionsAfterVersion(afterVersion data.Version) ([]*data.Version, error) {
	logrus.WithField("bootstrapper", k.bootstrapperType).Debugln("getting the released versions info")

	latestVersion := data.ParseVersion("v1.2.3")
	k8sVersions, err := k.getSupportedK8sVersions()
	if err != nil {
		return nil, err
	}

	k8sVersionNames := funk.Map(k8sVersions, func(v *data.Version) string {
		return v.String()
	}).([]string)
	latestVersion.ExtraMeta = map[string]interface{}{
		"kubernetes_version": k8sVersionNames,
	}

	return []*data.Version{latestVersion}, nil
}

func (k *RKEVersionFinder) GetLatestVersion() (*data.Version, error) {
	logrus.WithField("bootstrapper", k.bootstrapperType).Debugln("getting the latest released version info")

	return data.ParseVersion("v1.2.3"), nil
}

func (k *RKEVersionFinder) HasPatchVersion(version string) bool {
	v := data.ParseVersion(version)
	return v.Patch != ""
}

func (k *RKEVersionFinder) getLatestSupportedK8sVersion() (*data.Version, error) {
	versionsInfoMap, err := k.getRKEVersionsInfo()
	if err != nil {
		return nil, err
	}

	version, ok := versionsInfoMap["RKEDefaultK8sVersions"].(map[string]interface{})["default"]
	if ok {
		return data.ParseVersion(version.(string)), nil
	}

	return nil, interr.NotFoundError
}

func (k *RKEVersionFinder) getSupportedK8sVersions() ([]*data.Version, error) {
	versionsInfoMap, err := k.getRKEVersionsInfo()
	if err != nil {
		return nil, err
	}

	systemImages := versionsInfoMap["K8sVersionRKESystemImages"].(map[string]interface{})
	var versions []*data.Version
	for version := range systemImages {
		// example: v1.19.4-rancher1-2
		if strings.Contains(version, "-rancher1-") {
			versions = append(versions, data.ParseVersion(version))
		}
	}

	// data.SupportedMinorVersionCount
	if len(versions) == 0 {
		return nil, interr.NodeNotFoundError
	}

	sort.Slice(versions, func(i, j int) bool {
		v1 := versions[i]
		v2 := versions[j]

		return v1.Compare(v2) >= 0
	})

	latestVersion, err := k.getLatestSupportedK8sVersion()
	if err != nil {
		return nil, err
	}

	filteredVersions := []*data.Version{latestVersion}
	foundCount := data.SupportedMinorVersionCount - 1

	for _, version := range versions {
		if foundCount == 0 {
			break
		}

		lastVersion := filteredVersions[len(filteredVersions)-1]
		if version.MajorMinorString() == lastVersion.MajorMinorString() {
			continue
		}

		filteredVersions = append(filteredVersions, version)
		foundCount--
	}

	return filteredVersions, nil
}

func (k *RKEVersionFinder) getRKEVersionsInfo() (map[string]interface{}, error) {
	defaultBranch, err := k.getRKEDefaultBranch()
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf(RKEVersionInfoUrl, defaultBranch)
	body, _, err := util.HttpGet(url)
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

func (k *RKEVersionFinder) getRKEDefaultBranch() (string, error) {
	body, _, err := util.HttpGet(RKEDefaultBranchUrl)
	if err != nil {
		return "", err
	}

	branchInfoMap := map[string]interface{}{}
	decoder := json.NewDecoder(strings.NewReader(body))
	if err := decoder.Decode(&branchInfoMap); err != nil {
		return "", err
	}

	return branchInfoMap["default_branch"].(string), nil
}

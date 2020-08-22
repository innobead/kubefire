package config

import (
	"regexp"
)

var (
	BuildVersion = ""
	TagVersion   = "master"
)

//GetTagVersionForDownloadScript get the tag version to use in the URL of download scripts. If not from release tag, the return will be master.
func GetTagVersionForDownloadScript(version string) string {
	if !IsReleasedTagVersion(version) {
		return "master"
	}

	return version
}

//IsReleasedTagVersion determine if version is a release tag (format: v\d+.\d+.\d+).
func IsReleasedTagVersion(version string) bool {
	return regexp.MustCompile(`^v\d+\.\d+\.\d$`).MatchString(version)
}

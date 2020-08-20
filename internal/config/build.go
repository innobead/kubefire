package config

import "strings"

var (
	BuildVersion = ""
	TagVersion   = "master"
)

func GetDownloadTagVersion() string {
	if strings.Contains(TagVersion, "dirty") {
		return "master"
	}

	return TagVersion
}

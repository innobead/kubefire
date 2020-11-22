package data

import (
	"fmt"
	"github.com/Masterminds/semver/v3"
	"github.com/sirupsen/logrus"
	"strconv"
)

const SupportedMinorVersionCount = 3

type SubVersionType string

type Version struct {
	Major      SubVersionType
	Minor      SubVersionType
	Patch      SubVersionType
	PRERELEASE SubVersionType
	METADATA   SubVersionType
	internal   *semver.Version
	ExtraMeta  map[string]interface{}
}

func (s SubVersionType) ToInt() int {
	value, err := strconv.Atoi(string(s))
	if err != nil {
		return 0
	}

	return value
}

func (v *Version) String() string {
	return "v" + v.internal.String()
}

func (v *Version) MajorString() string {
	return fmt.Sprintf("v%s", v.Major)
}

func (v *Version) MajorMinorString() string {
	return fmt.Sprintf("v%s.%s", v.Major, v.Minor)
}

func (v *Version) Compare(version *Version) int {
	return v.internal.Compare(version.internal)
}

func ParseVersion(version string) *Version {
	v, err := semver.NewVersion(version)
	if err != nil {
		logrus.Errorf("failed to parse semantic version %s: %v", version, err)
		return nil
	}

	return &Version{
		Major:      SubVersionType(strconv.FormatUint(v.Major(), 10)),
		Minor:      SubVersionType(strconv.FormatUint(v.Minor(), 10)),
		Patch:      SubVersionType(strconv.FormatUint(v.Patch(), 10)),
		PRERELEASE: SubVersionType(v.Prerelease()),
		METADATA:   SubVersionType(v.Metadata()),
		internal:   v,
	}
}

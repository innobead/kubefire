package data

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

const SupportedMinorVersionCount = 3

type SubVersionType string

type Version struct {
	Major SubVersionType
	Minor SubVersionType
	Patch SubVersionType
}

func (s SubVersionType) ToInt() int {
	value, err := strconv.Atoi(string(s))
	if err != nil {
		return 0
	}

	return value
}

func (v *Version) String() string {
	return fmt.Sprintf("v%s.%s.%s", v.Major, v.Minor, v.Patch)
}

func (v *Version) MajorString() string {
	return fmt.Sprintf("v%s", v.Major)
}

func (v *Version) MajorMinorString() string {
	return fmt.Sprintf("v%s.%s", v.Major, v.Minor)
}

func (v *Version) Compare(version *Version) int {
	switch {
	case version == nil || v.Major.ToInt() > version.Major.ToInt() || v.Major.ToInt() == version.Major.ToInt() && v.Minor.ToInt() > version.Minor.ToInt() || v.Major.ToInt() == version.Major.ToInt() && v.Minor == version.Minor && v.Patch.ToInt() > version.Patch.ToInt():
		return 1
	case v.Major == version.Major && v.Minor == version.Minor && v.Patch == version.Patch:
		return 0
	default:
		return -1
	}
}

func ParseVersion(version string) *Version {
	pattern := regexp.MustCompile(`^v(\d+)\.(\d+)(\.\d+)?$`)

	submatch := pattern.FindStringSubmatch(version)
	if len(submatch) == 0 {
		return nil
	}

	var v = Version{
		Major: SubVersionType(submatch[1]),
		Minor: SubVersionType(submatch[2]),
	}
	if len(submatch) == 4 {
		v.Patch = SubVersionType(strings.TrimPrefix(submatch[3], "."))
	}

	return &v
}

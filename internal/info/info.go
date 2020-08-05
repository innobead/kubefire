package info

import (
	"fmt"
	"github.com/innobead/kubefire/internal/config"
	"github.com/sirupsen/logrus"
	"os/exec"
	"regexp"
	"strings"
)

type RuntimeVersionInfo struct {
	Containerd string
	Ignite     string
	Cni        string
	Runc       string
}

func CurrentRuntimeVersionInfo() *RuntimeVersionInfo {
	return &RuntimeVersionInfo{
		Containerd: RuntimeVersion("containerd --version", `(v\d+.\d+.\d+)`, config.ContainerdVersion),
		Ignite:     RuntimeVersion("ignite version -o short", `(v\d+.\d+.\d+)`, config.IgniteVersion),
		Cni:        RuntimeVersion("/opt/cni/bin/loopback", `(v\d+.\d+.\d+)`, config.CniVersion),
		Runc:       RuntimeVersion("runc -v", `runc version (\d+.\d+.\d+[-\w\d]+)`, strings.TrimPrefix(config.RuncVersion, "v")),
	}
}

func RuntimeVersion(cmdline, regexPattern, expectedVersion string) string {
	expectedVersionStr := fmt.Sprintf("(expected: %s)", expectedVersion)

	args := strings.Split(cmdline, " ")
	cmd := exec.Command(args[0], args[1:]...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		logrus.Debugf("failed to run %s", cmdline)
		return expectedVersionStr
	}

	compile, err := regexp.Compile(regexPattern)
	if err != nil {
		logrus.Debugf("failed to compile regular expression pattern `%s`", regexPattern)
		return expectedVersionStr
	}

	submatchs := compile.FindStringSubmatch(string(output))
	if len(submatchs) != 2 {
		logrus.Debugf("failed to compile regular expression pattern `%s`", regexPattern)
		return expectedVersionStr
	}

	version := submatchs[1]
	if version == expectedVersion {
		return version
	}

	return fmt.Sprintf("%s %s", version, expectedVersionStr)
}

func (r *RuntimeVersionInfo) ExpectedEnvVars() []string {
	return []string{
		fmt.Sprintf("CONTAINERD_VERSION=%s", config.ContainerdVersion),
		fmt.Sprintf("IGNITE_VERION=%s", config.IgniteVersion),
		fmt.Sprintf("CNI_VERSION=%s", config.CniVersion),
		fmt.Sprintf("RUNC_VERSION=%s", config.RuncVersion),
	}
}

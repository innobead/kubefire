package cmd

import (
	"fmt"
	gocni "github.com/containerd/go-cni"
	"github.com/innobead/kubefire/internal/config"
	"github.com/sirupsen/logrus"
	"os/exec"
	"reflect"
	"regexp"
	"strings"
)

const (
	CniConfigDir = "/etc/cni/net.d"
	CniBinDir    = "/opt/cni/bin"
)

type PrerequisiteMatcher interface {
	Matched() bool
}

type PrerequisitesInfos struct {
	Containerd PrerequisitesInfo
	Ignite     PrerequisitesInfo
	Cni        PrerequisitesInfo
	Runc       PrerequisitesInfo
	CniPlugin  PrerequisitesInfo
}

type PrerequisitesInfo struct {
	InstalledVersion string
	ExpectedVersion  string
}

func (r *PrerequisitesInfos) Matched() bool {
	value := reflect.ValueOf(r).Elem()
	numField := value.NumField()

	for i := 0; i < numField; i++ {
		field := value.Field(i)

		if field.Kind() == reflect.Struct {
			if matcher, ok := field.Addr().Interface().(PrerequisiteMatcher); ok {
				if !matcher.Matched() {
					return false
				}
			}
		}
	}

	return true
}

func (p PrerequisitesInfo) String() string {
	if p.Matched() {
		return p.InstalledVersion
	}

	return fmt.Sprintf("%s (expected: %s)", p.InstalledVersion, p.ExpectedVersion)
}

func (p PrerequisitesInfo) Matched() bool {
	return p.InstalledVersion == p.ExpectedVersion
}

func CurrentPrerequisitesInfos() *PrerequisitesInfos {
	return &PrerequisitesInfos{
		Containerd: prerequisiteVersion("containerd --version", `(v\d+.\d+.\d+)`, config.ContainerdVersion),
		Ignite:     prerequisiteVersion("ignite version -o short", `(v\d+.\d+.\d+)`, config.IgniteVersion),
		Cni:        prerequisiteVersion("/opt/cni/bin/loopback", `(v\d+.\d+.\d+)`, config.CniVersion),
		Runc:       prerequisiteVersion("runc -v", `runc version (\d+.\d+.\d+[-\w\d]+)`, strings.TrimPrefix(config.RuncVersion, "v")),
		CniPlugin:  cniVersion("0.4.0/kubefire-cni-bridge"),
	}
}

func prerequisiteVersion(cmdline, regexPattern, expectedVersion string) PrerequisitesInfo {
	info := PrerequisitesInfo{
		"",
		expectedVersion,
	}

	args := strings.Split(cmdline, " ")
	cmd := exec.Command(args[0], args[1:]...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		logrus.Debugf("failed to run %s", cmdline)
		return info
	}

	compile, err := regexp.Compile(regexPattern)
	if err != nil {
		logrus.Debugf("failed to compile regular expression pattern `%s`", regexPattern)
		return info
	}

	submatchs := compile.FindStringSubmatch(string(output))
	if len(submatchs) != 2 {
		logrus.Debugf("failed to compile regular expression pattern `%s`", regexPattern)
		return info
	}
	info.InstalledVersion = submatchs[1]

	return info
}

func cniVersion(expectedVersion string) PrerequisitesInfo {
	info := PrerequisitesInfo{
		"",
		expectedVersion,
	}

	client, err := gocni.New(
		gocni.WithMinNetworkCount(2),
		gocni.WithPluginConfDir(CniConfigDir),
		gocni.WithPluginDir([]string{CniBinDir}),
	)
	if err != nil {
		logrus.Errorf("failed to initialize cni library: %v", err)
		return info
	}

	if err := client.Load(gocni.WithLoNetwork, gocni.WithDefaultConf); err != nil {
		logrus.Errorf("failed to load cni configuration: %v", err)
		return info
	}

	if len(client.GetConfig().Networks) != 2 {
		logrus.Errorf("failed to load cni configuration because network configuration count is not 2")
		return info
	}

	n := client.GetConfig().Networks[1]
	info.InstalledVersion = fmt.Sprintf(
		"%s/%s",
		n.Config.CNIVersion,
		n.Config.Name,
	)

	return info
}

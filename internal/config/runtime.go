package config

import "fmt"

var (
	LogLevel     string
	Output       string
	Bootstrapper string
	GithubToken  string
)

var (
	ContainerdVersion string
	IgniteVersion     string
	CniVersion        string
	RuncVersion       string
)

type EnvVars []string

func (e EnvVars) String() string {
	var str string

	for _, envVar := range e {
		str += " " + envVar
	}

	return str
}

func ExpectedPrerequisiteVersionsEnvVars() EnvVars {
	return []string{
		fmt.Sprintf("KUBEFIRE_VERSION=%s", TagVersion),
		fmt.Sprintf("CONTAINERD_VERSION=%s", ContainerdVersion),
		fmt.Sprintf("IGNITE_VERION=%s", IgniteVersion),
		fmt.Sprintf("CNI_VERSION=%s", CniVersion),
		fmt.Sprintf("RUNC_VERSION=%s", RuncVersion),
	}
}

func KubeadmVersionsEnvVars(k8sVersion, kubeReleaseVersion, crictlVersion string) EnvVars {
	preVersions := ExpectedPrerequisiteVersionsEnvVars()

	preVersions = append(preVersions, fmt.Sprintf("KUBE_VERSION=%s", k8sVersion)) // https://dl.k8s.io/release/stable.txt
	preVersions = append(preVersions, fmt.Sprintf("KUBE_RELEASE_VERSION=%s", kubeReleaseVersion))
	preVersions = append(preVersions, fmt.Sprintf("CRICTL_VERSION=%s", crictlVersion))

	return preVersions
}

func K3sVersionsEnvVars(version string) EnvVars {
	return []string{
		fmt.Sprintf("K3S_VERSION=%s", version),
		// Need to use this option to forcibly ask k3s installer to install the specific version. Otherwise, it will choose a stable version from https://update.k3s.io/v1-release/channels.
		fmt.Sprintf("INSTALL_K3S_VERSION=%s", version),
	}
}

func RKEVersionsEnvVars(version string) EnvVars {
	return []string{
		fmt.Sprintf("RKE_VERSION=%s", version),
	}
}

func RKE2VersionsEnvVars(version string, configContent string) EnvVars {
	return []string{
		fmt.Sprintf("RKE2_VERSION=%s", version),
		fmt.Sprintf(`RKE2_CONFIG="%s"`, configContent),
	}
}

func RancherdVersionsEnvVars(version string, configContent string) EnvVars {
	return []string{
		fmt.Sprintf("RANCHERD_VERSION=%s", version),
		fmt.Sprintf(`RKE2_CONFIG="%s"`, configContent),
	}
}

func K0sVersionsEnvVars(version string, configContent string, cmdOpts string) EnvVars {
	return []string{
		fmt.Sprintf("K0S_VERSION=%s", version),
		fmt.Sprintf(`K0S_CONFIG="%s"`, configContent),
		fmt.Sprintf(`K0S_CMD_OPTS="%s"`, cmdOpts),
	}
}

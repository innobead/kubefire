package config

import "fmt"

var (
	LogLevel     string
	Output       string
	Bootstrapper string
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

func KubeadmVersionsEnvVars() EnvVars { // TODO planned to support selective version
	preVersions := ExpectedPrerequisiteVersionsEnvVars()

	preVersions = append(preVersions, fmt.Sprintf("KUBE_VERSION=%s", "v1.18.8")) // https://dl.k8s.io/release/stable.txt
	preVersions = append(preVersions, fmt.Sprintf("KUBE_RELEASE_VERSION=%s", "v0.3.4"))
	preVersions = append(preVersions, fmt.Sprintf("CRICTL_VERSION=%s", "v1.18.0"))

	return preVersions
}

func K3sVersionsEnvVars() EnvVars { // TODO planned to support selective version
	return []string{
		fmt.Sprintf("K3S_VERSION=%s", "v1.18.8"),
		// Need to use this option to forcibly ask k3s installer to install the specific version. Otherwise, it will choose a stable version from https://update.k3s.io/v1-release/channels.
		fmt.Sprintf("INSTALL_K3S_VERSION=%s", "v1.18.8+k3s1"),
	}
}

func SkubaVersionsEnvVars() EnvVars { // TODO planned to support selective version
	return []string{
		fmt.Sprintf("SKUBA_VERSION=%s", "v1.4.1"), // SUSE CaaSP 4.2.2
	}
}

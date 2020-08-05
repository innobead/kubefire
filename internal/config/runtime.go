package config

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

//
//CONTAINERD_VERSION=${CONTAINERD_VERSION:-"v1.3.4"}
//IGNITE_VERION=${IGNITE_VERION:-"v0.7.1"}
//CNI_VERSION=${CNI_VERSION:-"v0.8.6"}
//RUNC_VERSION=${RUNC_VERSION:-"v1.0.0-rc91"}

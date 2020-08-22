package config

type Manager interface {
	SaveCluster(cluster *Cluster) error
	DeleteCluster(cluster *Cluster) error
	GetCluster(name string) (*Cluster, error)
	ListClusters() ([]*Cluster, error)

	SaveBootstrapperVersions(latestVersion BootstrapperVersioner, versions []BootstrapperVersioner) error
	GetBootstrapperVersions(latestVersion BootstrapperVersioner) ([]BootstrapperVersioner, error)
	DeleteBootstrapperVersions(latestVersion BootstrapperVersioner) error
}

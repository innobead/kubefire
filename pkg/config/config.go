package config

type Manager interface {
	SaveCluster(name string, cluster *Cluster) error
	DeleteCluster(name string) error
	GetCluster(name string) (*Cluster, error)
	ListClusters() ([]*Cluster, error)
}

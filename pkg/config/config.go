package config

type Manager interface {
	SaveCluster(cluster *Cluster) error
	DeleteCluster(cluster *Cluster) error
	GetCluster(name string) (*Cluster, error)
	ListClusters() ([]*Cluster, error)
}

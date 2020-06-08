package config

type Manager interface {
	Save(name string, cluster *Cluster) error
	Delete(name string) error
	Get(name string) (*Cluster, error)
	List() ([]*Cluster, error)
}

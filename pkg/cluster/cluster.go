package cluster

type Cluster struct {
}

type Error error

type Manager interface {
	Init() (*Cluster, Error)
	Deploy(cluster *Cluster) Error
	Destroy(cluster *Cluster) Error
	Get(name string) (*Cluster, Error)
	List() ([]*Cluster, Error)
}

type DefaultManager struct {
	// TODO use IgniteNodeManager
}

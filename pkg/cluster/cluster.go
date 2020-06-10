package cluster

import (
	"github.com/innobead/kubefire/pkg/bootstrap"
	pkgconfig "github.com/innobead/kubefire/pkg/config"
	"github.com/innobead/kubefire/pkg/data"
	"github.com/innobead/kubefire/pkg/node"
	"github.com/sirupsen/logrus"
)

type Manager interface {
	Init(cluster *pkgconfig.Cluster) error
	Create(name string) error
	Delete(name string, force bool) error
	Get(name string) (*data.Cluster, error)
	List() ([]*data.Cluster, error)
}

type DefaultManager struct {
	NodeManager   node.Manager
	Bootstrapper  bootstrap.Bootstrapper
	ConfigManager pkgconfig.Manager
}

func NewDefaultManager(nodeManager node.Manager, bootstrapper bootstrap.Bootstrapper, configManager pkgconfig.Manager) (Manager, error) {
	if nodeManager == nil {
		nodeManager = node.NewIgniteNodeManager()
	}

	if bootstrapper == nil {
		bootstrapper = bootstrap.NewKubeadmBootstrapper()
	}

	if configManager == nil {
		configManager = pkgconfig.NewLocalConfigManager()
	}

	return &DefaultManager{
		NodeManager:   nodeManager,
		Bootstrapper:  bootstrapper,
		ConfigManager: configManager,
	}, nil
}

func (d *DefaultManager) Init(cluster *pkgconfig.Cluster) error {
	if _, err := d.ConfigManager.GetCluster(cluster.Name); err == nil {
		logrus.Warnf("cluster (%s) config (%s) already exists", cluster.Name, pkgconfig.ClusterConfigFile(cluster.Name))
	}

	return d.ConfigManager.SaveCluster(cluster.Name, cluster)
}

func (d *DefaultManager) Create(name string) error {
	cluster, err := d.ConfigManager.GetCluster(name)
	if err != nil {
		return err
	}

	nodeTypeConfigs := map[node.Type]*pkgconfig.Node{
		node.Admin:  &cluster.Admin,
		node.Master: &cluster.Master,
		node.Worker: &cluster.Worker,
	}
	for t, c := range nodeTypeConfigs {
		if err := d.NodeManager.CreateNodes(t, c); err != nil {
			return err
		}
	}

	return nil
}

func (d *DefaultManager) Delete(name string, force bool) error {
	cluster, err := d.ConfigManager.GetCluster(name)
	if err != nil {
		return err
	}

	nodeTypeConfigs := map[node.Type]*pkgconfig.Node{
		node.Admin:  &cluster.Admin,
		node.Master: &cluster.Master,
		node.Worker: &cluster.Worker,
	}
	for t, c := range nodeTypeConfigs {
		if err := d.NodeManager.DeleteNodes(t, c); err != nil {
			if !force {
				return err
			}

			logrus.WithError(err).Warnln("failed to delete nodes")
		}
	}

	if err := d.ConfigManager.DeleteCluster(name); err != nil {
		return err
	}

	return nil
}

func (d *DefaultManager) Get(name string) (*data.Cluster, error) {
	configCluster, err := d.ConfigManager.GetCluster(name)
	if err != nil {
		return nil, err
	}

	nodes, err := d.NodeManager.ListNodes(name)
	if err != nil {
		return nil, err
	}

	return &data.Cluster{
		Name:  configCluster.Name,
		Spec:  *configCluster,
		Nodes: nodes,
	}, nil
}

func (d *DefaultManager) List() ([]*data.Cluster, error) {
	configClusters, err := d.ConfigManager.ListClusters()
	if err != nil {
		return nil, err
	}

	var clusters []*data.Cluster

	for _, c := range configClusters {
		// no need to have nodes info
		clusters = append(clusters, &data.Cluster{
			Name: c.Name,
			Spec: *c,
		})
	}

	return clusters, nil
}

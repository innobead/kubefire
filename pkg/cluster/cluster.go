package cluster

import (
	"fmt"
	"github.com/innobead/kubefire/pkg/bootstrap"
	pkgconfig "github.com/innobead/kubefire/pkg/config"
	"github.com/innobead/kubefire/pkg/data"
	"github.com/innobead/kubefire/pkg/node"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type Manager interface {
	Init(cluster *pkgconfig.Cluster) error
	Create(name string) error
	Delete(name string, force bool) error
	Get(name string) (*data.Cluster, error)
	List() ([]*data.Cluster, error)
	GetNodeManager() node.Manager
	GetConfigManager() pkgconfig.Manager
}

type DefaultManager struct {
	NodeManager   node.Manager
	Bootstrapper  bootstrap.Bootstrapper
	ConfigManager pkgconfig.Manager
}

func NewDefaultManager(nodeManager node.Manager, bootstrapper bootstrap.Bootstrapper, configManager pkgconfig.Manager) Manager {
	return &DefaultManager{
		NodeManager:   nodeManager,
		Bootstrapper:  bootstrapper,
		ConfigManager: configManager,
	}
}

func (d *DefaultManager) Init(cluster *pkgconfig.Cluster) error {
	logrus.WithField("cluster", cluster.Name).Infof("initializing cluster configuration")
	logrus.Debugf("%+v", cluster)

	if _, err := d.ConfigManager.GetCluster(cluster.Name); err == nil {
		return errors.New(fmt.Sprintf("cluster (%s) configuration already exists", cluster.Name))
	}

	return d.ConfigManager.SaveCluster(cluster.Name, cluster)
}

func (d *DefaultManager) Create(name string) error {
	logrus.Infof("creating cluster (%s)", name)

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
		if c.Count == 0 {
			continue
		}

		if err := d.NodeManager.CreateNodes(t, c); err != nil {
			return err
		}
	}

	return nil
}

func (d *DefaultManager) Delete(name string, force bool) error {
	logrus.WithFields(logrus.Fields{
		"cluster": name,
		"force":   force,
	}).Infoln("deleting cluster")

	cluster, err := d.ConfigManager.GetCluster(name)
	if err != nil {
		return err
	}

	nodeTypeConfigs := map[node.Type]*pkgconfig.Node{
		node.Admin:  &cluster.Admin,
		node.Master: &cluster.Master,
		node.Worker: &cluster.Worker,
	}
	for t, n := range nodeTypeConfigs {
		if err := d.NodeManager.DeleteNodes(t, n); err != nil {
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
	logrus.Debugf("getting cluster (%s)", name)

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
	logrus.Debugln("listing clusters")

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

func (d *DefaultManager) GetNodeManager() node.Manager {
	return d.NodeManager
}

func (d *DefaultManager) GetConfigManager() pkgconfig.Manager {
	return d.ConfigManager
}

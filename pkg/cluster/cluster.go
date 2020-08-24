package cluster

import (
	"fmt"
	pkgconfig "github.com/innobead/kubefire/pkg/config"
	"github.com/innobead/kubefire/pkg/data"
	"github.com/innobead/kubefire/pkg/node"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type Manager interface {
	Init(cluster *pkgconfig.Cluster) error
	Create(name string, started bool) error
	Delete(name string, force bool) error
	Get(name string) (*data.Cluster, error)
	List() ([]*data.Cluster, error)
	GetNodeManager() node.Manager
	GetConfigManager() pkgconfig.Manager
}

type DefaultManager struct {
	nodeManager   node.Manager
	configManager pkgconfig.Manager
}

func NewDefaultManager() Manager {
	return &DefaultManager{}
}

func (d *DefaultManager) SetNodeManager(nodeManager node.Manager) {
	d.nodeManager = nodeManager
}

func (d *DefaultManager) SetConfigManager(configManager pkgconfig.Manager) {
	d.configManager = configManager
}

func (d *DefaultManager) Init(cluster *pkgconfig.Cluster) error {
	logrus.WithField("cluster", cluster.Name).Infoln("initializing cluster configuration")
	logrus.Debugf("%+v", cluster)

	if _, err := d.configManager.GetCluster(cluster.Name); err == nil {
		return errors.New(fmt.Sprintf("cluster (%s) configuration already exists", cluster.Name))
	}

	return d.configManager.SaveCluster(cluster)
}

func (d *DefaultManager) Create(name string, started bool) error {
	logrus.WithField("cluster", name).Infoln("creating cluster")

	cluster, err := d.configManager.GetCluster(name)
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

		if err := d.nodeManager.CreateNodes(t, c, started); err != nil {
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

	cluster, err := d.configManager.GetCluster(name)
	if err != nil {
		return err
	}

	nodeTypeConfigs := map[node.Type]*pkgconfig.Node{
		node.Admin:  &cluster.Admin,
		node.Master: &cluster.Master,
		node.Worker: &cluster.Worker,
	}
	for t, n := range nodeTypeConfigs {
		if err := d.nodeManager.DeleteNodes(t, n); err != nil {
			if !force {
				return err
			}

			logrus.WithError(err).Warnln("failed to delete nodes")
		}
	}

	if err := d.configManager.DeleteCluster(cluster); err != nil {
		return err
	}

	return nil
}

func (d *DefaultManager) Get(name string) (*data.Cluster, error) {
	logrus.WithField("cluster", name).Debugln("getting cluster")

	configCluster, err := d.configManager.GetCluster(name)
	if err != nil {
		return nil, err
	}

	nodes, err := d.nodeManager.ListNodes(name)
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

	configClusters, err := d.configManager.ListClusters()
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
	return d.nodeManager
}

func (d *DefaultManager) GetConfigManager() pkgconfig.Manager {
	return d.configManager
}

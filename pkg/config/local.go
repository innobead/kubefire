package config

import (
	"github.com/goccy/go-yaml"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path"
)

const (
	ProjectDir = ".kubefire"
)

var (
	HomeDir, _     = os.UserHomeDir()
	ClusterRootDir = path.Join(HomeDir, ProjectDir)
)

func init() {
	_ = os.MkdirAll(ClusterRootDir, 0755)
}

func ClusterDir(name string) string {
	return path.Join(ClusterRootDir, name)
}

func ClusterConfigFile(name string) string {
	return path.Join(ClusterDir(name), "cluster.yaml")
}

type LocalConfigManager struct {
}

func NewLocalConfigManager() *LocalConfigManager {
	return &LocalConfigManager{}
}

func (l *LocalConfigManager) SaveCluster(name string, cluster *Cluster) error {
	logrus.Infof("Saving cluster (%s) configurations", name)

	d := ClusterDir(name)

	if _, err := os.Stat(d); os.IsNotExist(err) {
		if err := os.MkdirAll(ClusterDir(name), 0755); err != nil {
			return err
		}
	}

	bytes, err := yaml.Marshal(cluster)
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(ClusterConfigFile(name), bytes, 0755); err != nil {
		return err
	}

	return nil
}

func (l *LocalConfigManager) DeleteCluster(name string) error {
	logrus.Infof("Deleting cluster (%s) configurations", name)

	return os.RemoveAll(ClusterDir(name))
}

func (l *LocalConfigManager) GetCluster(name string) (*Cluster, error) {
	logrus.Debugf("Getting cluster (%s) configurations", name)

	bytes, err := ioutil.ReadFile(ClusterConfigFile(name))
	if err != nil {
		return nil, err
	}

	c := &Cluster{}
	if err := yaml.Unmarshal(bytes, c); err != nil {
		return nil, err
	}

	c.Admin.Cluster = c
	c.Master.Cluster = c
	c.Worker.Cluster = c

	return c, nil
}

func (l *LocalConfigManager) ListClusters() ([]*Cluster, error) {
	logrus.Debugln("Getting the list of cluster configurations")

	clusterDirs, err := ioutil.ReadDir(ClusterRootDir)
	if err != nil {
		return nil, err
	}

	var clusters []*Cluster

	for _, clusterDir := range clusterDirs {
		if !clusterDir.IsDir() {
			continue
		}

		c, err := l.GetCluster(clusterDir.Name())
		if err != nil {
			return nil, err
		}

		clusters = append(clusters, c)
	}

	return clusters, nil
}

package config

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"github.com/goccy/go-yaml"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
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

type LocalConfigManager struct {
}

func NewLocalConfigManager() *LocalConfigManager {
	return &LocalConfigManager{}
}

func (l *LocalConfigManager) SaveCluster(name string, cluster *Cluster) error {
	logrus.WithField("cluster", name).Infoln("Saving cluster configurations")

	d := LocalClusterDir(name)

	if _, err := os.Stat(d); os.IsNotExist(err) {
		if err := os.MkdirAll(LocalClusterDir(name), 0755); err != nil {
			return errors.WithStack(err)
		}
	}

	if err := l.generateKeys(cluster); err != nil {
		return errors.WithStack(err)
	}

	bytes, err := yaml.Marshal(cluster)
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(LocalClusterConfigFile(name), bytes, 0755); err != nil {
		return err
	}

	return nil
}

func (l *LocalConfigManager) DeleteCluster(name string) error {
	logrus.Infof("deleting cluster (%s) configurations", name)

	return os.RemoveAll(LocalClusterDir(name))
}

func (l *LocalConfigManager) GetCluster(name string) (*Cluster, error) {
	logrus.WithField("cluster", name).Debugln("Getting cluster configurations")

	bytes, err := ioutil.ReadFile(LocalClusterConfigFile(name))
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

func (l *LocalConfigManager) generateKeys(cluster *Cluster) error {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	cluster.Prikey, cluster.Pubkey = LocalClusterKeyFiles(cluster.Name)

	_ = os.Remove(cluster.Prikey)
	_ = os.Remove(cluster.Pubkey)

	keysInfo := []struct {
		keyType string
		path    string
		private bool
	}{
		{
			"PRIVATE KEY",
			cluster.Prikey,
			true,
		},
		{
			"PUBLIC KEY",
			cluster.Pubkey,
			false,
		},
	}

	var f *os.File
	defer func() {
		if f != nil {
			f.Close()
		}
	}()

	for _, keyInfo := range keysInfo {
		f, err = os.Create(keyInfo.path)
		if err != nil {
			return err
		}

		var encodeErr error

		switch keyInfo.private {
		case true:
			encodeErr = pem.Encode(f, &pem.Block{
				Type:  "RSA PRIVATE KEY",
				Bytes: x509.MarshalPKCS1PrivateKey(key),
			})

		default:
			pubkey, err := ssh.NewPublicKey(&key.PublicKey)
			if err != nil {
				return err
			}

			_, encodeErr = f.Write(ssh.MarshalAuthorizedKey(pubkey))
		}

		f.Close()

		if encodeErr != nil {
			_ = os.Remove(cluster.Prikey)
			_ = os.Remove(cluster.Pubkey)

			return encodeErr
		}
	}

	return nil
}

func LocalClusterDir(name string) string {
	return path.Join(ClusterRootDir, name)
}

func LocalClusterConfigFile(name string) string {
	return path.Join(LocalClusterDir(name), "cluster.yaml")
}

func LocalClusterKeyFiles(name string) (string, string) {
	return path.Join(LocalClusterDir(name), "key"), path.Join(LocalClusterDir(name), "key.pub")
}

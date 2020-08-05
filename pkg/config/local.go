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

var (
	HomeDir, _     = os.UserHomeDir()
	RootDir        = path.Join(HomeDir, ".kubefire")
	ClusterRootDir = path.Join(RootDir, "clusters")
	BinDir         = path.Join(RootDir, "bin")
)

type LocalConfigManager struct {
}

func init() {
	_ = os.MkdirAll(RootDir, 0755)
}

func NewLocalConfigManager() *LocalConfigManager {
	return &LocalConfigManager{}
}

func (l *LocalConfigManager) SaveCluster(cluster *Cluster) error {
	logrus.WithField("cluster", cluster.Name).Infoln("saving cluster configurations")

	if err := os.MkdirAll(cluster.LocalClusterDir(), 0755); err != nil && err != os.ErrExist {
		return errors.WithStack(err)
	}

	if err := l.generateKeys(cluster); err != nil {
		return errors.WithStack(err)
	}

	bytes, err := yaml.Marshal(cluster)
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(cluster.LocalClusterConfigFile(), bytes, 0755); err != nil {
		return err
	}

	return nil
}

func (l *LocalConfigManager) DeleteCluster(cluster *Cluster) error {
	logrus.WithField("cluster", cluster.Name).Infoln("deleting cluster configurations")

	return errors.WithStack(os.RemoveAll(cluster.LocalClusterDir()))
}

func (l *LocalConfigManager) GetCluster(name string) (*Cluster, error) {
	logrus.WithField("cluster", name).Debugln("getting cluster configurations")

	c := NewCluster()
	c.Name = name

	bytes, err := ioutil.ReadFile(c.LocalClusterConfigFile())
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if err := yaml.Unmarshal(bytes, c); err != nil {
		return nil, errors.WithStack(err)
	}

	c.Admin.Cluster = c
	c.Master.Cluster = c
	c.Worker.Cluster = c

	return c, nil
}

func (l *LocalConfigManager) ListClusters() ([]*Cluster, error) {
	logrus.Debugln("getting the list of cluster configurations")

	clusterDirs, err := ioutil.ReadDir(ClusterRootDir)
	if err != nil {
		return nil, errors.WithStack(err)
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
		return errors.WithStack(err)
	}

	cluster.Prikey, cluster.Pubkey = cluster.LocalClusterKeyFiles()

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
			return errors.WithStack(err)
		}

		if err := f.Chmod(0600); err != nil {
			return errors.WithStack(err)
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

			return errors.WithStack(encodeErr)
		}
	}

	return nil
}

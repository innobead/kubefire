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
	HomeDir, _          = os.UserHomeDir()
	RootDir             = path.Join(HomeDir, ".kubefire")
	ClusterRootDir      = path.Join(RootDir, "clusters")
	BinDir              = path.Join(RootDir, "bin")
	BootstrapperRootDir = path.Join(RootDir, "bootstrappers")
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

	if _, err := os.Stat(cluster.Pubkey); os.IsNotExist(err) {
		if err := l.generateKeys(cluster); err != nil {
			return errors.WithStack(err)
		}
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

	err := os.RemoveAll(cluster.LocalClusterDir())
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (l *LocalConfigManager) GetCluster(name string) (*Cluster, error) {
	logrus.WithField("cluster", name).Debugln("getting cluster configurations")

	c := NewCluster()
	c.Name = name

	bytes, err := ioutil.ReadFile(c.LocalClusterConfigFile())
	if err != nil {
		return c, errors.WithStack(err)
	}

	if err := yaml.Unmarshal(bytes, c); err != nil {
		return nil, errors.WithStack(err)
	}

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
			if os.IsNotExist(errors.Cause(err)) {
				logrus.WithField("cluster", clusterDir.Name()).Debugln("no cluster configurations found")
				continue
			}

			return nil, err
		}

		clusters = append(clusters, c)
	}

	return clusters, nil
}

func (l *LocalConfigManager) SaveBootstrapperVersions(latestVersion BootstrapperVersioner, versions []BootstrapperVersioner) error {
	logrus.WithField("bootstrapper", latestVersion.Type()).Debugln("saving bootstrapper version configurations")

	bytes, err := yaml.Marshal(versions)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(path.Dir(latestVersion.LocalVersionFile()), 0755); err != nil && err != os.ErrExist {
		return errors.WithStack(err)
	}

	return ioutil.WriteFile(latestVersion.LocalVersionFile(), bytes, 0755)
}

func (l *LocalConfigManager) GetBootstrapperVersions(latestVersion BootstrapperVersioner) ([]BootstrapperVersioner, error) {
	logrus.WithField("bootstrapper", latestVersion.Type()).Debugln("getting bootstrapper version configurations")

	bytes, err := ioutil.ReadFile(latestVersion.LocalVersionFile())
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var bootstrapperVersions []BootstrapperVersioner

	switch latestVersion.(type) {
	case *KubeadmBootstrapperVersion:
		var versions []KubeadmBootstrapperVersion
		if err := yaml.Unmarshal(bytes, &versions); err != nil {
			return nil, errors.WithStack(err)
		}

		for _, v := range versions {
			v := v
			bootstrapperVersions = append(bootstrapperVersions, &v)
		}

	case *K3sBootstrapperVersion:
		var versions []K3sBootstrapperVersion
		if err := yaml.Unmarshal(bytes, &versions); err != nil {
			return nil, errors.WithStack(err)
		}

		for _, v := range versions {
			v := v
			bootstrapperVersions = append(bootstrapperVersions, &v)
		}

	case *SkubaBootstrapperVersion:
		var versions []SkubaBootstrapperVersion
		if err := yaml.Unmarshal(bytes, &versions); err != nil {
			return nil, errors.WithStack(err)
		}

		for _, v := range versions {
			v := v
			bootstrapperVersions = append(bootstrapperVersions, &v)
		}
	}

	return bootstrapperVersions, nil
}

func (l *LocalConfigManager) DeleteBootstrapperVersions(latestVersion BootstrapperVersioner) error {
	logrus.WithField("bootstrapper", latestVersion.Type()).Infoln("deleting bootstrapper version configurations")

	err := os.RemoveAll(path.Dir(latestVersion.LocalVersionFile()))
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
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

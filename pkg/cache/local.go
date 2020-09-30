package cache

import (
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path/filepath"
)

const (
	ClusterCacheType      Type = "cluster"
	BootstrapperCacheType Type = "bootstrapper"
	BinCacheType          Type = "bin"
)

type LocalManager struct {
	rootDir string
}

func NewLocalManager(rootDir string) Manager {
	return &LocalManager{
		rootDir: rootDir,
	}
}

func (l *LocalManager) typeDir(t Type, create bool) string {
	dir := string(t)

	switch t {
	case ClusterCacheType:
		dir = "clusters"
	case BootstrapperCacheType:
		dir = "bootstrappers"
	}

	p := filepath.Join(l.rootDir, dir)

	if create {
		err := os.MkdirAll(p, 0755)
		if err != nil {
			panic(err)
		}
	}

	return p
}

func (l *LocalManager) pathFile(t Type, path Path, create bool) string {
	p := filepath.Join(l.typeDir(t, create), string(path))

	if create {
		err := os.MkdirAll(filepath.Dir(p), 0755)
		if err != nil {
			panic(err)
		}
	}

	return p
}

func (l *LocalManager) Create(t Type, path Path, value Value) error {
	file := l.pathFile(t, path, true)
	logrus.Infof("Creating local cache %s\n", file)

	err := ioutil.WriteFile(file, value, 0755)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (l *LocalManager) Update(t Type, path Path, value Value) error {
	return l.Create(t, path, value)
}

func (l *LocalManager) Get(t Type, path Path, withValue bool) (*Cache, error) {
	file := string(path)
	if !filepath.IsAbs(file) {
		file = l.pathFile(t, path, true)
	}

	if _, err := os.Stat(file); os.IsNotExist(err) {
		return nil, errors.WithStack(err)
	}

	cache := &Cache{
		Type:        t,
		Path:        path,
		Description: "",
	}

	if withValue {
		bytes, err := ioutil.ReadFile(file)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		cache.Value = bytes
	}

	return cache, nil
}

func (l *LocalManager) List(t Type, withValue bool) ([]*Cache, error) {
	dir := l.typeDir(t, false)

	var cachePaths []*Cache

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}

			return err
		}

		if info.IsDir() {
			return nil
		}

		cache, err := l.Get(t, Path(path), withValue)
		if err != nil {
			return err
		}

		cachePaths = append(
			cachePaths,
			cache,
		)

		return nil
	})
	if err != nil {
		return nil, err
	}

	return cachePaths, nil
}

func (l *LocalManager) ListAll(withValue bool) ([]*Cache, error) {
	var cachePaths []*Cache

	for _, t := range []Type{BootstrapperCacheType, BinCacheType} {
		caches, err := l.List(t, withValue)
		if err != nil {
			return nil, err
		}

		cachePaths = append(cachePaths, caches...)
	}

	return cachePaths, nil
}

func (l *LocalManager) Delete(t Type) error {
	dir := l.typeDir(t, false)
	logrus.Infof("Delete local cache dir %s\n", dir)

	return os.RemoveAll(dir)
}

func (l *LocalManager) DeleteAll() error {
	for _, t := range []Type{BootstrapperCacheType, BinCacheType} {
		err := l.Delete(t)
		if err != nil {
			return err
		}
	}

	return nil
}

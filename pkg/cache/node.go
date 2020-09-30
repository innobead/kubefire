package cache

import (
	"github.com/innobead/kubefire/pkg/node"
	"github.com/sirupsen/logrus"
)

const (
	NodeCacheType Type = "node"
)

type NodeCache struct {
	nodeManager node.Manager
}

func NewNodeCache(nodeManager node.Manager) Manager {
	return &NodeCache{
		nodeManager: nodeManager,
	}
}

func (n *NodeCache) Create(t Type, path Path, value Value) error {
	panic("implement me")
}

func (n *NodeCache) Update(t Type, path Path, value Value) error {
	panic("implement me")
}

func (n *NodeCache) Get(t Type, path Path, withValue bool) (*Cache, error) {
	panic("implement me")
}

func (n *NodeCache) List(t Type, withValue bool) ([]*Cache, error) {
	panic("implement me")
}

func (n *NodeCache) ListAll(withValue bool) ([]*Cache, error) {
	nodeCaches, err := n.nodeManager.GetCaches()
	if err != nil {
		return nil, err
	}

	var caches []*Cache

	for _, nc := range nodeCaches {
		cache := &Cache{}

		switch nc := nc.(type) {
		case *node.IgniteCache:
			cache.Type = NodeCacheType
			cache.Path = Path(nc.Name)
			cache.Description = nc.Description

		default:
			continue
		}

		caches = append(caches, cache)
	}

	return caches, nil
}

func (n *NodeCache) Delete(t Type) error {
	panic("implement me")
}

func (n *NodeCache) DeleteAll() error {
	logrus.Infof("Delete node caches\n")

	return n.nodeManager.DeleteCaches()
}

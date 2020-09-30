package cache

import (
	pkgconfig "github.com/innobead/kubefire/pkg/config"
	"github.com/innobead/kubefire/pkg/node"
)

type Type string
type Path string
type Value []byte

type Cache struct {
	Type        Type
	Path        Path
	Value       Value
	Description string
}

type Manager interface {
	Create(t Type, path Path, value Value) error
	Update(t Type, path Path, value Value) error
	Get(t Type, path Path, withValue bool) (*Cache, error)
	List(t Type, withValue bool) ([]*Cache, error)
	ListAll(withValue bool) ([]*Cache, error)
	Delete(t Type) error
	DeleteAll() error
}

func DefaultManagers(nodeManager node.Manager) []Manager {
	return []Manager{
		NewLocalManager(pkgconfig.RootDir),
		NewNodeCache(nodeManager),
	}
}

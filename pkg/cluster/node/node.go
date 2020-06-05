package node

type Node struct {
}

type Error error

type Manager interface {
	Create() (*Node, Error)
	Delete(*Node) Error
	Get(name string) (*Node, Error)
	List() ([]*Node, Error)
}

package output

import (
	"github.com/pkg/errors"
	"io"
)

type Type int

const (
	DEFAULT Type = iota
	JSON
	YAML
)

type Outputer interface {
	Print(obj interface{}, filters []string, title string) error
}

func NewOutput(t Type, writer io.Writer) (Outputer, error) {
	d := DefaultOutput{writer}

	switch t {
	case JSON:
		return &JsonOutput{d}, nil

	case YAML:
		return &YamlOutput{d}, nil

	case DEFAULT:
		return &DefaultOutput{&d}, nil

	default:
		return nil, errors.New("")
	}
}

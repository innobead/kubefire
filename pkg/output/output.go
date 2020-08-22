package output

import (
	interr "github.com/innobead/kubefire/internal/error"
	"io"
)

type Type string

const (
	DEFAULT Type = "default"
	JSON    Type = "json"
	YAML    Type = "yaml"
)

var BuiltinTypes = []Type{DEFAULT, JSON, YAML}

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
		return nil, interr.NotFoundError
	}
}

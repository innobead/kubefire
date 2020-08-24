package output

import (
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

func NewOutput(t Type, writer io.Writer) Outputer {
	d := DefaultOutput{writer}

	switch t {
	case JSON:
		return &JsonOutput{d}

	case YAML:
		return &YamlOutput{d}

	case DEFAULT:
		return &DefaultOutput{&d}

	default:
		panic("no supported output")
	}
}

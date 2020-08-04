package output

import (
	"fmt"
	"github.com/goccy/go-yaml"
	"github.com/pkg/errors"
)

type YamlOutput struct {
	DefaultOutput
}

func (j *YamlOutput) Print(obj interface{}, filters []string, title string) error {
	bytes, err := yaml.Marshal(obj)
	if err != nil {
		return errors.WithStack(err)
	}

	fmt.Print(string(bytes))

	return nil
}

package output

import (
	"fmt"
	"github.com/go-yaml/yaml"
)

type YamlOutput struct {
	DefaultOutput
}

func (j *YamlOutput) Print(obj interface{}) error {
	bytes, err := yaml.Marshal(obj)
	if err != nil {
		return err
	}

	fmt.Println(bytes)

	return nil
}

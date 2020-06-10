package output

import (
	"fmt"
	"github.com/goccy/go-yaml"
)

type YamlOutput struct {
	DefaultOutput
}

func (j *YamlOutput) Print(obj interface{}, filters []string, title string) error {
	bytes, err := yaml.Marshal(obj)
	if err != nil {
		return err
	}

	fmt.Println(string(bytes))

	return nil
}

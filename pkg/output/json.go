package output

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
)

type JsonOutput struct {
	DefaultOutput
}

func (j *JsonOutput) Print(obj interface{}, filters []string, title string) error {
	bytes, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return errors.WithStack(err)
	}

	fmt.Println(string(bytes))

	return nil
}

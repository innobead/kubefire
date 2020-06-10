package output

import (
	"encoding/json"
	"fmt"
)

type JsonOutput struct {
	DefaultOutput
}

func (j *JsonOutput) Print(obj interface{}, filters []string, title string) error {
	bytes, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(bytes))

	return nil
}

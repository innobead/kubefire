package output

import (
	"encoding/json"
	"fmt"
)

type JsonOutput struct {
	DefaultOutput
}

func (j *JsonOutput) Print(obj interface{}) error {
	bytes, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	fmt.Println(bytes)

	return nil
}

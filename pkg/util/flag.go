package util

import (
	"fmt"
	"github.com/innobead/kubefire/pkg/output"
	"reflect"
	"strings"
)

func FlagsValuesUsage(prefix string, candidates interface{}) string {
	value := reflect.ValueOf(candidates)

	if value.Kind() != reflect.Slice {
		return "unknown"
	}

	var strs []string

	for i := 0; i < value.Len(); i++ {
		switch v := value.Index(i).Interface().(type) {
		case fmt.Stringer:
			strs = append(strs, v.String())
		case output.Type:
			strs = append(strs, string(v))
		case string:
			strs = append(strs, v)
		}
	}

	return fmt.Sprintf("%s, options: [%s]", prefix, strings.Join(strs, ", "))
}

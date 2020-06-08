package output

import (
	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
	"io"
	"os"
	"reflect"
)

type Type int

const (
	DEFAULT Type = iota
	JSON
	YAML
)

type Outputer interface {
	Print(obj interface{}) error
}

type DefaultOutput struct {
	io.Writer
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

func (d *DefaultOutput) Print(obj interface{}) error {
	value := reflect.ValueOf(obj)

	var headers []string
	var data [][]string

	if value.Kind() == reflect.Slice {
		for i := 0; i < value.Len(); i++ {
			if value.Kind() == reflect.Ptr {
				d.parse(reflect.Indirect(value).Index(i), &headers, &data)
			} else {
				d.parse(value.Index(i), &headers, &data)
			}
		}
	} else {
		if value.Kind() == reflect.Ptr {
			d.parse(reflect.Indirect(value), &headers, &data)
		} else {
			d.parse(value, &headers, &data)
		}
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(headers)
	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(true)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeaderLine(false)
	table.SetBorder(false)
	table.SetTablePadding("\t") // pad with tabs
	table.SetNoWhiteSpace(true)
	table.AppendBulk(data) // Add Bulk Data
	table.Render()

	return nil
}

func (d *DefaultOutput) parse(v reflect.Value, headers *[]string, data *[][]string) {
	var subdata []string

	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		fs := v.Type().Field(i)

		switch f.Kind() {
		case reflect.String:
			subdata = append(subdata, f.String())
		case reflect.Int:
			subdata = append(subdata, string(f.Int()))
		default:
			continue
		}

		if len(*headers) <= i {
			*headers = append(*headers, fs.Name)
		}
	}

	*data = append(*data, subdata)
}

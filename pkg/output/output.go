package output

import (
	"fmt"
	"github.com/innobead/kubefire/pkg/data"
	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
	"io"
	"os"
	"reflect"
	"strconv"
	"strings"
)

type Type int

const (
	DEFAULT Type = iota
	JSON
	YAML
)

type Outputer interface {
	Print(obj interface{}, filters []string, title string) error
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

func (d *DefaultOutput) Print(obj interface{}, filters []string, title string) error {
	value := reflect.ValueOf(obj)

	type subObjectType struct {
		title string
		obj   interface{}
	}
	var subObjs []subObjectType

	var tableHeaders []string
	var tableData [][]string

	if value.Kind() == reflect.Slice {
		fmt.Printf("= %s\n", title)

		for i := 0; i < value.Len(); i++ {
			if value.Index(i).Kind() == reflect.Ptr {
				d.parse(reflect.Indirect(value.Index(i)), filters, &tableHeaders, &tableData)
			} else {
				d.parse(value.Index(i), filters, &tableHeaders, &tableData)
			}
		}
	} else {
		if value.Kind() == reflect.Ptr {
			if err := d.Print(reflect.Indirect(value).Interface(), filters, title); err != nil {
				return err
			}
		} else {
			switch value.Interface().(type) {
			case data.Cluster:
				fmt.Println("= Cluster Configuration")

				clusterSpecField := value.FieldByName("Spec")

				d.parse(clusterSpecField, filters, &tableHeaders, &tableData)
				subObjs = append(
					subObjs,
					subObjectType{title: "Admin Node Configuration", obj: clusterSpecField.FieldByName("Admin").Interface()},
					subObjectType{title: "Master Node Configuration", obj: clusterSpecField.FieldByName("Master").Interface()},
					subObjectType{title: "Worker Node Configuration", obj: clusterSpecField.FieldByName("Worker").Interface()},
				)

				if value.FieldByName("Nodes").IsValid() {
					subObjs = append(
						subObjs,
						subObjectType{title: "Node Runtime", obj: value.FieldByName("Nodes").Interface()},
					)
				}

			default:
				fmt.Printf("= %s\n", title)

				d.parse(value, filters, &tableHeaders, &tableData)
			}
		}
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(tableHeaders)
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
	table.AppendBulk(tableData) // Add Bulk Data
	table.Render()

	fmt.Println("")

	for _, o := range subObjs {
		if err := d.Print(o.obj, nil, o.title); err != nil {
			return err
		}
	}

	return nil
}

func (d *DefaultOutput) parse(v reflect.Value, filters []string, tableHeaders *[]string, tableData *[][]string) {
	switch v.Interface().(type) {
	case data.Node:
		filters = append(filters, "Name", "Status.Running")
	}

	var subdata []string
	filterCount := len(filters)

loop:
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		fs := v.Type().Field(i)

		found := false

		if filters != nil {
			newFilters := filters

		filterLoop:
			for _, filter := range filters {
				subFilters := strings.Split(filter, ".")

				for i, subFilter := range subFilters {
					if i > 0 {
						fs, _ = f.Type().FieldByName(subFilter)
						f = f.FieldByName(subFilter)
					}

					if i == len(subFilters)-1 {
						if strings.EqualFold(subFilter, fs.Name) {
							found = true
							break filterLoop
						}
					}

					newFilters = append(newFilters, filter)
				}
			}

			if !found {
				continue loop
			} else {
				filters = newFilters
			}
		}

		switch f.Kind() {
		case reflect.String:
			subdata = append(subdata, f.String())
		case reflect.Int:
			subdata = append(subdata, strconv.FormatInt(f.Int(), 10))
		case reflect.Bool:
			subdata = append(subdata, strconv.FormatBool(f.Bool()))
		default:
			continue
		}

		if filters != nil && len(*tableHeaders) < filterCount || filterCount == 0 && len(*tableHeaders) < v.NumField() {
			*tableHeaders = append(*tableHeaders, fs.Name)
		}
	}

	*tableData = append(*tableData, subdata)
}

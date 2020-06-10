package output

import (
	"fmt"
	"github.com/innobead/kubefire/pkg/data"
	"github.com/olekukonko/tablewriter"
	"io"
	"os"
	"reflect"
	"strconv"
	"strings"
)

type DefaultOutput struct {
	io.Writer
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

				specField := value.FieldByName("Spec")

				d.parse(specField, filters, &tableHeaders, &tableData)
				subObjs = append(
					subObjs,
					subObjectType{title: "Admin Node Configuration", obj: specField.FieldByName("Admin").Interface()},
					subObjectType{title: "Master Node Configuration", obj: specField.FieldByName("Master").Interface()},
					subObjectType{title: "Worker Node Configuration", obj: specField.FieldByName("Worker").Interface()},
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

	var subTableData []string
	filterCount := len(filters)

loop:
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		fs := v.Type().Field(i)

		if filters != nil {
			newFilters := filters
			filterMatched := false

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
							filterMatched = true
							break filterLoop
						}
					}

					newFilters = append(newFilters, filter)
				}
			}

			if !filterMatched {
				continue loop
			} else {
				filters = newFilters
			}
		}

		switch f.Kind() {
		case reflect.String:
			subTableData = append(subTableData, f.String())

		case reflect.Int:
			subTableData = append(subTableData, strconv.FormatInt(f.Int(), 10))

		case reflect.Bool:
			subTableData = append(subTableData, strconv.FormatBool(f.Bool()))

		default:
			continue
		}

		if filters != nil && len(*tableHeaders) < filterCount || filterCount == 0 && len(*tableHeaders) < v.NumField() {
			*tableHeaders = append(*tableHeaders, fs.Name)
		}
	}

	*tableData = append(*tableData, subTableData)
}

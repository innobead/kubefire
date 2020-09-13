package output

import (
	"fmt"
	"github.com/innobead/kubefire/pkg/config"
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
		if value.Len() == 0 {
			fmt.Println("No clusters created")
			return nil
		}

		if title != "" {
			fmt.Printf("### %s\n", title)
		}

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
				fmt.Println("### Cluster Configuration")

				specField := value.FieldByName("Spec")

				d.parse(specField, filters, &tableHeaders, &tableData)
				subObjs = append(
					subObjs,
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
				if title != "" {
					fmt.Printf("### %s\n", title)
				}

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

	for _, o := range subObjs {
		fmt.Println("")

		if err := d.Print(o.obj, nil, o.title); err != nil {
			return err
		}
	}

	return nil
}

func (d *DefaultOutput) parse(v reflect.Value, filters []string, tableHeaders *[]string, tableData *[][]string) {
	switch v.Interface().(type) {
	case data.Node:
		filters = append(
			filters,
			"Name",
			"Status.Running",
			"Status.IPAddresses",
			"Status.Image",
			"Status.Kernel",
		)

	case config.Cluster:
		if len(filters) == 0 {
			filters = append(
				filters,
				"Name",
				"Bootstrapper",
				"Image",
				"KernelImage",
				"KernelArgs",
				"ExtraOptions",
			)
		}
	}

	updateTableData := func(f reflect.Value, subTableData *[]string) {
		switch f.Kind() {
		case reflect.Struct:
			v := f
			if f.CanAddr() {
				v = f.Addr()
			}

			if v, ok := v.Interface().(fmt.Stringer); ok {
				*subTableData = append(*subTableData, v.String())
			}

		case reflect.String:
			*subTableData = append(*subTableData, f.String())

		case reflect.Int:
			*subTableData = append(*subTableData, strconv.FormatInt(f.Int(), 10))

		case reflect.Bool:
			*subTableData = append(*subTableData, strconv.FormatBool(f.Bool()))
		}
	}

	var subTableData []string

	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		fs := v.Type().Field(i)

		if filters != nil {
			for _, filter := range filters {
				newField := f
				newFieldStruct := fs

				subFilters := strings.Split(filter, ".")

				for i, subFilter := range subFilters {
					if i == 0 && newFieldStruct.Name != subFilter {
						break
					}

					if i > 0 {
						newFieldStruct, _ = newField.Type().FieldByName(subFilter)
						newField = newField.FieldByName(subFilter)
					}

					if i == len(subFilters)-1 {
						if strings.EqualFold(subFilter, newFieldStruct.Name) {
							updateTableData(newField, &subTableData)

							if len(*tableHeaders) < len(filters) {
								*tableHeaders = append(*tableHeaders, newFieldStruct.Name)
							}

							break
						}
					}
				}
			}

			continue
		}

		updateTableData(f, &subTableData)

		if len(*tableHeaders) < v.NumField() {
			*tableHeaders = append(*tableHeaders, fs.Name)
		}
	}

	*tableData = append(*tableData, subTableData)
}

package static

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func Merge(format string, data any) (string, error) {
	t, err := template.New("fstring").Funcs(
		template.FuncMap{
			"DerefBool":   func(i *bool) bool { return *i },
			"DerefString": func(i *string) string { return *i },
			"DerefInt64":  func(i *int64) int64 { return *i },
			"StringValue": func(i types.String) string { return i.ValueString() },
			"Int64Value":  func(i types.Int64) int64 { return i.ValueInt64() },
		},
	).Parse(format)
	if err != nil {
		return "", fmt.Errorf("Merge: error parsing template: %w", err)
	}
	output := new(bytes.Buffer)

	if err := t.Execute(output, data); err != nil {
		return "", fmt.Errorf("Merge: error executing template: %w", err)
	}

	return output.String(), nil
}

func Pointer[T any](d T) *T {
	return &d
}

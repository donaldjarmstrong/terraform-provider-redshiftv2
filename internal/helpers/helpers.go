package helpers

import (
	"bytes"
	"context"
	"fmt"
	"slices"
	"strings"
	"text/template"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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
			"StringsJoin": strings.Join,
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

// returns the items from the left that are not in the right.
func MissingFrom(left []string, right []string) []string {
	var diff []string

	for _, v := range left {
		if !slices.Contains(right, v) {
			diff = append(diff, v)
		}
	}

	return diff
}

// Set the set value to the elements or return null.
func SetValueOrNull[T any](ctx context.Context, elementType attr.Type, elements []T, diags *diag.Diagnostics) types.Set {
	if len(elements) == 0 {
		return types.SetNull(elementType)
	}

	result, d := types.SetValueFrom(ctx, elementType, elements)
	diags.Append(d...)
	return result
}

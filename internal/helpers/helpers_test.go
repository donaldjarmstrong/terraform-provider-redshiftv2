package helpers

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func Test_Merge_func_map(t *testing.T) {
	yes := true
	name := "test"
	age := int64(44)
	user := "username"
	size := int64(11)
	names := []string{"a", "b", "c"}

	tests := []struct {
		Yesp  *bool
		Namep *string
		Agep  *int64
		User  types.String
		Size  types.Int64
		Names *[]string
	}{
		{
			Yesp:  &yes,
			Namep: &name,
			Agep:  &age,
			User:  types.StringValue(user),
			Size:  types.Int64Value(size),
			Names: &names,
		},
	}

	for _, tt := range tests {
		tplt := `yes={{(DerefBool .Yesp)}} name={{(DerefString .Namep)}} age={{(DerefInt64 .Agep)}} user={{(StringValue .User)}} size={{(Int64Value .Size)}} names={{(StringsJoin .Names ", ")}}`
		expected := fmt.Sprintf("yes=%t name=%s age=%d user=%s size=%d names=%s", yes, name, age, user, size, strings.Join(names, ", "))

		actual, err := Merge(tplt, tt)
		assert.Nil(t, err)
		assert.Equal(t, expected, actual)
	}
}

func Test_Pointer(t *testing.T) {
	yes := true
	f := Pointer[bool](yes)
	assert.Equal(t, &yes, f)

	name := "name"
	g := Pointer[string](name)
	assert.Equal(t, &name, g)
}

func Test_MissingFrom(t *testing.T) {
	slice1 := []string{"foo", "bar", "hello"}
	slice2 := []string{"foo", "world", "bar", "foo"}

	assert.Equal(t, []string{"hello"}, MissingFrom(slice1, slice2))
	assert.Equal(t, []string{"world"}, MissingFrom(slice2, slice1))
}

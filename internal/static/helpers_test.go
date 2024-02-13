package static

import (
	"fmt"
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

	tests := []struct {
		Yesp  *bool
		Namep *string
		Agep  *int64
		User  types.String
		Size  types.Int64
	}{
		{
			Yesp:  &yes,
			Namep: &name,
			Agep:  &age,
			User:  types.StringValue(user),
			Size:  types.Int64Value(size),
		},
	}

	for _, tt := range tests {
		tplt := "yes={{(DerefBool .Yesp)}} name={{(DerefString .Namep)}} age={{(DerefInt64 .Agep)}} user={{(StringValue .User)}} size={{(Int64Value .Size)}}"
		expected := fmt.Sprintf("yes=%t name=%s age=%d user=%s size=%d", yes, name, age, user, size)

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

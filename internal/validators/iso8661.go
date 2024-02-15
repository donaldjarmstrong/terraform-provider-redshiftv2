package validators

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ validator.String = iso86601Validator{}

type iso86601Validator struct {
	layout string
}

// Description describes the validation in plain text formatting.
func (v iso86601Validator) Description(_ context.Context) string {
	return fmt.Sprintf("value must be a valid ISO8861 date time using layout '%s'", v.layout)
}

// MarkdownDescription describes the validation in Markdown formatting.
func (v iso86601Validator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// Validate performs the validation.
func (v iso86601Validator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		// Only validate if value is known
		return
	}

	_, err := time.Parse(time.DateTime, req.ConfigValue.ValueString())
	if err != nil {
		resp.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
			req.Path,
			v.Description(ctx),
			req.ConfigValue.ValueString(),
		))
	}
}

// iso86601Validator returns a validator which ensures that any configured
// attribute value can be parsed into a time.Time from a layout of time.DateTime.
func Iso8601Validator() validator.String {
	return iso86601Validator{
		layout: time.DateTime,
	}
}

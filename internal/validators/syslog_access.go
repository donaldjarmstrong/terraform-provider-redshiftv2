package validators

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ validator.String = syslogAccessValidator{}

type syslogAccessValidator struct {
	syslog_access string
	createdb      bool
}

// Description describes the validation in plain text formatting.
func (v syslogAccessValidator) Description(_ context.Context) string {
	return fmt.Sprintf("syslog_access must be set to %s when createdb is true", v.syslog_access)
}

// MarkdownDescription describes the validation in Markdown formatting.
func (v syslogAccessValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// Validate performs the validation.
func (v syslogAccessValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	createuserPath := req.Path.ParentPath().AtName("createuser")

	var createuser types.Bool
	diags := req.Config.GetAttribute(ctx, createuserPath, &createuser)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	// syslog_access is always known, it is a required attribute. no need to check

	if createuser.IsNull() || createuser.IsUnknown() {
		// Only validate if createuser value is known
		return
	}

	// only validate if the value is true
	if createuser.ValueBool() {
		if req.ConfigValue.ValueString() != v.syslog_access {
			resp.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
				req.Path,
				v.Description(ctx),
				fmt.Sprintf("%s", req.ConfigValue.ValueString()),
			))
		}
	}
}

// syslogAccessValidator returns an AttributeValidator which ensures that for syslog_access
// attribute value:
//
//   - When Createdb is true, syslog_access must be UNRESTRICTED
func SyslogAccessValidator() validator.String {
	return syslogAccessValidator{
		syslog_access: "UNRESTRICTED",
		createdb:      true,
	}
}

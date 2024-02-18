// Code generated by terraform-plugin-framework-generator DO NOT EDIT.

package generated

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-redshift/internal/helpers"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func GroupResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				Description:         "Built-in identifier",
				MarkdownDescription: "Built-in identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				Description:         "Name of the new user group. Group names beginning with two underscores are reserved for Amazon Redshift internal use. For more information about valid names, see Names and identifiers.",
				MarkdownDescription: "Name of the new user group. Group names beginning with two underscores are reserved for Amazon Redshift internal use. For more information about valid names, see Names and identifiers.",
				Validators: []validator.String{
					stringvalidator.NoneOfCaseInsensitive(`public`),
					stringvalidator.UTF8LengthBetween(1, 127),
					stringvalidator.NoneOfCaseInsensitive(helpers.ReservedWords...),
				},
			},
			"usernames": schema.SetAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "Name(s) of the user to add to the group.",
				MarkdownDescription: "Name(s) of the user to add to the group.",
				Validators: []validator.Set{
					setvalidator.IsRequired(),
				},
			},
		},
	}
}

type GroupModel struct {
	Id        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	Usernames types.Set    `tfsdk:"usernames"`
}
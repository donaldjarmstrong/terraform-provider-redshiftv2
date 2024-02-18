package resources

import (
	"context"
	"fmt"
	"terraform-provider-redshift/internal/generated"
	"terraform-provider-redshift/internal/helpers"
	"terraform-provider-redshift/internal/redshift"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/tracelog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                   = &roleResource{}
	_ resource.ResourceWithConfigure      = &roleResource{}
	_ resource.ResourceWithValidateConfig = &roleResource{}
	_ resource.ResourceWithImportState    = &roleResource{}
)

func NewRoleResource() resource.Resource {
	return &roleResource{}
}

type roleResource struct {
	ConnCfg *pgx.ConnConfig
}

func (r *roleResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_role"
}

func (r *roleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = generated.RoleResourceSchema(ctx)
}

func (r *roleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan generated.RoleModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.ConnCfg.Tracer = &tracelog.TraceLog{
		Logger:   helpers.NewLogger(ctx, "role_resource.Create"),
		LogLevel: tracelog.LogLevelTrace,
	}

	createDDL := redshift.CreateRoleDDLParams{
		Name:       plan.Name.ValueString(),
		ExternalId: plan.ExternalId.ValueStringPointer(),
	}

	svc, err := redshift.NewRoleService(ctx, r.ConnCfg)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to execute NewRoleService",
			"An unexpected error occurred when calling NewRoleService. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"Unable to Create: "+err.Error(),
		)
		return
	}

	role, err := svc.CreateRole(createDDL)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to execute CreateRole on service RoleService",
			"An unexpected error occurred when calling CreateRole on service RoleService. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"Unable to Create: "+err.Error(),
		)
		return
	}

	// Only set those undefaulted computed
	plan.Id = types.StringValue(role.Id)
	plan.ExternalId = types.StringPointerValue(role.ExternalId)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *roleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state generated.RoleModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.ConnCfg.Tracer = &tracelog.TraceLog{
		Logger:   helpers.NewLogger(ctx, "role_resource.Read"),
		LogLevel: tracelog.LogLevelTrace,
	}

	svc, err := redshift.NewRoleService(ctx, r.ConnCfg)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to execute NewRoleService",
			"An unexpected error occurred when calling NewRoleService. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"Unable to Read: "+err.Error(),
		)
		return
	}

	role, err := svc.FindRole(state.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to execute FindRole on service RoleService",
			"An unexpected error occurred when calling FindRole on service RoleService. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"Unable to Read: "+err.Error(),
		)
		return
	}

	state.ExternalId = types.StringPointerValue(role.ExternalId)
	state.Name = types.StringValue(role.RoleName)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *roleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state generated.RoleModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.ConnCfg.Tracer = &tracelog.TraceLog{
		Logger:   helpers.NewLogger(ctx, "role_resource.Update"),
		LogLevel: tracelog.LogLevelTrace,
	}

	ddl := redshift.AlterRoleDDLParams{
		Name: state.Name.ValueString(),
	}

	if !plan.Name.Equal(state.Name) {
		newName := plan.Name.ValueString()
		ddl.RenameTo = &newName
	}

	if !plan.ExternalId.Equal(state.ExternalId) {
		ddl.ExternalId = plan.ExternalId.ValueStringPointer()
	}

	svc, err := redshift.NewRoleService(ctx, r.ConnCfg)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to execute NewRoleService",
			"An unexpected error occurred when calling NewRoleService. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"Unable to Update: "+err.Error(),
		)
		return
	}

	err = svc.AlterRole(ddl)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to execute AlterRole on service RoleService",
			"An unexpected error occurred when calling AlterRole on service RoleService. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"Unable to Update: "+err.Error(),
		)
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *roleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state generated.RoleModel

	// Read Terraform prior state into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	r.ConnCfg.Tracer = &tracelog.TraceLog{
		Logger:   helpers.NewLogger(ctx, "role_resource.Delete"),
		LogLevel: tracelog.LogLevelTrace,
	}

	svc, err := redshift.NewRoleService(ctx, r.ConnCfg)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to execute NewRoleService",
			"An unexpected error occurred when calling NewRoleService. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"Unable to Delete: "+err.Error(),
		)
		return
	}

	err = svc.DropRole(state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to execute DropRole on service RoleService",
			"An unexpected error occurred when calling DropRole on service RoleService. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"Unable to Delete: "+err.Error(),
		)
		return
	}

	// If the logic reaches here, it implicitly succeeded and will remove
	// the resource from state if there are no other errors
}

func (r *roleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *roleResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	cfg, ok := req.ProviderData.(*pgx.ConnConfig)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Type",
			fmt.Sprintf("Expected *pgx.ConnConfig, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.ConnCfg = cfg
}

func (r *roleResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var plan generated.RoleModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}
}

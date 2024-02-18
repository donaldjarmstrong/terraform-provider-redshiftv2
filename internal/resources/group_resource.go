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
	_ resource.Resource                   = &groupResource{}
	_ resource.ResourceWithConfigure      = &groupResource{}
	_ resource.ResourceWithValidateConfig = &groupResource{}
	_ resource.ResourceWithImportState    = &groupResource{}
)

func NewGroupResource() resource.Resource {
	return &groupResource{}
}

type groupResource struct {
	ConnCfg *pgx.ConnConfig
}

func (r *groupResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group"
}

func (r *groupResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = generated.GroupResourceSchema(ctx)
}

func (r *groupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan generated.GroupModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.ConnCfg.Tracer = &tracelog.TraceLog{
		Logger:   helpers.NewLogger(ctx, "group_resource.Create"),
		LogLevel: tracelog.LogLevelTrace,
	}

	var usernames []string
	if !plan.Usernames.IsUnknown() {
		diags := plan.Usernames.ElementsAs(ctx, &usernames, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	createDDL := redshift.CreateGroupDDLParams{
		Name:      plan.Name.ValueString(),
		Usernames: &usernames,
	}

	svc, err := redshift.NewGroupService(ctx, r.ConnCfg)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to execute NewGroupService",
			"An unexpected error occurred when calling NewGroupService. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"Unable to Create: "+err.Error(),
		)
		return
	}

	group, err := svc.CreateGroup(createDDL)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to execute CreateGroup on service GroupService",
			"An unexpected error occurred when calling CreateGroup on service GroupService. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"Unable to Create: "+err.Error(),
		)
		return
	}

	// Only set those undefaulted computed
	plan.Id = types.StringValue(group.Id)
	plan.Usernames = helpers.SetValueOrNull[string](ctx, types.StringType, *group.Users, &resp.Diagnostics)
	// if resp.Diagnostics.HasError() {
	// 	return
	// }
	g, diags := types.SetValueFrom(ctx, types.StringType, *group.Users)
	diags.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.Usernames = g

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *groupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state generated.GroupModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read API call logic
	r.ConnCfg.Tracer = &tracelog.TraceLog{
		Logger:   helpers.NewLogger(ctx, "group_resource.Read"),
		LogLevel: tracelog.LogLevelTrace,
	}

	svc, err := redshift.NewGroupService(ctx, r.ConnCfg)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to execute NewGroupService",
			"An unexpected error occurred when calling NewGroupService. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"Unable to Read: "+err.Error(),
		)
		return
	}

	group, err := svc.FindGroup(state.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to execute FindGroup on service GroupService",
			"An unexpected error occurred when calling FindGroup on service GroupService. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"Unable to Read: "+err.Error(),
		)
		return
	}

	// state.Usernames = helpers.SetValueOrNull[string](ctx, types.StringType, *group.Users, &resp.Diagnostics)
	// if resp.Diagnostics.HasError() {
	// 	return
	// }
	g, diags := types.SetValueFrom(ctx, types.StringType, *group.Users)
	diags.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.Usernames = g

	state.Name = types.StringValue(group.Name)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *groupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state generated.GroupModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update API call logic
	r.ConnCfg.Tracer = &tracelog.TraceLog{
		Logger:   helpers.NewLogger(ctx, "group_resource.Update"),
		LogLevel: tracelog.LogLevelTrace,
	}

	ddl := redshift.AlterGroupDDLParams{
		Name: state.Name.ValueString(),
	}

	if !plan.Name.Equal(state.Name) {
		newName := plan.Name.ValueString()
		ddl.RenameTo = &newName
	}

	if !plan.Usernames.Equal(state.Usernames) {
		var plan_usernames, state_usernames []string

		diags := plan.Usernames.ElementsAs(ctx, &plan_usernames, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		diags = state.Usernames.ElementsAs(ctx, &state_usernames, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		// if plan has more, those are add
		adds := helpers.MissingFrom(plan_usernames, state_usernames)

		// if state has more, those are drop
		drops := helpers.MissingFrom(state_usernames, plan_usernames)

		ddl.Add = &adds
		ddl.Drop = &drops

		fmt.Printf("add=%s drop=%s in=%s out=%s plan=%s state=%s", *ddl.Add, *ddl.Drop, plan_usernames, state_usernames, plan.Usernames, state.Usernames)
	}

	svc, err := redshift.NewGroupService(ctx, r.ConnCfg)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to execute NewGroupService",
			"An unexpected error occurred when calling NewGroupService. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"Unable to Update: "+err.Error(),
		)
		return
	}

	err = svc.AlterGroup(ddl)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to execute AlterGroup on service GroupService",
			"An unexpected error occurred when calling AlterGroup on service GroupService. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"Unable to Update: "+err.Error(),
		)
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *groupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state generated.GroupModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	r.ConnCfg.Tracer = &tracelog.TraceLog{
		Logger:   helpers.NewLogger(ctx, "group_resource.Delete"),
		LogLevel: tracelog.LogLevelTrace,
	}

	svc, err := redshift.NewGroupService(ctx, r.ConnCfg)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to execute NewGroupService",
			"An unexpected error occurred when calling NewGroupService. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"Unable to Delete: "+err.Error(),
		)
		return
	}

	err = svc.DropGroup(state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to execute DropGroup on service GroupService",
			"An unexpected error occurred when calling DropGroup on service GroupService. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"Unable to Delete: "+err.Error(),
		)
		return
	}
}

func (r *groupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *groupResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *groupResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var plan generated.GroupModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}
}

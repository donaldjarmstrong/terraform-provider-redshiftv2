package resources

import (
	"context"
	"fmt"
	"terraform-provider-redshift/internal/generated"
	"terraform-provider-redshift/internal/redshift"
	"terraform-provider-redshift/internal/static"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/tracelog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                   = &userResource{}
	_ resource.ResourceWithConfigure      = &userResource{}
	_ resource.ResourceWithValidateConfig = &userResource{}
	_ resource.ResourceWithImportState    = &userResource{}
)

func NewUserResource() resource.Resource {
	return &userResource{}
}

type userResource struct {
	ConnCfg *pgx.ConnConfig
}

func (r *userResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (r *userResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = generated.UserResourceSchema(ctx)
}

func (r *userResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan generated.UserModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.ConnCfg.Tracer = &tracelog.TraceLog{
		Logger:   static.NewLogger(ctx, "user_resource.Create"),
		LogLevel: tracelog.LogLevelTrace,
	}

	createDDL := redshift.CreateUserDDLParams{
		Name:            plan.Name.ValueString(),
		Password:        plan.Password.ValueStringPointer(),
		CreateDb:        plan.Createdb.ValueBool(),
		CreateUser:      plan.Createuser.ValueBool(),
		SyslogAccess:    plan.SyslogAccess.ValueString(),
		ValidUntil:      plan.ValidUntil.ValueString(),
		ConnectionLimit: plan.ConnectionLimit.ValueString(),
		SessionTimeout:  plan.SessionTimeout.ValueInt64(),
		ExternalId:      plan.ExternalId.ValueStringPointer(),
	}

	svc, err := redshift.NewUserService(ctx, r.ConnCfg)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to execute NewUserService",
			"An unexpected error occurred when calling NewUserService. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"Unable to Create: "+err.Error(),
		)
		return
	}

	user, err := svc.CreateUser(createDDL)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to execute CreateUser on service UserService",
			"An unexpected error occurred when calling CreateUser on service UserService. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"Unable to Create: "+err.Error(),
		)
		return
	}

	// Only set those undefaulted computed
	plan.Id = types.StringValue(user.Id)
	plan.ExternalId = types.StringPointerValue(user.ExternalId)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *userResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state generated.UserModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.ConnCfg.Tracer = &tracelog.TraceLog{
		Logger:   static.NewLogger(ctx, "user_resource.Read"),
		LogLevel: tracelog.LogLevelTrace,
	}

	svc, err := redshift.NewUserService(ctx, r.ConnCfg)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to execute NewUserService",
			"An unexpected error occurred when calling NewUserService. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"Unable to Read: "+err.Error(),
		)
		return
	}

	svv_data, err := svc.FindUser(state.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to execute FindUser on service UserService",
			"An unexpected error occurred when calling FindUser on service UserService. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"Unable to Read: "+err.Error(),
		)
		return
	}

	state.ConnectionLimit = types.StringValue(svv_data.ConnectionLimit)
	state.Createdb = types.BoolValue(svv_data.CreateDb)
	state.Createuser = types.BoolValue(svv_data.CreateUser)
	state.ExternalId = types.StringPointerValue(svv_data.ExternalId)
	state.Name = types.StringValue(svv_data.UserName)
	state.SessionTimeout = types.Int64Value(svv_data.SessionTimeout)
	state.SyslogAccess = types.StringValue(svv_data.SyslogAccess)
	state.ValidUntil = types.StringValue(svv_data.ValidUntil)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *userResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state generated.UserModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	r.ConnCfg.Tracer = &tracelog.TraceLog{
		Logger:   static.NewLogger(ctx, "user_resource.Update"),
		LogLevel: tracelog.LogLevelTrace,
	}

	alterUserDDL := redshift.AlterUserDDLParams{
		Name: state.Name.ValueString(),
	}

	if !plan.Name.Equal(state.Name) {
		newName := plan.Name.ValueString()
		alterUserDDL.RenameTo = &newName
	}

	if !plan.Createdb.Equal(state.Createdb) {
		if plan.Createdb.ValueBool() {
			alterUserDDL.CreateDb = static.Pointer[bool](true)
		} else {
			alterUserDDL.CreateDb = static.Pointer[bool](false)
		}
	}

	if !plan.Createuser.Equal(state.Createuser) {
		if plan.Createuser.ValueBool() {
			alterUserDDL.CreateUser = static.Pointer[bool](true)
		} else {
			alterUserDDL.CreateUser = static.Pointer[bool](false)
		}
	}

	if !plan.SyslogAccess.Equal(state.SyslogAccess) {
		alterUserDDL.SyslogAccess = plan.SyslogAccess.ValueStringPointer()
	}

	if !plan.Password.Equal(state.Password) {
		if !plan.Password.IsNull() && !plan.Password.IsUnknown() {
			alterUserDDL.Password = plan.Password.ValueStringPointer()
		} else {
			k := "" // no password
			alterUserDDL.Password = &k
		}
	}

	if !plan.ValidUntil.Equal(state.ValidUntil) {
		alterUserDDL.ValidUntil = plan.ValidUntil.ValueStringPointer()
	}

	if !plan.ConnectionLimit.Equal(state.ConnectionLimit) {
		alterUserDDL.ConnectionLimit = plan.ConnectionLimit.ValueStringPointer()
	}

	if !plan.SessionTimeout.Equal(state.SessionTimeout) {
		alterUserDDL.SessionTimeout = plan.SessionTimeout.ValueInt64Pointer()
	}

	if !plan.ExternalId.Equal(state.ExternalId) {
		alterUserDDL.ExternalId = plan.ExternalId.ValueStringPointer()
	}

	svc, err := redshift.NewUserService(ctx, r.ConnCfg)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to execute NewUserService",
			"An unexpected error occurred when calling NewUserService. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"Unable to Update: "+err.Error(),
		)
		return
	}

	err = svc.AlterUser(alterUserDDL)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to execute AlterUser on service UserService",
			"An unexpected error occurred when calling AlterUser on service UserService. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"Unable to Update: "+err.Error(),
		)
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *userResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state generated.UserModel

	// Read Terraform prior state into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	r.ConnCfg.Tracer = &tracelog.TraceLog{
		Logger:   static.NewLogger(ctx, "user_resource.Delete"),
		LogLevel: tracelog.LogLevelTrace,
	}

	svc, err := redshift.NewUserService(ctx, r.ConnCfg)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to execute NewUserService",
			"An unexpected error occurred when calling NewUserService. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"Unable to Delete: "+err.Error(),
		)
		return
	}

	err = svc.DropUser(state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to execute DropUser on service UserService",
			"An unexpected error occurred when calling DropUser on service UserService. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"Unable to Delete: "+err.Error(),
		)
		return
	}

	// If the logic reaches here, it implicitly succeeded and will remove
	// the resource from state if there are no other errors
}

func (r *userResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *userResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *userResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var plan generated.UserModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}
}

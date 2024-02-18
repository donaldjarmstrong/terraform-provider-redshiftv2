package provider

import (
	"context"
	"fmt"
	"runtime/debug"
	"terraform-provider-redshift/internal/generated"
	"terraform-provider-redshift/internal/helpers"
	"terraform-provider-redshift/internal/resources"

	"github.com/jackc/pgx/v5"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure RedshiftProvider satisfies various provider interfaces.
var (
	_ provider.ProviderWithConfigValidators = &RedshiftProvider{}
)

// RedshiftProvider defines the provider implementation.
type RedshiftProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

func (p *RedshiftProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "redshift"
	resp.Version = p.version
}

func (p *RedshiftProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = generated.RedshiftProviderSchema(ctx)
}

func (p *RedshiftProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configuring Provider")

	// Retrieve provider data from configuration
	var cfg generated.RedshiftModel
	diags := req.Config.Get(ctx, &cfg)

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if cfg.ApplicationName.IsNull() {
		info, _ := debug.ReadBuildInfo()

		ctx = tflog.SetField(ctx, "build_info", info)
		tflog.Debug(ctx, "fetched buildinfo")

		cfg.ApplicationName = types.StringValue(fmt.Sprintf("terraform-provider-redshift-%s-%s", p.version, info.GoVersion))
	}

	t := "user={{(StringValue .Username)}} password={{(StringValue .Password)}} host={{(StringValue .Host)}} port={{(Int64Value .Port)}} dbname={{(StringValue .Dbname)}} {{if (StringValue .Sslmode) -}} sslmode={{(StringValue .Sslmode)}} {{end}}application_name={{(StringValue .ApplicationName)}}{{if (Int64Value .Timeout)}} connect_timeout={{(Int64Value .Timeout)}}{{end}}"
	dsn, err := helpers.Merge(t, cfg)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create database dsn",
			"An unexpected error occurred when creating the Redshift jackc/pgx client config. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"Unable to parse config to database: "+err.Error(),
		)
		return
	}
	ctx = tflog.SetField(ctx, "dsn", dsn)
	tflog.Debug(ctx, "converted dsn")

	conn_cfg, err := pgx.ParseConfig(dsn)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Redshift jackc/pgx client config",
			"An unexpected error occurred when creating the Redshift jackc/pgx client config. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"Unable to parse config to database: "+err.Error(),
		)
		return
	}

	ctx = tflog.SetField(ctx, "timeout", conn_cfg.ConnectTimeout.Seconds())
	ctx = tflog.SetField(ctx, "params", conn_cfg.RuntimeParams)
	tflog.Debug(ctx, "dumping config")

	conn, err := pgx.ConnectConfig(ctx, conn_cfg)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Redshift jackc/pgx Client",
			"An unexpected error occurred when creating the Redshift jackc/pgx client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"Unable to connect to database: "+err.Error(),
		)
		return
	}
	defer conn.Close(ctx)

	err = conn.Ping(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to ping datasource",
			"Unexpected error "+err.Error(),
		)
	}

	resp.ResourceData = conn_cfg
}

func (p *RedshiftProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		resources.NewUserResource,
		resources.NewRoleResource,
		resources.NewGroupResource,
	}
}

func (p *RedshiftProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		// NewExampleDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &RedshiftProvider{
			version: version,
		}
	}
}

func (p RedshiftProvider) ConfigValidators(ctx context.Context) []provider.ConfigValidator {
	tflog.Info(ctx, "Configuring Validators")

	return []provider.ConfigValidator{}
}

func (p RedshiftProvider) ValidateConfig(ctx context.Context, req provider.ValidateConfigRequest, resp *provider.ValidateConfigResponse) {
	var data generated.RedshiftModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.Host.IsUnknown() {
		return
	}

	if data.Port.IsUnknown() {
		return
	}
}

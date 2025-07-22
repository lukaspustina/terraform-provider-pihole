package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const providerTypeName = "pihole"

var _ provider.Provider = &PiholeProvider{}

type PiholeProvider struct {
	version string
}

type PiholeProviderModel struct {
	URL              types.String `tfsdk:"url"`
	Password         types.String `tfsdk:"password"`
	MaxConnections   types.Int64  `tfsdk:"max_connections"`
	RequestDelay     types.Int64  `tfsdk:"request_delay_ms"`
	RetryAttempts    types.Int64  `tfsdk:"retry_attempts"`
	RetryBackoffBase types.Int64  `tfsdk:"retry_backoff_base_ms"`
}

func (p *PiholeProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = providerTypeName
	resp.Version = p.version
}

func (p *PiholeProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"url": schema.StringAttribute{
				MarkdownDescription: "Pi-hole server URL",
				Required:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "Pi-hole admin password",
				Required:            true,
				Sensitive:           true,
			},
			"max_connections": schema.Int64Attribute{
				MarkdownDescription: "Maximum number of concurrent connections to Pi-hole (default: 1)",
				Optional:            true,
			},
			"request_delay_ms": schema.Int64Attribute{
				MarkdownDescription: "Delay in milliseconds between API requests (default: 300)",
				Optional:            true,
			},
			"retry_attempts": schema.Int64Attribute{
				MarkdownDescription: "Number of retry attempts for failed requests (default: 3)",
				Optional:            true,
			},
			"retry_backoff_base_ms": schema.Int64Attribute{
				MarkdownDescription: "Base delay in milliseconds for retry backoff (default: 500)",
				Optional:            true,
			},
		},
	}
}

func (p *PiholeProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data PiholeProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Set defaults for optional parameters
	config := ClientConfig{
		MaxConnections: 1,
		RequestDelayMs: 300,
		RetryAttempts:  3,
		RetryBackoffMs: 500,
	}

	// Override defaults with user-provided values
	if !data.MaxConnections.IsNull() {
		config.MaxConnections = int(data.MaxConnections.ValueInt64())
	}
	if !data.RequestDelay.IsNull() {
		config.RequestDelayMs = int(data.RequestDelay.ValueInt64())
	}
	if !data.RetryAttempts.IsNull() {
		config.RetryAttempts = int(data.RetryAttempts.ValueInt64())
	}
	if !data.RetryBackoffBase.IsNull() {
		config.RetryBackoffMs = int(data.RetryBackoffBase.ValueInt64())
	}

	client, err := NewPiholeClient(data.URL.ValueString(), data.Password.ValueString(), config)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Pi-hole API Client",
			"An unexpected error occurred when creating the Pi-hole API client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"Pi-hole Client Error: "+err.Error(),
		)
		return
	}

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *PiholeProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewDNSRecordResource,
		NewCNAMERecordResource,
	}
}

func (p *PiholeProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &PiholeProvider{
			version: version,
		}
	}
}

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &ConfigDataSource{}

func NewConfigDataSource() datasource.DataSource {
	return &ConfigDataSource{}
}

type ConfigDataSource struct {
	client *PiholeClient
}

type ConfigDataSourceModel struct {
	Key   types.String `tfsdk:"key"`
	Value types.String `tfsdk:"value"`
	ID    types.String `tfsdk:"id"`
}

func (d *ConfigDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_config"
}

func (d *ConfigDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads Pi-hole configuration settings. " +
			"This data source can be used to read current configuration values, " +
			"such as checking if `webserver.api.app_sudo` is enabled.",

		Attributes: map[string]schema.Attribute{
			"key": schema.StringAttribute{
				MarkdownDescription: "Configuration key to read (e.g., 'webserver.api.app_sudo'). " +
					"This uses dot notation to specify nested configuration values.",
				Required: true,
			},
			"value": schema.StringAttribute{
				MarkdownDescription: "Current configuration value. Boolean values are returned as 'true' or 'false'.",
				Computed:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "Data source identifier (same as key)",
				Computed:            true,
			},
		},
	}
}

func (d *ConfigDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*PiholeClient)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *PiholeClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *ConfigDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ConfigDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	key := data.Key.ValueString()

	// Get current configuration value
	configSetting, err := d.client.GetConfig(key)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Pi-hole Configuration",
			fmt.Sprintf("Could not read configuration setting '%s': %s", key, err.Error()),
		)
		return
	}

	// Convert the value to string
	var valueStr string
	switch v := configSetting.Value.(type) {
	case bool:
		if v {
			valueStr = "true"
		} else {
			valueStr = "false"
		}
	case string:
		valueStr = v
	case float64:
		valueStr = fmt.Sprintf("%.0f", v)
	default:
		valueStr = fmt.Sprintf("%v", v)
	}

	data.Value = types.StringValue(valueStr)
	data.ID = types.StringValue(key)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

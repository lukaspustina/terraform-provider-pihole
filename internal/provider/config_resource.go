package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &ConfigResource{}
var _ resource.ResourceWithImportState = &ConfigResource{}

func NewConfigResource() resource.Resource {
	return &ConfigResource{}
}

type ConfigResource struct {
	client *PiholeClient
}

type ConfigResourceModel struct {
	Key   types.String `tfsdk:"key"`
	Value types.String `tfsdk:"value"`
	ID    types.String `tfsdk:"id"`
}

func (r *ConfigResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_config"
}

func (r *ConfigResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages Pi-hole configuration settings. " +
			"**Important**: Configuration changes require admin password, not application password. " +
			"Application passwords cannot modify Pi-hole configuration settings unless " +
			"`webserver.api.app_sudo` is enabled. This setting can be enabled via the Pi-hole web interface " +
			"under Settings → API/Web interface → \"Permit destructive actions via API\".",

		Attributes: map[string]schema.Attribute{
			"key": schema.StringAttribute{
				MarkdownDescription: "Configuration key (e.g., 'webserver.api.app_sudo'). " +
					"This uses dot notation to specify nested configuration values.",
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"value": schema.StringAttribute{
				MarkdownDescription: "Configuration value. For boolean settings, use 'true' or 'false'.",
				Required:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "Resource identifier (same as key)",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *ConfigResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*PiholeClient)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *PiholeClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *ConfigResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ConfigResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	key := data.Key.ValueString()
	value := data.Value.ValueString()

	// Convert string value to appropriate type for boolean settings
	var configValue interface{} = value
	if strings.ToLower(value) == "true" {
		configValue = true
	} else if strings.ToLower(value) == "false" {
		configValue = false
	}

	// Set the configuration using the client
	err := r.client.SetConfig(key, configValue)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Pi-hole Configuration",
			fmt.Sprintf("Could not create configuration setting '%s': %s", key, err.Error()),
		)
		return
	}

	// Set the ID to the key
	data.ID = data.Key

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConfigResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ConfigResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	key := data.Key.ValueString()

	// Get current configuration value
	configSetting, err := r.client.GetConfig(key)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Pi-hole Configuration",
			fmt.Sprintf("Could not read configuration setting '%s': %s", key, err.Error()),
		)
		return
	}

	// Convert the value back to string
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
	data.ID = data.Key

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConfigResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ConfigResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	key := data.Key.ValueString()
	value := data.Value.ValueString()

	// Convert string value to appropriate type for boolean settings
	var configValue interface{} = value
	if strings.ToLower(value) == "true" {
		configValue = true
	} else if strings.ToLower(value) == "false" {
		configValue = false
	}

	// Update the configuration using the client
	err := r.client.SetConfig(key, configValue)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Pi-hole Configuration",
			fmt.Sprintf("Could not update configuration setting '%s': %s", key, err.Error()),
		)
		return
	}

	data.ID = data.Key

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConfigResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ConfigResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// For configuration settings, we typically don't delete them but reset to default
	// For webserver.api.app_sudo, the safe default is false
	key := data.Key.ValueString()

	var defaultValue interface{} = false
	if key == "webserver.api.app_sudo" {
		defaultValue = false
	}

	err := r.client.SetConfig(key, defaultValue)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Pi-hole Configuration",
			fmt.Sprintf("Could not reset configuration setting '%s' to default: %s", key, err.Error()),
		)
		return
	}
}

func (r *ConfigResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("key"), req, resp)
}

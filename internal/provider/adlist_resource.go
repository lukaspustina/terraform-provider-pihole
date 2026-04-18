package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &AdlistResource{}
var _ resource.ResourceWithImportState = &AdlistResource{}

func NewAdlistResource() resource.Resource {
	return &AdlistResource{}
}

type AdlistResource struct {
	client *PiholeClient
}

type AdlistResourceModel struct {
	ID      types.String `tfsdk:"id"`
	Address types.String `tfsdk:"address"`
	Enabled types.Bool   `tfsdk:"enabled"`
	Comment types.String `tfsdk:"comment"`
	Type    types.String `tfsdk:"type"`
	Groups  types.List   `tfsdk:"groups"`
}

func (r *AdlistResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_adlist"
}

func (r *AdlistResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Pi-hole adlist resource for managing block and allow lists",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Adlist identifier (integer ID assigned by Pi-hole)",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"address": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "URL of the adlist",
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"enabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether the adlist is enabled (default: true)",
				Default:             booldefault.StaticBool(true),
			},
			"comment": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Optional comment describing the adlist",
				Default:             stringdefault.StaticString(""),
			},
			"type": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: `Type of list: "block" or "allow" (default: "block")`,
				Default:             stringdefault.StaticString("block"),
				Validators: []validator.String{
					stringvalidator.OneOf("block", "allow"),
				},
			},
			"groups": schema.ListAttribute{
				ElementType:         types.Int64Type,
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "List of group IDs this adlist belongs to (default: [0])",
				Default: listdefault.StaticValue(
					types.ListValueMust(types.Int64Type, []attr.Value{types.Int64Value(0)}),
				),
			},
		},
	}
}

func (r *AdlistResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *AdlistResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data AdlistResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	adlist, err := r.client.CreateAdlist(
		data.Address.ValueString(),
		data.Comment.ValueString(),
		data.Type.ValueString(),
		data.Enabled.ValueBool(),
		listToInts(data.Groups),
	)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create adlist, got error: %s", err))
		return
	}

	adlistToModel(adlist, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AdlistResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data AdlistResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.Atoi(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", fmt.Sprintf("Unable to parse adlist ID %q: %s", data.ID.ValueString(), err))
		return
	}

	adlist, err := r.client.GetAdlist(id)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read adlist %d, got error: %s", id, err))
		return
	}

	if adlist == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	adlistToModel(adlist, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AdlistResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data AdlistResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.Atoi(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", fmt.Sprintf("Unable to parse adlist ID %q: %s", data.ID.ValueString(), err))
		return
	}

	adlist, err := r.client.UpdateAdlist(
		id,
		data.Address.ValueString(),
		data.Comment.ValueString(),
		data.Type.ValueString(),
		data.Enabled.ValueBool(),
		listToInts(data.Groups),
	)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update adlist %d, got error: %s", id, err))
		return
	}

	adlistToModel(adlist, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AdlistResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data AdlistResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.Atoi(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", fmt.Sprintf("Unable to parse adlist ID %q: %s", data.ID.ValueString(), err))
		return
	}

	if err := r.client.DeleteAdlist(id); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete adlist %d, got error: %s", id, err))
		return
	}
}

func (r *AdlistResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if _, err := strconv.Atoi(req.ID); err != nil {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Adlist ID must be an integer, got: %q", req.ID),
		)
		return
	}

	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// adlistToModel populates a model from an Adlist API response.
func adlistToModel(adlist *Adlist, data *AdlistResourceModel, diagnostics *diag.Diagnostics) {
	data.ID = types.StringValue(strconv.Itoa(adlist.ID))
	data.Address = types.StringValue(adlist.Address)
	data.Enabled = types.BoolValue(adlist.Enabled)
	data.Comment = types.StringValue(adlist.Comment)
	data.Type = types.StringValue(adlist.Type)

	groupList, diags := intsToList(adlist.Groups)
	diagnostics.Append(diags...)
	if !diagnostics.HasError() {
		data.Groups = groupList
	}
}

func intsToList(ints []int) (types.List, diag.Diagnostics) {
	vals := make([]attr.Value, len(ints))
	for i, v := range ints {
		vals[i] = types.Int64Value(int64(v))
	}
	return types.ListValue(types.Int64Type, vals)
}

func listToInts(list types.List) []int {
	result := make([]int, 0, len(list.Elements()))
	for _, v := range list.Elements() {
		if i, ok := v.(types.Int64); ok {
			result = append(result, int(i.ValueInt64()))
		}
	}
	return result
}

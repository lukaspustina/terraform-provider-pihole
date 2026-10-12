package provider

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &AdlistDataSource{}

func NewAdlistDataSource() datasource.DataSource {
	return &AdlistDataSource{}
}

type AdlistDataSource struct {
	client *PiholeClient
}

type AdlistDataSourceSingleModel struct {
	ID      types.String `tfsdk:"id"`
	Address types.String `tfsdk:"address"`
	Enabled types.Bool   `tfsdk:"enabled"`
	Comment types.String `tfsdk:"comment"`
	Type    types.String `tfsdk:"type"`
	Groups  types.List   `tfsdk:"groups"`
}

func (d *AdlistDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_adlist"
}

func (d *AdlistDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves a specific adlist from Pi-hole by address URL",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Adlist ID assigned by Pi-hole",
				Computed:            true,
			},
			"address": schema.StringAttribute{
				MarkdownDescription: "URL of the adlist to look up",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether the adlist is enabled",
				Computed:            true,
			},
			"comment": schema.StringAttribute{
				MarkdownDescription: "Comment describing the adlist",
				Computed:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: `Type of list: "block" or "allow"`,
				Computed:            true,
			},
			"groups": schema.ListAttribute{
				ElementType:         types.Int64Type,
				MarkdownDescription: "Group IDs this adlist belongs to",
				Computed:            true,
			},
		},
	}
}

func (d *AdlistDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*PiholeClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			"Expected *PiholeClient, got something else",
		)
		return
	}

	d.client = client
}

func (d *AdlistDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data AdlistDataSourceSingleModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	address := data.Address.ValueString()

	adlists, err := d.client.GetAdlists()
	if err != nil {
		resp.Diagnostics.AddError("Client Error", "Unable to read adlists: "+err.Error())
		return
	}

	found := findAdlistByAddress(adlists, address)

	if found == nil {
		resp.Diagnostics.AddError(
			"Adlist Not Found",
			"No adlist found with address: "+address,
		)
		return
	}

	data.ID = types.StringValue(strconv.Itoa(found.ID))
	data.Address = types.StringValue(found.Address)
	data.Enabled = types.BoolValue(found.Enabled)
	data.Comment = types.StringValue(found.Comment)
	data.Type = types.StringValue(found.Type)

	groupList, diags := intsToList(found.Groups)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Groups = groupList

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func findAdlistByAddress(adlists []Adlist, address string) *Adlist {
	for i := range adlists {
		if adlists[i].Address == address {
			return &adlists[i]
		}
	}
	return nil
}

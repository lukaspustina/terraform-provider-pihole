package provider

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &AdlistsDataSource{}

func NewAdlistsDataSource() datasource.DataSource {
	return &AdlistsDataSource{}
}

type AdlistsDataSource struct {
	client *PiholeClient
}

type AdlistsDataSourceModel struct {
	ID      types.String            `tfsdk:"id"`
	Adlists []AdlistDataSourceModel `tfsdk:"adlists"`
}

type AdlistDataSourceModel struct {
	ID      types.String `tfsdk:"id"`
	Address types.String `tfsdk:"address"`
	Enabled types.Bool   `tfsdk:"enabled"`
	Comment types.String `tfsdk:"comment"`
	Type    types.String `tfsdk:"type"`
	Groups  types.List   `tfsdk:"groups"`
}

func (d *AdlistsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_adlists"
}

func (d *AdlistsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves all adlists (block and allow lists) from Pi-hole",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Data source identifier",
				Computed:            true,
			},
			"adlists": schema.ListNestedAttribute{
				MarkdownDescription: "List of adlists",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "Adlist ID",
							Computed:            true,
						},
						"address": schema.StringAttribute{
							MarkdownDescription: "URL of the adlist",
							Computed:            true,
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
				},
			},
		},
	}
}

func (d *AdlistsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *AdlistsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data AdlistsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	adlists, err := d.client.GetAdlists()
	if err != nil {
		resp.Diagnostics.AddError("Client Error", "Unable to read adlists: "+err.Error())
		return
	}

	models, diags := adlistsToDataSourceModels(adlists)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.ID = types.StringValue("adlists")
	data.Adlists = models

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func adlistsToDataSourceModels(adlists []Adlist) ([]AdlistDataSourceModel, diag.Diagnostics) {
	var allDiags diag.Diagnostics
	models := make([]AdlistDataSourceModel, 0, len(adlists))
	for _, al := range adlists {
		groupList, diags := intsToList(al.Groups)
		allDiags.Append(diags...)
		if allDiags.HasError() {
			return nil, allDiags
		}
		models = append(models, AdlistDataSourceModel{
			ID:      types.StringValue(strconv.Itoa(al.ID)),
			Address: types.StringValue(al.Address),
			Enabled: types.BoolValue(al.Enabled),
			Comment: types.StringValue(al.Comment),
			Type:    types.StringValue(al.Type),
			Groups:  groupList,
		})
	}
	return models, allDiags
}

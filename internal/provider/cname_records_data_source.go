package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &CNAMERecordsDataSource{}

func NewCNAMERecordsDataSource() datasource.DataSource {
	return &CNAMERecordsDataSource{}
}

type CNAMERecordsDataSource struct {
	client *PiholeClient
}

type CNAMERecordsDataSourceModel struct {
	ID      types.String                 `tfsdk:"id"`
	Records []CNAMERecordDataSourceModel `tfsdk:"records"`
}

type CNAMERecordDataSourceModel struct {
	Domain types.String `tfsdk:"domain"`
	Target types.String `tfsdk:"target"`
}

func (d *CNAMERecordsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cname_records"
}

func (d *CNAMERecordsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves all CNAME records from Pi-hole",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Data source identifier",
				Computed:            true,
			},
			"records": schema.ListNestedAttribute{
				MarkdownDescription: "List of CNAME records",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"domain": schema.StringAttribute{
							MarkdownDescription: "The CNAME domain name",
							Computed:            true,
						},
						"target": schema.StringAttribute{
							MarkdownDescription: "The target domain name",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *CNAMERecordsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *CNAMERecordsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data CNAMERecordsDataSourceModel

	// Read configuration
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get CNAME records from Pi-hole
	records, err := d.client.GetCNAMERecords()
	if err != nil {
		resp.Diagnostics.AddError("Client Error", "Unable to read CNAME records: "+err.Error())
		return
	}

	// Convert to data source model
	recordModels := make([]CNAMERecordDataSourceModel, 0, len(records))
	for _, record := range records {
		recordModels = append(recordModels, CNAMERecordDataSourceModel{
			Domain: types.StringValue(record.Domain),
			Target: types.StringValue(record.Target),
		})
	}

	data.ID = types.StringValue("cname_records")
	data.Records = recordModels

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

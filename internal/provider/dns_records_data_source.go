package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &DNSRecordsDataSource{}

func NewDNSRecordsDataSource() datasource.DataSource {
	return &DNSRecordsDataSource{}
}

type DNSRecordsDataSource struct {
	client *PiholeClient
}

type DNSRecordsDataSourceModel struct {
	ID      types.String               `tfsdk:"id"`
	Records []DNSRecordDataSourceModel `tfsdk:"records"`
}

type DNSRecordDataSourceModel struct {
	Domain types.String `tfsdk:"domain"`
	IP     types.String `tfsdk:"ip"`
}

func (d *DNSRecordsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dns_records"
}

func (d *DNSRecordsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves all DNS A records from Pi-hole",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Data source identifier",
				Computed:            true,
			},
			"records": schema.ListNestedAttribute{
				MarkdownDescription: "List of DNS A records",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"domain": schema.StringAttribute{
							MarkdownDescription: "The domain name",
							Computed:            true,
						},
						"ip": schema.StringAttribute{
							MarkdownDescription: "The IP address",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *DNSRecordsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *DNSRecordsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DNSRecordsDataSourceModel

	// Read configuration
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get DNS records from Pi-hole
	records, err := d.client.GetDNSRecords()
	if err != nil {
		resp.Diagnostics.AddError("Client Error", "Unable to read DNS records: "+err.Error())
		return
	}

	// Convert to data source model
	recordModels := make([]DNSRecordDataSourceModel, 0, len(records))
	for _, record := range records {
		recordModels = append(recordModels, DNSRecordDataSourceModel{
			Domain: types.StringValue(record.Domain),
			IP:     types.StringValue(record.IP),
		})
	}

	data.ID = types.StringValue("dns_records")
	data.Records = recordModels

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

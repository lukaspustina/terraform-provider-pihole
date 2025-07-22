package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &DNSRecordDataSource{}

func NewDNSRecordDataSource() datasource.DataSource {
	return &DNSRecordDataSource{}
}

type DNSRecordDataSource struct {
	client *PiholeClient
}

type DNSRecordDataSourceSingleModel struct {
	ID     types.String `tfsdk:"id"`
	Domain types.String `tfsdk:"domain"`
	IP     types.String `tfsdk:"ip"`
}

func (d *DNSRecordDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dns_record"
}

func (d *DNSRecordDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves a specific DNS A record from Pi-hole by domain name",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Data source identifier",
				Computed:            true,
			},
			"domain": schema.StringAttribute{
				MarkdownDescription: "The domain name to look up",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"ip": schema.StringAttribute{
				MarkdownDescription: "The IP address that the domain resolves to",
				Computed:            true,
			},
		},
	}
}

func (d *DNSRecordDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *DNSRecordDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DNSRecordDataSourceSingleModel

	// Read configuration
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := data.Domain.ValueString()

	// Get all DNS records from Pi-hole
	records, err := d.client.GetDNSRecords()
	if err != nil {
		resp.Diagnostics.AddError("Client Error", "Unable to read DNS records: "+err.Error())
		return
	}

	// Find the specific record
	var foundRecord *DNSRecord
	for _, record := range records {
		if record.Domain == domain {
			foundRecord = &record
			break
		}
	}

	if foundRecord == nil {
		resp.Diagnostics.AddError(
			"DNS Record Not Found",
			"No DNS record found for domain: "+domain,
		)
		return
	}

	// Set the data
	data.ID = types.StringValue(domain)
	data.Domain = types.StringValue(foundRecord.Domain)
	data.IP = types.StringValue(foundRecord.IP)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

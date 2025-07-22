package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &CNAMERecordDataSource{}

func NewCNAMERecordDataSource() datasource.DataSource {
	return &CNAMERecordDataSource{}
}

type CNAMERecordDataSource struct {
	client *PiholeClient
}

type CNAMERecordDataSourceSingleModel struct {
	ID     types.String `tfsdk:"id"`
	Domain types.String `tfsdk:"domain"`
	Target types.String `tfsdk:"target"`
}

func (d *CNAMERecordDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cname_record"
}

func (d *CNAMERecordDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves a specific CNAME record from Pi-hole by domain name",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Data source identifier",
				Computed:            true,
			},
			"domain": schema.StringAttribute{
				MarkdownDescription: "The CNAME domain name to look up",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"target": schema.StringAttribute{
				MarkdownDescription: "The target domain name that the CNAME points to",
				Computed:            true,
			},
		},
	}
}

func (d *CNAMERecordDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *CNAMERecordDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data CNAMERecordDataSourceSingleModel

	// Read configuration
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := data.Domain.ValueString()

	// Get all CNAME records from Pi-hole
	records, err := d.client.GetCNAMERecords()
	if err != nil {
		resp.Diagnostics.AddError("Client Error", "Unable to read CNAME records: "+err.Error())
		return
	}

	// Find the specific record
	var foundRecord *CNAMERecord
	for _, record := range records {
		if record.Domain == domain {
			foundRecord = &record
			break
		}
	}

	if foundRecord == nil {
		resp.Diagnostics.AddError(
			"CNAME Record Not Found",
			"No CNAME record found for domain: "+domain,
		)
		return
	}

	// Set the data
	data.ID = types.StringValue(domain)
	data.Domain = types.StringValue(foundRecord.Domain)
	data.Target = types.StringValue(foundRecord.Target)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

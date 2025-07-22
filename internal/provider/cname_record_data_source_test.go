package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccPiholeCNAMERecordDataSource_basic(t *testing.T) {
	testAccPreCheck(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create DNS and CNAME records first, then read with data source
			{
				Config: testAccPiholeCNAMERecordDataSourceConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the DNS target record was created
					resource.TestCheckResourceAttr("pihole_dns_record.target", "domain", "cname-target.example.com"),
					resource.TestCheckResourceAttr("pihole_dns_record.target", "ip", "192.168.1.50"),

					// Verify the CNAME record was created
					resource.TestCheckResourceAttr("pihole_cname_record.test", "domain", "cname-data-test.example.com"),
					resource.TestCheckResourceAttr("pihole_cname_record.test", "target", "cname-target.example.com"),

					// Verify the data source can find the CNAME record
					resource.TestCheckResourceAttr("data.pihole_cname_record.test", "domain", "cname-data-test.example.com"),
					resource.TestCheckResourceAttr("data.pihole_cname_record.test", "target", "cname-target.example.com"),
					resource.TestCheckResourceAttr("data.pihole_cname_record.test", "id", "cname-data-test.example.com"),
				),
			},
		},
	})
}

func TestAccPiholeCNAMERecordDataSource_notFound(t *testing.T) {
	testAccPreCheck(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Test data source with non-existent CNAME record
			{
				Config:      testAccPiholeCNAMERecordDataSourceConfig_notFound(),
				ExpectError: testExpectErrorRegex("CNAME Record Not Found"),
			},
		},
	})
}

func TestAccPiholeCNAMERecordDataSource_externalTarget(t *testing.T) {
	testAccPreCheck(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Test CNAME record pointing to external domain
			{
				Config: testAccPiholeCNAMERecordDataSourceConfig_external(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the CNAME record was created with external target
					resource.TestCheckResourceAttr("pihole_cname_record.external", "domain", "external-alias.example.com"),
					resource.TestCheckResourceAttr("pihole_cname_record.external", "target", "external.com"),

					// Verify the data source can find the external CNAME record
					resource.TestCheckResourceAttr("data.pihole_cname_record.external", "domain", "external-alias.example.com"),
					resource.TestCheckResourceAttr("data.pihole_cname_record.external", "target", "external.com"),
				),
			},
		},
	})
}

func TestAccPiholeCNAMERecordDataSource_multipleRecords(t *testing.T) {
	testAccPreCheck(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create multiple CNAME records, then query specific ones
			{
				Config: testAccPiholeCNAMERecordDataSourceConfig_multiple(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify DNS targets were created
					resource.TestCheckResourceAttr("pihole_dns_record.web_server", "domain", "web-server.example.com"),
					resource.TestCheckResourceAttr("pihole_dns_record.web_server", "ip", "192.168.1.10"),
					resource.TestCheckResourceAttr("pihole_dns_record.api_server", "domain", "api-server.example.com"),
					resource.TestCheckResourceAttr("pihole_dns_record.api_server", "ip", "192.168.1.20"),

					// Verify CNAME records were created
					resource.TestCheckResourceAttr("pihole_cname_record.www", "domain", "www.example.com"),
					resource.TestCheckResourceAttr("pihole_cname_record.www", "target", "web-server.example.com"),
					resource.TestCheckResourceAttr("pihole_cname_record.api", "domain", "api.example.com"),
					resource.TestCheckResourceAttr("pihole_cname_record.api", "target", "api-server.example.com"),

					// Verify data sources can find specific CNAME records
					resource.TestCheckResourceAttr("data.pihole_cname_record.lookup_www", "domain", "www.example.com"),
					resource.TestCheckResourceAttr("data.pihole_cname_record.lookup_www", "target", "web-server.example.com"),
					resource.TestCheckResourceAttr("data.pihole_cname_record.lookup_api", "domain", "api.example.com"),
					resource.TestCheckResourceAttr("data.pihole_cname_record.lookup_api", "target", "api-server.example.com"),
				),
			},
		},
	})
}

func TestAccPiholeCNAMERecordDataSource_chainedDependency(t *testing.T) {
	testAccPreCheck(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Test using CNAME data source to create dependent resources
			{
				Config: testAccPiholeCNAMERecordDataSourceConfig_chained(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify DNS server record
					resource.TestCheckResourceAttr("pihole_dns_record.server", "domain", "main-server.example.com"),
					resource.TestCheckResourceAttr("pihole_dns_record.server", "ip", "192.168.1.100"),

					// Verify first CNAME record
					resource.TestCheckResourceAttr("pihole_cname_record.primary_alias", "domain", "primary.example.com"),
					resource.TestCheckResourceAttr("pihole_cname_record.primary_alias", "target", "main-server.example.com"),

					// Verify data source lookup
					resource.TestCheckResourceAttr("data.pihole_cname_record.primary_lookup", "domain", "primary.example.com"),
					resource.TestCheckResourceAttr("data.pihole_cname_record.primary_lookup", "target", "main-server.example.com"),

					// Verify dependent CNAME record uses same target as data source result
					resource.TestCheckResourceAttr("pihole_cname_record.secondary_alias", "domain", "secondary.example.com"),
					resource.TestCheckResourceAttr("pihole_cname_record.secondary_alias", "target", "main-server.example.com"),
				),
			},
		},
	})
}

// Unit test for CNAME record data source schema
func TestPiholeCNAMERecordDataSource_Schema(t *testing.T) {
	ctx := testContext()
	req := testDataSourceSchemaRequest()
	resp := &testDataSourceSchemaResponse{}

	dataSource := NewCNAMERecordDataSource()
	dataSource.Schema(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema method diagnostics: %+v", resp.Diagnostics)
	}

	// Verify required attributes exist
	schema := resp.Schema
	if schema.Attributes["id"] == nil {
		t.Error("Expected id attribute in schema")
	}
	if schema.Attributes["domain"] == nil {
		t.Error("Expected domain attribute in schema")
	}
	if schema.Attributes["target"] == nil {
		t.Error("Expected target attribute in schema")
	}

	// Verify domain is required
	domainAttr := schema.Attributes["domain"]
	if !domainAttr.IsRequired() {
		t.Error("Expected domain attribute to be required")
	}

	// Verify id and target are computed
	if !schema.Attributes["id"].IsComputed() {
		t.Error("Expected id attribute to be computed")
	}
	if !schema.Attributes["target"].IsComputed() {
		t.Error("Expected target attribute to be computed")
	}
}

// Test configuration functions
func testAccPiholeCNAMERecordDataSourceConfig_basic() string {
	return fmt.Sprintf(`
%s

resource "pihole_dns_record" "target" {
  domain = "cname-target.example.com"
  ip     = "192.168.1.50"
}

resource "pihole_cname_record" "test" {
  domain = "cname-data-test.example.com"
  target = pihole_dns_record.target.domain
}

data "pihole_cname_record" "test" {
  domain = pihole_cname_record.test.domain
}
`, testAccPiholeProviderBlock())
}

func testAccPiholeCNAMERecordDataSourceConfig_notFound() string {
	return fmt.Sprintf(`
%s

data "pihole_cname_record" "not_found" {
  domain = "this-cname-does-not-exist.example.com"
}
`, testAccPiholeProviderBlock())
}

func testAccPiholeCNAMERecordDataSourceConfig_external() string {
	return fmt.Sprintf(`
%s

resource "pihole_cname_record" "external" {
  domain = "external-alias.example.com"
  target = "external.com"
}

data "pihole_cname_record" "external" {
  domain = pihole_cname_record.external.domain
}
`, testAccPiholeProviderBlock())
}

func testAccPiholeCNAMERecordDataSourceConfig_multiple() string {
	return fmt.Sprintf(`
%s

resource "pihole_dns_record" "web_server" {
  domain = "web-server.example.com"
  ip     = "192.168.1.10"
}

resource "pihole_dns_record" "api_server" {
  domain = "api-server.example.com"
  ip     = "192.168.1.20"
}

resource "pihole_cname_record" "www" {
  domain = "www.example.com"
  target = pihole_dns_record.web_server.domain
}

resource "pihole_cname_record" "api" {
  domain = "api.example.com"
  target = pihole_dns_record.api_server.domain
}

data "pihole_cname_record" "lookup_www" {
  domain = pihole_cname_record.www.domain
}

data "pihole_cname_record" "lookup_api" {
  domain = pihole_cname_record.api.domain
}
`, testAccPiholeProviderBlock())
}

func testAccPiholeCNAMERecordDataSourceConfig_chained() string {
	return fmt.Sprintf(`
%s

resource "pihole_dns_record" "server" {
  domain = "main-server.example.com"
  ip     = "192.168.1.100"
}

resource "pihole_cname_record" "primary_alias" {
  domain = "primary.example.com"
  target = pihole_dns_record.server.domain
}

data "pihole_cname_record" "primary_lookup" {
  domain = pihole_cname_record.primary_alias.domain
}

resource "pihole_cname_record" "secondary_alias" {
  domain = "secondary.example.com"
  target = data.pihole_cname_record.primary_lookup.target
}
`, testAccPiholeProviderBlock())
}

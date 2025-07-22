package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccPiholeDNSRecordDataSource_basic(t *testing.T) {
	testAccPreCheck(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create a DNS record first, then read it with data source
			{
				Config: testAccPiholeDNSRecordDataSourceConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the DNS record was created
					resource.TestCheckResourceAttr("pihole_dns_record.test", "domain", "dns-data-test.example.com"),
					resource.TestCheckResourceAttr("pihole_dns_record.test", "ip", "192.168.1.50"),

					// Verify the data source can find the record
					resource.TestCheckResourceAttr("data.pihole_dns_record.test", "domain", "dns-data-test.example.com"),
					resource.TestCheckResourceAttr("data.pihole_dns_record.test", "ip", "192.168.1.50"),
					resource.TestCheckResourceAttr("data.pihole_dns_record.test", "id", "dns-data-test.example.com"),
				),
			},
		},
	})
}

func TestAccPiholeDNSRecordDataSource_notFound(t *testing.T) {
	testAccPreCheck(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Test data source with non-existent record
			{
				Config:      testAccPiholeDNSRecordDataSourceConfig_notFound(),
				ExpectError: testExpectErrorRegex("DNS Record Not Found"),
			},
		},
	})
}

func TestAccPiholeDNSRecordDataSource_multipleRecords(t *testing.T) {
	testAccPreCheck(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create multiple DNS records, then query specific ones
			{
				Config: testAccPiholeDNSRecordDataSourceConfig_multiple(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify all DNS records were created
					resource.TestCheckResourceAttr("pihole_dns_record.server1", "domain", "server1.example.com"),
					resource.TestCheckResourceAttr("pihole_dns_record.server1", "ip", "192.168.1.10"),
					resource.TestCheckResourceAttr("pihole_dns_record.server2", "domain", "server2.example.com"),
					resource.TestCheckResourceAttr("pihole_dns_record.server2", "ip", "192.168.1.20"),

					// Verify data sources can find specific records
					resource.TestCheckResourceAttr("data.pihole_dns_record.lookup_server1", "domain", "server1.example.com"),
					resource.TestCheckResourceAttr("data.pihole_dns_record.lookup_server1", "ip", "192.168.1.10"),
					resource.TestCheckResourceAttr("data.pihole_dns_record.lookup_server2", "domain", "server2.example.com"),
					resource.TestCheckResourceAttr("data.pihole_dns_record.lookup_server2", "ip", "192.168.1.20"),
				),
			},
		},
	})
}

func TestAccPiholeDNSRecordDataSource_dependsOnResource(t *testing.T) {
	testAccPreCheck(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Test using data source output in other resources
			{
				Config: testAccPiholeDNSRecordDataSourceConfig_dependsOn(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify base record
					resource.TestCheckResourceAttr("pihole_dns_record.base", "domain", "base.example.com"),
					resource.TestCheckResourceAttr("pihole_dns_record.base", "ip", "192.168.1.100"),

					// Verify data source lookup
					resource.TestCheckResourceAttr("data.pihole_dns_record.base_lookup", "domain", "base.example.com"),
					resource.TestCheckResourceAttr("data.pihole_dns_record.base_lookup", "ip", "192.168.1.100"),

					// Verify dependent CNAME record uses data source
					resource.TestCheckResourceAttr("pihole_cname_record.alias", "domain", "alias.example.com"),
					resource.TestCheckResourceAttr("pihole_cname_record.alias", "target", "base.example.com"),
				),
			},
		},
	})
}

// Unit test for DNS record data source schema
func TestPiholeDNSRecordDataSource_Schema(t *testing.T) {
	ctx := testContext()
	req := testDataSourceSchemaRequest()
	resp := &testDataSourceSchemaResponse{}

	dataSource := NewDNSRecordDataSource()
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
	if schema.Attributes["ip"] == nil {
		t.Error("Expected ip attribute in schema")
	}

	// Verify domain is required
	domainAttr := schema.Attributes["domain"]
	if !domainAttr.IsRequired() {
		t.Error("Expected domain attribute to be required")
	}

	// Verify id and ip are computed
	if !schema.Attributes["id"].IsComputed() {
		t.Error("Expected id attribute to be computed")
	}
	if !schema.Attributes["ip"].IsComputed() {
		t.Error("Expected ip attribute to be computed")
	}
}

// Test configuration functions
func testAccPiholeDNSRecordDataSourceConfig_basic() string {
	return fmt.Sprintf(`
%s

resource "pihole_dns_record" "test" {
  domain = "dns-data-test.example.com"
  ip     = "192.168.1.50"
}

data "pihole_dns_record" "test" {
  domain = pihole_dns_record.test.domain
}
`, testAccPiholeProviderBlock())
}

func testAccPiholeDNSRecordDataSourceConfig_notFound() string {
	return fmt.Sprintf(`
%s

data "pihole_dns_record" "not_found" {
  domain = "this-domain-does-not-exist.example.com"
}
`, testAccPiholeProviderBlock())
}

func testAccPiholeDNSRecordDataSourceConfig_multiple() string {
	return fmt.Sprintf(`
%s

resource "pihole_dns_record" "server1" {
  domain = "server1.example.com"
  ip     = "192.168.1.10"
}

resource "pihole_dns_record" "server2" {
  domain = "server2.example.com"
  ip     = "192.168.1.20"
}

data "pihole_dns_record" "lookup_server1" {
  domain = pihole_dns_record.server1.domain
}

data "pihole_dns_record" "lookup_server2" {
  domain = pihole_dns_record.server2.domain
}
`, testAccPiholeProviderBlock())
}

func testAccPiholeDNSRecordDataSourceConfig_dependsOn() string {
	return fmt.Sprintf(`
%s

resource "pihole_dns_record" "base" {
  domain = "base.example.com"
  ip     = "192.168.1.100"
}

data "pihole_dns_record" "base_lookup" {
  domain = pihole_dns_record.base.domain
}

resource "pihole_cname_record" "alias" {
  domain = "alias.example.com"
  target = data.pihole_dns_record.base_lookup.domain
}
`, testAccPiholeProviderBlock())
}

// Helper function for regex error matching
func testExpectErrorRegex(pattern string) *regexp.Regexp {
	return regexp.MustCompile(pattern)
}

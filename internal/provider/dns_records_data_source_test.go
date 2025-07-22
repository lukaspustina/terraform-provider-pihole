package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccPiholeDNSRecordsDataSource_basic(t *testing.T) {
	testAccPreCheck(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Test basic data source functionality
			{
				Config: testAccPiholeDNSRecordsDataSourceConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the data source returns results
					resource.TestCheckResourceAttrSet("data.pihole_dns_records.test", "id"),
					resource.TestCheckResourceAttr("data.pihole_dns_records.test", "id", "dns_records"),
					// Check that records attribute exists (count may be 0 or more, or empty string)
					resource.TestMatchResourceAttr("data.pihole_dns_records.test", "records.#", regexp.MustCompile(`^(\d+|)$`)),
				),
			},
		},
	})
}

func TestAccPiholeDNSRecordsDataSource_withExistingRecords(t *testing.T) {
	testAccPreCheck(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create some DNS records first, then read them with data source
			{
				Config: testAccPiholeDNSRecordsDataSourceConfig_withRecords(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the resources were created
					resource.TestCheckResourceAttr("pihole_dns_record.test1", "domain", "data-test1.example.com"),
					resource.TestCheckResourceAttr("pihole_dns_record.test1", "ip", "192.168.1.10"),
					resource.TestCheckResourceAttr("pihole_dns_record.test2", "domain", "data-test2.example.com"),
					resource.TestCheckResourceAttr("pihole_dns_record.test2", "ip", "192.168.1.20"),

					// Verify the data source can find the records
					resource.TestCheckResourceAttrSet("data.pihole_dns_records.all", "records.#"),
					resource.TestMatchResourceAttr("data.pihole_dns_records.all", "records.#", regexp.MustCompile(`^[1-9]\d*$`)), // At least 1 record

					// Check that our test records appear in the data source results
					resource.TestCheckTypeSetElemNestedAttrs("data.pihole_dns_records.all", "records.*", map[string]string{
						"domain": "data-test1.example.com",
						"ip":     "192.168.1.10",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("data.pihole_dns_records.all", "records.*", map[string]string{
						"domain": "data-test2.example.com",
						"ip":     "192.168.1.20",
					}),
				),
			},
		},
	})
}

func TestAccPiholeDNSRecordsDataSource_emptyResult(t *testing.T) {
	testAccPreCheck(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Test data source when no records exist (or very few)
			{
				Config: testAccPiholeDNSRecordsDataSourceConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Should still work even if no records exist
					resource.TestCheckResourceAttr("data.pihole_dns_records.test", "id", "dns_records"),
					resource.TestMatchResourceAttr("data.pihole_dns_records.test", "records.#", regexp.MustCompile(`^(\d+|)$`)),
				),
			},
		},
	})
}

// Unit test for DNS records data source schema
func TestPiholeDNSRecordsDataSource_Schema(t *testing.T) {
	ctx := testContext()
	req := testDataSourceSchemaRequest()
	resp := &testDataSourceSchemaResponse{}

	dataSource := NewDNSRecordsDataSource()
	dataSource.Schema(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema method diagnostics: %+v", resp.Diagnostics)
	}

	// Verify required attributes exist
	schema := resp.Schema
	if schema.Attributes["id"] == nil {
		t.Error("Expected id attribute in schema")
	}
	if schema.Attributes["records"] == nil {
		t.Error("Expected records attribute in schema")
	}

	// Verify id is computed
	if !schema.Attributes["id"].IsComputed() {
		t.Error("Expected id attribute to be computed")
	}

	// Verify records is computed and is a list
	recordsAttr := schema.Attributes["records"]
	if !recordsAttr.IsComputed() {
		t.Error("Expected records attribute to be computed")
	}
}

// Test configuration functions
func testAccPiholeDNSRecordsDataSourceConfig_basic() string {
	return fmt.Sprintf(`
%s

data "pihole_dns_records" "test" {}
`, testAccPiholeProviderBlock())
}

func testAccPiholeDNSRecordsDataSourceConfig_withRecords() string {
	return fmt.Sprintf(`
%s

resource "pihole_dns_record" "test1" {
  domain = "data-test1.example.com"
  ip     = "192.168.1.10"
}

resource "pihole_dns_record" "test2" {
  domain = "data-test2.example.com"
  ip     = "192.168.1.20"
}

data "pihole_dns_records" "all" {
  depends_on = [
    pihole_dns_record.test1,
    pihole_dns_record.test2,
  ]
}
`, testAccPiholeProviderBlock())
}

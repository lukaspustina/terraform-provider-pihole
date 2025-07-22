package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccPiholeCNAMERecordsDataSource_basic(t *testing.T) {
	testAccPreCheck(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Test basic data source functionality
			{
				Config: testAccPiholeCNAMERecordsDataSourceConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the data source returns results
					resource.TestCheckResourceAttrSet("data.pihole_cname_records.test", "id"),
					resource.TestCheckResourceAttr("data.pihole_cname_records.test", "id", "cname_records"),
					// Check that records attribute exists (count may be 0 or more, or empty string)
					resource.TestMatchResourceAttr("data.pihole_cname_records.test", "records.#", regexp.MustCompile(`^(\d+|)$`)),
				),
			},
		},
	})
}

func TestAccPiholeCNAMERecordsDataSource_withExistingRecords(t *testing.T) {
	testAccPreCheck(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create DNS records and CNAME records first, then read them with data source
			{
				Config: testAccPiholeCNAMERecordsDataSourceConfig_withRecords(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the DNS records were created (targets)
					resource.TestCheckResourceAttr("pihole_dns_record.target1", "domain", "target1.example.com"),
					resource.TestCheckResourceAttr("pihole_dns_record.target1", "ip", "192.168.1.10"),
					resource.TestCheckResourceAttr("pihole_dns_record.target2", "domain", "target2.example.com"),
					resource.TestCheckResourceAttr("pihole_dns_record.target2", "ip", "192.168.1.20"),

					// Verify the CNAME records were created
					resource.TestCheckResourceAttr("pihole_cname_record.alias1", "domain", "cname-test1.example.com"),
					resource.TestCheckResourceAttr("pihole_cname_record.alias1", "target", "target1.example.com"),
					resource.TestCheckResourceAttr("pihole_cname_record.alias2", "domain", "cname-test2.example.com"),
					resource.TestCheckResourceAttr("pihole_cname_record.alias2", "target", "target2.example.com"),

					// Verify the data source can find the CNAME records
					resource.TestCheckResourceAttrSet("data.pihole_cname_records.all", "records.#"),
					resource.TestMatchResourceAttr("data.pihole_cname_records.all", "records.#", regexp.MustCompile(`^[1-9]\d*$`)), // At least 1 record

					// Check that our test CNAME records appear in the data source results
					resource.TestCheckTypeSetElemNestedAttrs("data.pihole_cname_records.all", "records.*", map[string]string{
						"domain": "cname-test1.example.com",
						"target": "target1.example.com",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("data.pihole_cname_records.all", "records.*", map[string]string{
						"domain": "cname-test2.example.com",
						"target": "target2.example.com",
					}),
				),
			},
		},
	})
}

func TestAccPiholeCNAMERecordsDataSource_mixedTargets(t *testing.T) {
	testAccPreCheck(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Test CNAME records with both internal and external targets
			{
				Config: testAccPiholeCNAMERecordsDataSourceConfig_mixedTargets(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify internal target DNS record
					resource.TestCheckResourceAttr("pihole_dns_record.internal", "domain", "internal.example.com"),
					resource.TestCheckResourceAttr("pihole_dns_record.internal", "ip", "192.168.1.100"),

					// Verify CNAME records
					resource.TestCheckResourceAttr("pihole_cname_record.internal_alias", "domain", "internal-alias.example.com"),
					resource.TestCheckResourceAttr("pihole_cname_record.internal_alias", "target", "internal.example.com"),
					resource.TestCheckResourceAttr("pihole_cname_record.external_alias", "domain", "external-alias.example.com"),
					resource.TestCheckResourceAttr("pihole_cname_record.external_alias", "target", "external.com"),

					// Verify the data source includes both types
					resource.TestCheckTypeSetElemNestedAttrs("data.pihole_cname_records.mixed", "records.*", map[string]string{
						"domain": "internal-alias.example.com",
						"target": "internal.example.com",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("data.pihole_cname_records.mixed", "records.*", map[string]string{
						"domain": "external-alias.example.com",
						"target": "external.com",
					}),
				),
			},
		},
	})
}

func TestAccPiholeCNAMERecordsDataSource_emptyResult(t *testing.T) {
	testAccPreCheck(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Test data source when no CNAME records exist
			{
				Config: testAccPiholeCNAMERecordsDataSourceConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Should still work even if no CNAME records exist
					resource.TestCheckResourceAttr("data.pihole_cname_records.test", "id", "cname_records"),
					resource.TestMatchResourceAttr("data.pihole_cname_records.test", "records.#", regexp.MustCompile(`^(\d+|)$`)),
				),
			},
		},
	})
}

// Unit test for CNAME records data source schema
func TestPiholeCNAMERecordsDataSource_Schema(t *testing.T) {
	ctx := testContext()
	req := testDataSourceSchemaRequest()
	resp := &testDataSourceSchemaResponse{}

	dataSource := NewCNAMERecordsDataSource()
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

	// Verify records is computed
	recordsAttr := schema.Attributes["records"]
	if !recordsAttr.IsComputed() {
		t.Error("Expected records attribute to be computed")
	}
}

// Test configuration functions
func testAccPiholeCNAMERecordsDataSourceConfig_basic() string {
	return fmt.Sprintf(`
%s

data "pihole_cname_records" "test" {}
`, testAccPiholeProviderBlock())
}

func testAccPiholeCNAMERecordsDataSourceConfig_withRecords() string {
	return fmt.Sprintf(`
%s

# Create target DNS records
resource "pihole_dns_record" "target1" {
  domain = "target1.example.com"
  ip     = "192.168.1.10"
}

resource "pihole_dns_record" "target2" {
  domain = "target2.example.com"
  ip     = "192.168.1.20"
}

# Create CNAME records pointing to targets
resource "pihole_cname_record" "alias1" {
  domain = "cname-test1.example.com"
  target = pihole_dns_record.target1.domain
}

resource "pihole_cname_record" "alias2" {
  domain = "cname-test2.example.com"
  target = pihole_dns_record.target2.domain
}

data "pihole_cname_records" "all" {
  depends_on = [
    pihole_cname_record.alias1,
    pihole_cname_record.alias2,
  ]
}
`, testAccPiholeProviderBlock())
}

func testAccPiholeCNAMERecordsDataSourceConfig_mixedTargets() string {
	return fmt.Sprintf(`
%s

# Create internal target DNS record
resource "pihole_dns_record" "internal" {
  domain = "internal.example.com"
  ip     = "192.168.1.100"
}

# Create CNAME pointing to internal target
resource "pihole_cname_record" "internal_alias" {
  domain = "internal-alias.example.com"
  target = pihole_dns_record.internal.domain
}

# Create CNAME pointing to external target
resource "pihole_cname_record" "external_alias" {
  domain = "external-alias.example.com"
  target = "external.com"
}

data "pihole_cname_records" "mixed" {
  depends_on = [
    pihole_cname_record.internal_alias,
    pihole_cname_record.external_alias,
  ]
}
`, testAccPiholeProviderBlock())
}

package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccPiholeDataSources_integration tests all data sources working together
func TestAccPiholeDataSources_integration(t *testing.T) {
	testAccPreCheck(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPiholeDataSourcesIntegrationConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify DNS records were created
					resource.TestCheckResourceAttr("pihole_dns_record.server1", "domain", "integration-server1.example.com"),
					resource.TestCheckResourceAttr("pihole_dns_record.server1", "ip", "192.168.1.10"),
					resource.TestCheckResourceAttr("pihole_dns_record.server2", "domain", "integration-server2.example.com"),
					resource.TestCheckResourceAttr("pihole_dns_record.server2", "ip", "192.168.1.20"),

					// Verify CNAME records were created
					resource.TestCheckResourceAttr("pihole_cname_record.www1", "domain", "www1.example.com"),
					resource.TestCheckResourceAttr("pihole_cname_record.www1", "target", "integration-server1.example.com"),
					resource.TestCheckResourceAttr("pihole_cname_record.www2", "domain", "www2.example.com"),
					resource.TestCheckResourceAttr("pihole_cname_record.www2", "target", "integration-server2.example.com"),

					// Test pihole_dns_records data source
					resource.TestCheckResourceAttrSet("data.pihole_dns_records.all", "records.#"),
					resource.TestMatchResourceAttr("data.pihole_dns_records.all", "records.#", regexp.MustCompile(`^[1-9]\d*$`)),

					// Verify our test DNS records appear in the all records data source
					resource.TestCheckTypeSetElemNestedAttrs("data.pihole_dns_records.all", "records.*", map[string]string{
						"domain": "integration-server1.example.com",
						"ip":     "192.168.1.10",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("data.pihole_dns_records.all", "records.*", map[string]string{
						"domain": "integration-server2.example.com",
						"ip":     "192.168.1.20",
					}),

					// Test pihole_cname_records data source
					resource.TestCheckResourceAttrSet("data.pihole_cname_records.all", "records.#"),
					resource.TestMatchResourceAttr("data.pihole_cname_records.all", "records.#", regexp.MustCompile(`^[1-9]\d*$`)),

					// Verify our test CNAME records appear in the all records data source
					resource.TestCheckTypeSetElemNestedAttrs("data.pihole_cname_records.all", "records.*", map[string]string{
						"domain": "www1.example.com",
						"target": "integration-server1.example.com",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("data.pihole_cname_records.all", "records.*", map[string]string{
						"domain": "www2.example.com",
						"target": "integration-server2.example.com",
					}),

					// Test individual DNS record data sources
					resource.TestCheckResourceAttr("data.pihole_dns_record.server1_lookup", "domain", "integration-server1.example.com"),
					resource.TestCheckResourceAttr("data.pihole_dns_record.server1_lookup", "ip", "192.168.1.10"),
					resource.TestCheckResourceAttr("data.pihole_dns_record.server2_lookup", "domain", "integration-server2.example.com"),
					resource.TestCheckResourceAttr("data.pihole_dns_record.server2_lookup", "ip", "192.168.1.20"),

					// Test individual CNAME record data sources
					resource.TestCheckResourceAttr("data.pihole_cname_record.www1_lookup", "domain", "www1.example.com"),
					resource.TestCheckResourceAttr("data.pihole_cname_record.www1_lookup", "target", "integration-server1.example.com"),
					resource.TestCheckResourceAttr("data.pihole_cname_record.www2_lookup", "domain", "www2.example.com"),
					resource.TestCheckResourceAttr("data.pihole_cname_record.www2_lookup", "target", "integration-server2.example.com"),
				),
			},
		},
	})
}

// TestAccPiholeDataSources_dynamic tests using data sources for dynamic resource creation
func TestAccPiholeDataSources_dynamic(t *testing.T) {
	testAccPreCheck(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPiholeDataSourcesDynamicConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify base infrastructure
					resource.TestCheckResourceAttr("pihole_dns_record.base_server", "domain", "base.example.com"),
					resource.TestCheckResourceAttr("pihole_dns_record.base_server", "ip", "192.168.1.100"),

					// Verify data source lookup works
					resource.TestCheckResourceAttr("data.pihole_dns_record.base_lookup", "domain", "base.example.com"),
					resource.TestCheckResourceAttr("data.pihole_dns_record.base_lookup", "ip", "192.168.1.100"),

					// Verify dependent resources were created using data source
					resource.TestCheckResourceAttr("pihole_cname_record.api_alias", "domain", "api.example.com"),
					resource.TestCheckResourceAttr("pihole_cname_record.api_alias", "target", "base.example.com"),
					resource.TestCheckResourceAttr("pihole_cname_record.web_alias", "domain", "web.example.com"),
					resource.TestCheckResourceAttr("pihole_cname_record.web_alias", "target", "base.example.com"),

					// Verify dependent CNAME data source lookup
					resource.TestCheckResourceAttr("data.pihole_cname_record.api_lookup", "domain", "api.example.com"),
					resource.TestCheckResourceAttr("data.pihole_cname_record.api_lookup", "target", "base.example.com"),

					// Verify monitoring record created with computed IP
					resource.TestCheckResourceAttr("pihole_dns_record.monitoring", "domain", "monitor.example.com"),
					resource.TestCheckResourceAttr("pihole_dns_record.monitoring", "ip", "192.168.1.200"),
				),
			},
		},
	})
}

// TestAccPiholeDataSources_errorHandling tests error conditions
func TestAccPiholeDataSources_errorHandling(t *testing.T) {
	testAccPreCheck(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Test non-existent DNS record
			{
				Config:      testAccPiholeDataSourcesErrorConfig_dnsNotFound(),
				ExpectError: regexp.MustCompile("DNS Record Not Found"),
			},
			// Test non-existent CNAME record
			{
				Config:      testAccPiholeDataSourcesErrorConfig_cnameNotFound(),
				ExpectError: regexp.MustCompile("CNAME Record Not Found"),
			},
		},
	})
}

// TestAccPiholeDataSources_validation tests input validation
func TestAccPiholeDataSources_validation(t *testing.T) {
	testAccPreCheck(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Test empty domain name for DNS record data source
			{
				Config:      testAccPiholeDataSourcesValidationConfig_emptyDNSDomain(),
				ExpectError: regexp.MustCompile("Attribute domain string length must be at least 1"),
			},
			// Test empty domain name for CNAME record data source
			{
				Config:      testAccPiholeDataSourcesValidationConfig_emptyCNAMEDomain(),
				ExpectError: regexp.MustCompile("Attribute domain string length must be at least 1"),
			},
		},
	})
}

// TestAccPiholeDataSources_pagination tests handling of large result sets
func TestAccPiholeDataSources_pagination(t *testing.T) {
	testAccPreCheck(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPiholeDataSourcesPaginationConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Create several records and verify data sources can handle them
					resource.TestCheckResourceAttrSet("data.pihole_dns_records.many", "records.#"),
					resource.TestCheckResourceAttrSet("data.pihole_cname_records.many", "records.#"),

					// Verify at least our test records are present
					resource.TestMatchResourceAttr("data.pihole_dns_records.many", "records.#", regexp.MustCompile(`^[5-9]|[1-9]\d+$`)),   // At least 5
					resource.TestMatchResourceAttr("data.pihole_cname_records.many", "records.#", regexp.MustCompile(`^[3-9]|[1-9]\d+$`)), // At least 3
				),
			},
		},
	})
}

// Configuration functions for acceptance tests

func testAccPiholeDataSourcesIntegrationConfig() string {
	return fmt.Sprintf(`
%s

# Create DNS records
resource "pihole_dns_record" "server1" {
  domain = "integration-server1.example.com"
  ip     = "192.168.1.10"
}

resource "pihole_dns_record" "server2" {
  domain = "integration-server2.example.com"
  ip     = "192.168.1.20"
}

# Create CNAME records
resource "pihole_cname_record" "www1" {
  domain = "www1.example.com"
  target = pihole_dns_record.server1.domain
}

resource "pihole_cname_record" "www2" {
  domain = "www2.example.com"
  target = pihole_dns_record.server2.domain
}

# Test all data sources
data "pihole_dns_records" "all" {
  depends_on = [
    pihole_dns_record.server1,
    pihole_dns_record.server2,
  ]
}

data "pihole_cname_records" "all" {
  depends_on = [
    pihole_cname_record.www1,
    pihole_cname_record.www2,
  ]
}

data "pihole_dns_record" "server1_lookup" {
  domain = pihole_dns_record.server1.domain
}

data "pihole_dns_record" "server2_lookup" {
  domain = pihole_dns_record.server2.domain
}

data "pihole_cname_record" "www1_lookup" {
  domain = pihole_cname_record.www1.domain
}

data "pihole_cname_record" "www2_lookup" {
  domain = pihole_cname_record.www2.domain
}
`, testAccPiholeProviderBlock())
}

func testAccPiholeDataSourcesDynamicConfig() string {
	return fmt.Sprintf(`
%s

# Base infrastructure
resource "pihole_dns_record" "base_server" {
  domain = "base.example.com"
  ip     = "192.168.1.100"
}

# Look up the base server
data "pihole_dns_record" "base_lookup" {
  domain = pihole_dns_record.base_server.domain
}

# Create dependent resources using data source
resource "pihole_cname_record" "api_alias" {
  domain = "api.example.com"
  target = data.pihole_dns_record.base_lookup.domain
}

resource "pihole_cname_record" "web_alias" {
  domain = "web.example.com"
  target = data.pihole_dns_record.base_lookup.domain
}

# Look up a CNAME record
data "pihole_cname_record" "api_lookup" {
  domain = pihole_cname_record.api_alias.domain
}

# Create monitoring record with computed IP (base IP + 100)
resource "pihole_dns_record" "monitoring" {
  domain = "monitor.example.com"
  ip     = "192.168.1.200"  # Could be computed from base in real scenario
}
`, testAccPiholeProviderBlock())
}

func testAccPiholeDataSourcesErrorConfig_dnsNotFound() string {
	return fmt.Sprintf(`
%s

data "pihole_dns_record" "nonexistent" {
  domain = "this-absolutely-does-not-exist.example.com"
}
`, testAccPiholeProviderBlock())
}

func testAccPiholeDataSourcesErrorConfig_cnameNotFound() string {
	return fmt.Sprintf(`
%s

data "pihole_cname_record" "nonexistent" {
  domain = "this-cname-absolutely-does-not-exist.example.com"
}
`, testAccPiholeProviderBlock())
}

func testAccPiholeDataSourcesValidationConfig_emptyDNSDomain() string {
	return fmt.Sprintf(`
%s

data "pihole_dns_record" "empty_domain" {
  domain = ""
}
`, testAccPiholeProviderBlock())
}

func testAccPiholeDataSourcesValidationConfig_emptyCNAMEDomain() string {
	return fmt.Sprintf(`
%s

data "pihole_cname_record" "empty_domain" {
  domain = ""
}
`, testAccPiholeProviderBlock())
}

func testAccPiholeDataSourcesPaginationConfig() string {
	return fmt.Sprintf(`
%s

# Create multiple DNS records to test pagination
resource "pihole_dns_record" "many" {
  count = 5
  
  domain = "test-many-${count.index + 1}.example.com"
  ip     = "192.168.2.${count.index + 10}"
}

# Create multiple CNAME records  
resource "pihole_cname_record" "many" {
  count = 3
  
  domain = "alias-many-${count.index + 1}.example.com"
  target = pihole_dns_record.many[count.index].domain
}

# Test data sources with many records
data "pihole_dns_records" "many" {
  depends_on = [pihole_dns_record.many]
}

data "pihole_cname_records" "many" {
  depends_on = [pihole_cname_record.many]
}
`, testAccPiholeProviderBlock())
}

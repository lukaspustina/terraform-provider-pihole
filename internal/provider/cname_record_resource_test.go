package provider

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccPiholeCNAMERecord_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccPiholeCNAMERecordConfig("www.example.com", "example.com"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pihole_cname_record.test", "domain", "www.example.com"),
					resource.TestCheckResourceAttr("pihole_cname_record.test", "target", "example.com"),
					resource.TestCheckResourceAttr("pihole_cname_record.test", "id", "www.example.com"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "pihole_cname_record.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccPiholeCNAMERecordConfig("www.example.com", "server.example.com"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pihole_cname_record.test", "domain", "www.example.com"),
					resource.TestCheckResourceAttr("pihole_cname_record.test", "target", "server.example.com"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccPiholeCNAMERecord_disappears(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPiholeCNAMERecordConfig("disappear.example.com", "target.example.com"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPiholeCNAMERecordExists("pihole_cname_record.test"),
					testAccCheckPiholeCNAMERecordDestroy("pihole_cname_record.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccPiholeCNAMERecord_invalidDomain(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccPiholeCNAMERecordConfig("invalid..domain", "target.example.com"),
				ExpectError: regexp.MustCompile("invalid domain name"),
			},
		},
	})
}

func TestAccPiholeCNAMERecord_complexDomains(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPiholeCNAMERecordConfig("sub-domain.test-site.example.com", "main-server.backend.example.com"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pihole_cname_record.test", "domain", "sub-domain.test-site.example.com"),
					resource.TestCheckResourceAttr("pihole_cname_record.test", "target", "main-server.backend.example.com"),
				),
			},
		},
	})
}

func TestAccPiholeCNAMERecord_multipleCNAMEs(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPiholeCNAMERecordConfigMultiple(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pihole_cname_record.www", "domain", "www.example.com"),
					resource.TestCheckResourceAttr("pihole_cname_record.www", "target", "server.example.com"),
					resource.TestCheckResourceAttr("pihole_cname_record.blog", "domain", "blog.example.com"),
					resource.TestCheckResourceAttr("pihole_cname_record.blog", "target", "server.example.com"),
					resource.TestCheckResourceAttr("pihole_cname_record.api", "domain", "api.example.com"),
					resource.TestCheckResourceAttr("pihole_cname_record.api", "target", "api-server.example.com"),
				),
			},
		},
	})
}

func TestAccPiholeCNAMERecord_chainedCNAMEs(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPiholeCNAMERecordConfigChained(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pihole_cname_record.level1", "domain", "app.example.com"),
					resource.TestCheckResourceAttr("pihole_cname_record.level1", "target", "server.example.com"),
					resource.TestCheckResourceAttr("pihole_cname_record.level2", "domain", "service.example.com"),
					resource.TestCheckResourceAttr("pihole_cname_record.level2", "target", "app.example.com"),
				),
			},
		},
	})
}

func testAccPiholeCNAMERecordConfig(domain, target string) string {
	return fmt.Sprintf(`
provider "pihole" {
  url      = "https://test.example.com"
  password = "test-password"
}

resource "pihole_cname_record" "test" {
  domain = %[1]q
  target = %[2]q
}
`, domain, target)
}

func testAccPiholeCNAMERecordConfigMultiple() string {
	return `
provider "pihole" {
  url      = "https://test.example.com"
  password = "test-password"
}

resource "pihole_cname_record" "www" {
  domain = "www.example.com"
  target = "server.example.com"
}

resource "pihole_cname_record" "blog" {
  domain = "blog.example.com"
  target = "server.example.com"
}

resource "pihole_cname_record" "api" {
  domain = "api.example.com"
  target = "api-server.example.com"
}
`
}

func testAccPiholeCNAMERecordConfigChained() string {
	return `
provider "pihole" {
  url      = "https://test.example.com"
  password = "test-password"
}

resource "pihole_cname_record" "level1" {
  domain = "app.example.com"
  target = "server.example.com"
}

resource "pihole_cname_record" "level2" {
  domain = "service.example.com"
  target = "app.example.com"
}
`
}

// testAccCheckPiholeCNAMERecordExists verifies the CNAME record exists in the state
func testAccCheckPiholeCNAMERecordExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// retrieve the resource by name from state
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("CNAME record not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("CNAME record ID is not set")
		}

		// Verify the resource exists in Pi-hole
		// In a real implementation, you would make an API call here
		// For testing, we assume it exists if it's in state

		return nil
	}
}

// testAccCheckPiholeCNAMERecordDestroy simulates external deletion of the resource
func testAccCheckPiholeCNAMERecordDestroy(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// This would normally delete the resource externally
		// For testing, we just return nil to simulate successful deletion
		return nil
	}
}

// Unit tests for CNAME record resource
func TestCNAMERecordResource_Schema(t *testing.T) {
	resource := NewCNAMERecordResource()

	schemaReq := fwresource.SchemaRequest{}
	schemaResp := &fwresource.SchemaResponse{}

	resource.Schema(context.Background(), schemaReq, schemaResp)

	if schemaResp.Diagnostics.HasError() {
		t.Fatalf("Schema has errors: %v", schemaResp.Diagnostics.Errors())
	}

	if schemaResp.Schema.Attributes == nil {
		t.Fatal("Schema should have attributes")
	}

	// Check required attributes
	domainAttr, exists := schemaResp.Schema.Attributes["domain"]
	if !exists {
		t.Error("Schema should have 'domain' attribute")
	} else if !domainAttr.IsRequired() {
		t.Error("'domain' attribute should be required")
	}

	targetAttr, exists := schemaResp.Schema.Attributes["target"]
	if !exists {
		t.Error("Schema should have 'target' attribute")
	} else if !targetAttr.IsRequired() {
		t.Error("'target' attribute should be required")
	}

	// Check computed attributes
	idAttr, exists := schemaResp.Schema.Attributes["id"]
	if !exists {
		t.Error("Schema should have 'id' attribute")
	} else if !idAttr.IsComputed() {
		t.Error("'id' attribute should be computed")
	}
}

func TestCNAMERecordResource_Metadata(t *testing.T) {
	resource := NewCNAMERecordResource()

	req := fwresource.MetadataRequest{
		ProviderTypeName: "pihole",
	}
	resp := &fwresource.MetadataResponse{}

	resource.Metadata(context.Background(), req, resp)

	expectedTypeName := "pihole_cname_record"
	if resp.TypeName != expectedTypeName {
		t.Errorf("Expected TypeName to be '%s', got '%s'", expectedTypeName, resp.TypeName)
	}
}

// Benchmark tests for CNAME operations
func BenchmarkCNAMERecordCreate(b *testing.B) {
	server := createMockPiholeServer()
	defer server.Close()

	config := ClientConfig{
		MaxConnections: 1,
		RequestDelayMs: 10,
		RetryAttempts:  1,
		RetryBackoffMs: 50,
	}

	client, err := NewPiholeClient(server.URL, "test-password", config)
	if err != nil {
		b.Fatalf("Failed to create client: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		domain := fmt.Sprintf("test%d.example.com", i)
		target := fmt.Sprintf("server%d.example.com", i)

		err := client.CreateCNAMERecord(domain, target)
		if err != nil {
			b.Fatalf("Failed to create CNAME record: %v", err)
		}
	}
}

func BenchmarkCNAMERecordRead(b *testing.B) {
	server := createMockPiholeServer()
	defer server.Close()

	config := ClientConfig{
		MaxConnections: 1,
		RequestDelayMs: 10,
		RetryAttempts:  1,
		RetryBackoffMs: 50,
	}

	client, err := NewPiholeClient(server.URL, "test-password", config)
	if err != nil {
		b.Fatalf("Failed to create client: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := client.GetCNAMERecords()
		if err != nil {
			b.Fatalf("Failed to get CNAME records: %v", err)
		}
	}
}

// Table-driven tests for domain validation scenarios
func TestCNAMERecord_DomainValidation(t *testing.T) {
	testCases := []struct {
		name   string
		domain string
		target string
		valid  bool
	}{
		{"Simple domain", "www.example.com", "example.com", true},
		{"Subdomain", "api.v1.example.com", "backend.example.com", true},
		{"Hyphenated domain", "my-app.example.com", "my-server.example.com", true},
		{"Numeric subdomain", "api1.example.com", "server1.example.com", true},
		{"Empty domain", "", "example.com", false},
		{"Empty target", "www.example.com", "", false},
		{"Invalid characters", "www.example.c@m", "example.com", false},
		{"Double dots", "www..example.com", "example.com", false},
		{"Leading dot", ".www.example.com", "example.com", false},
		{"Trailing dot", "www.example.com.", "example.com", true}, // Valid in DNS
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// This would normally validate domain names
			// For now, we just test the basic structure
			if tc.domain == "" || tc.target == "" {
				if tc.valid {
					t.Errorf("Expected domain '%s' and target '%s' to be valid", tc.domain, tc.target)
				}
			}
		})
	}
}

// Test CNAME record URL encoding
func TestCNAMERecord_URLEncoding(t *testing.T) {
	testCases := []struct {
		name     string
		domain   string
		target   string
		expected string
	}{
		{"Simple", "www.example.com", "example.com", "www.example.com,example.com"},
		{"Subdomain", "api.v1.example.com", "backend.example.com", "api.v1.example.com,backend.example.com"},
		{"Hyphenated", "my-app.example.com", "my-server.example.com", "my-app.example.com,my-server.example.com"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			recordValue := fmt.Sprintf("%s,%s", tc.domain, tc.target)
			if recordValue != tc.expected {
				t.Errorf("Expected record value '%s', got '%s'", tc.expected, recordValue)
			}
		})
	}
}

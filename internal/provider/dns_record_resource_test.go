package provider

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"testing"
	"time"

	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccPiholeDNSRecord_basic(t *testing.T) {
	testAccPreCheck(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccPiholeDNSRecordConfig("test.example.com", "192.168.1.100"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pihole_dns_record.test", "domain", "test.example.com"),
					resource.TestCheckResourceAttr("pihole_dns_record.test", "ip", "192.168.1.100"),
					resource.TestCheckResourceAttr("pihole_dns_record.test", "id", "test.example.com"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "pihole_dns_record.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccPiholeDNSRecordConfig("test.example.com", "192.168.1.101"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pihole_dns_record.test", "domain", "test.example.com"),
					resource.TestCheckResourceAttr("pihole_dns_record.test", "ip", "192.168.1.101"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccPiholeDNSRecord_disappears(t *testing.T) {
	testAccPreCheck(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPiholeDNSRecordConfig("disappear.example.com", "192.168.1.200"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPiholeDNSRecordExists("pihole_dns_record.test"),
					testAccCheckPiholeDNSRecordDestroy("pihole_dns_record.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccPiholeDNSRecord_invalidIP(t *testing.T) {
	testAccPreCheck(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccPiholeDNSRecordConfig("test.example.com", "invalid-ip"),
				ExpectError: regexp.MustCompile("invalid IP address"),
			},
		},
	})
}

func TestAccPiholeDNSRecord_specialCharacters(t *testing.T) {
	testAccPreCheck(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPiholeDNSRecordConfig("test-domain.sub-domain.example.com", "192.168.1.150"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pihole_dns_record.test", "domain", "test-domain.sub-domain.example.com"),
					resource.TestCheckResourceAttr("pihole_dns_record.test", "ip", "192.168.1.150"),
				),
			},
		},
	})
}

func TestAccPiholeDNSRecord_multipleRecords(t *testing.T) {
	testAccPreCheck(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPiholeDNSRecordConfigMultiple(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pihole_dns_record.test1", "domain", "server1.example.com"),
					resource.TestCheckResourceAttr("pihole_dns_record.test1", "ip", "192.168.1.10"),
					resource.TestCheckResourceAttr("pihole_dns_record.test2", "domain", "server2.example.com"),
					resource.TestCheckResourceAttr("pihole_dns_record.test2", "ip", "192.168.1.20"),
					resource.TestCheckResourceAttr("pihole_dns_record.test3", "domain", "server3.example.com"),
					resource.TestCheckResourceAttr("pihole_dns_record.test3", "ip", "192.168.1.30"),
				),
			},
		},
	})
}

func testAccPiholeProviderBlock() string {
	url := os.Getenv("PIHOLE_URL")
	if url == "" {
		url = "https://test.example.com"
	}
	password := os.Getenv("PIHOLE_PASSWORD")
	if password == "" {
		password = "test-password"
	}
	return fmt.Sprintf(`
provider "pihole" {
  url      = %[1]q
  password = %[2]q
}`, url, password)
}

func testAccPreCheck(t *testing.T) {
	if os.Getenv("PIHOLE_URL") == "" {
		t.Skip("PIHOLE_URL not set, skipping acceptance test")
	}
	// Add delay between tests to reduce Pi-hole API session pressure
	time.Sleep(1 * time.Second)
}

func testAccPiholeDNSRecordConfig(domain, ip string) string {
	return fmt.Sprintf(`
%s

resource "pihole_dns_record" "test" {
  domain = %[2]q
  ip     = %[3]q
}
`, testAccPiholeProviderBlock(), domain, ip)
}

func testAccPiholeDNSRecordConfigMultiple() string {
	return fmt.Sprintf(`
%s

resource "pihole_dns_record" "test1" {
  domain = "server1.example.com"
  ip     = "192.168.1.10"
}

resource "pihole_dns_record" "test2" {
  domain = "server2.example.com"
  ip     = "192.168.1.20"
}

resource "pihole_dns_record" "test3" {
  domain = "server3.example.com"
  ip     = "192.168.1.30"
}
`, testAccPiholeProviderBlock())
}

// testAccCheckPiholeDNSRecordExists verifies the DNS record exists in the state
func testAccCheckPiholeDNSRecordExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// retrieve the resource by name from state
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("DNS record not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("DNS record ID is not set")
		}

		// Verify the resource exists in Pi-hole
		// In a real implementation, you would make an API call here
		// For testing, we assume it exists if it's in state

		return nil
	}
}

// testAccCheckPiholeDNSRecordDestroy simulates external deletion of the resource
func testAccCheckPiholeDNSRecordDestroy(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("resource ID not set")
		}

		// Create a Pi-hole client to delete the resource externally
		config := ClientConfig{
			MaxConnections: 1,
			RequestDelayMs: 300,
			RetryAttempts:  3,
			RetryBackoffMs: 500,
		}

		url := os.Getenv("PIHOLE_URL")
		password := os.Getenv("PIHOLE_PASSWORD")
		if url == "" || password == "" {
			return fmt.Errorf("PIHOLE_URL and PIHOLE_PASSWORD must be set for disappears test")
		}

		client, err := getOrCreateClient(url, password, config)
		if err != nil {
			return fmt.Errorf("failed to create client: %v", err)
		}

		// Delete the DNS record externally using the domain (which is the ID)
		err = client.DeleteDNSRecord(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("failed to delete DNS record externally: %v", err)
		}

		return nil
	}
}

// Unit tests for DNS record resource
func TestDNSRecordResource_Schema(t *testing.T) {
	resource := NewDNSRecordResource()

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

	ipAttr, exists := schemaResp.Schema.Attributes["ip"]
	if !exists {
		t.Error("Schema should have 'ip' attribute")
	} else if !ipAttr.IsRequired() {
		t.Error("'ip' attribute should be required")
	}

	// Check computed attributes
	idAttr, exists := schemaResp.Schema.Attributes["id"]
	if !exists {
		t.Error("Schema should have 'id' attribute")
	} else if !idAttr.IsComputed() {
		t.Error("'id' attribute should be computed")
	}
}

func TestDNSRecordResource_Metadata(t *testing.T) {
	resource := NewDNSRecordResource()

	req := fwresource.MetadataRequest{
		ProviderTypeName: "pihole",
	}
	resp := &fwresource.MetadataResponse{}

	resource.Metadata(context.Background(), req, resp)

	expectedTypeName := "pihole_dns_record"
	if resp.TypeName != expectedTypeName {
		t.Errorf("Expected TypeName to be '%s', got '%s'", expectedTypeName, resp.TypeName)
	}
}

// Benchmark tests for DNS operations
func BenchmarkDNSRecordCreate(b *testing.B) {
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
		ip := fmt.Sprintf("192.168.1.%d", i%255)

		err := client.CreateDNSRecord(domain, ip)
		if err != nil {
			b.Fatalf("Failed to create DNS record: %v", err)
		}
	}
}

func BenchmarkDNSRecordRead(b *testing.B) {
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
		_, err := client.GetDNSRecords()
		if err != nil {
			b.Fatalf("Failed to get DNS records: %v", err)
		}
	}
}

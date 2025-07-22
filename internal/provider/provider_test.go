package provider

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"pihole": providerserver.NewProtocol6WithError(New("test")()),
}

func TestAccPiholeProvider(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Verify provider can be configured
			{
				Config: testAccPiholeProviderConfig(),
				Check:  resource.ComposeAggregateTestCheckFunc(
				// Verify provider is properly configured
				),
			},
		},
	})
}

func testAccPiholeProviderConfig() string {
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
}
`, url, password)
}

func TestAccPiholeProviderWithConfiguration(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPiholeProviderConfigWithSettings(),
				Check:  resource.ComposeAggregateTestCheckFunc(
				// Verify provider accepts custom configuration
				),
			},
		},
	})
}

func testAccPiholeProviderConfigWithSettings() string {
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
  url                   = %[1]q
  password              = %[2]q
  max_connections       = 2
  request_delay_ms      = 100
  retry_attempts        = 5
  retry_backoff_base_ms = 250
}
`, url, password)
}

// Unit tests for provider configuration
func TestPiholeProvider_Schema(t *testing.T) {
	ctx := context.Background()
	piholeProvider := &PiholeProvider{}

	req := provider.SchemaRequest{}
	resp := &provider.SchemaResponse{}

	piholeProvider.Schema(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Provider schema has errors: %v", resp.Diagnostics.Errors())
	}

	if resp.Schema.Attributes == nil {
		t.Fatal("Provider schema should have attributes")
	}

	// Check required attributes
	if _, exists := resp.Schema.Attributes["url"]; !exists {
		t.Error("Provider schema should have 'url' attribute")
	}

	if _, exists := resp.Schema.Attributes["password"]; !exists {
		t.Error("Provider schema should have 'password' attribute")
	}

	// Check optional attributes
	if _, exists := resp.Schema.Attributes["max_connections"]; !exists {
		t.Error("Provider schema should have 'max_connections' attribute")
	}

	if _, exists := resp.Schema.Attributes["request_delay_ms"]; !exists {
		t.Error("Provider schema should have 'request_delay_ms' attribute")
	}

	if _, exists := resp.Schema.Attributes["retry_attempts"]; !exists {
		t.Error("Provider schema should have 'retry_attempts' attribute")
	}

	if _, exists := resp.Schema.Attributes["retry_backoff_base_ms"]; !exists {
		t.Error("Provider schema should have 'retry_backoff_base_ms' attribute")
	}
}

func TestPiholeProvider_Metadata(t *testing.T) {
	ctx := context.Background()
	piholeProvider := &PiholeProvider{version: "test-version"}

	req := provider.MetadataRequest{}
	resp := &provider.MetadataResponse{}

	piholeProvider.Metadata(ctx, req, resp)

	if resp.TypeName != "pihole" {
		t.Errorf("Expected provider type name to be 'pihole', got '%s'", resp.TypeName)
	}

	if resp.Version != "test-version" {
		t.Errorf("Expected provider version to be 'test-version', got '%s'", resp.Version)
	}
}

func TestPiholeProvider_Resources(t *testing.T) {
	ctx := context.Background()
	provider := &PiholeProvider{}

	resources := provider.Resources(ctx)

	if len(resources) != 2 {
		t.Errorf("Expected 2 resources, got %d", len(resources))
	}

	// Test that resource functions can be called without panic
	for i, resourceFunc := range resources {
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Resource function %d panicked: %v", i, r)
				}
			}()
			resourceFunc()
		}()
	}
}

func TestPiholeProvider_DataSources(t *testing.T) {
	ctx := context.Background()
	provider := &PiholeProvider{}

	dataSources := provider.DataSources(ctx)

	// Currently no data sources, but should return empty slice
	if len(dataSources) != 0 {
		t.Errorf("Expected 0 data sources, got %d", len(dataSources))
	}
}

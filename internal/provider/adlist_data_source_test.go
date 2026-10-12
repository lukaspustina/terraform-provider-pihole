package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccPiholeAdlistDataSource_basic(t *testing.T) {
	testAccPreCheck(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPiholeAdlistDataSourceConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.pihole_adlist.test", "id"),
					resource.TestCheckResourceAttr("data.pihole_adlist.test", "address", "https://raw.githubusercontent.com/StevenBlack/hosts/master/hosts"),
					resource.TestCheckResourceAttr("data.pihole_adlist.test", "type", "block"),
					resource.TestCheckResourceAttrSet("data.pihole_adlist.test", "enabled"),
				),
			},
		},
	})
}

func TestAccPiholeAdlistDataSource_notFound(t *testing.T) {
	testAccPreCheck(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccPiholeAdlistDataSourceConfig_notFound(),
				ExpectError: regexp.MustCompile("Adlist Not Found"),
			},
		},
	})
}

func TestPiholeAdlistDataSource_Schema(t *testing.T) {
	ctx := testContext()
	req := testDataSourceSchemaRequest()
	resp := &testDataSourceSchemaResponse{}

	ds := NewAdlistDataSource()
	ds.Schema(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema has errors: %v", resp.Diagnostics)
	}

	for _, name := range []string{"id", "address", "enabled", "comment", "type", "groups"} {
		if resp.Schema.Attributes[name] == nil {
			t.Errorf("Expected %q attribute in schema", name)
		}
	}

	if !resp.Schema.Attributes["address"].IsRequired() {
		t.Error("Expected 'address' attribute to be required")
	}
	if !resp.Schema.Attributes["id"].IsComputed() {
		t.Error("Expected 'id' attribute to be computed")
	}
}

func TestFindAdlistByAddress_Found(t *testing.T) {
	adlists := []Adlist{
		{ID: 1, Address: "https://example.com/a.txt"},
		{ID: 2, Address: "https://example.com/b.txt"},
	}

	result := findAdlistByAddress(adlists, "https://example.com/b.txt")
	if result == nil {
		t.Fatal("Expected to find adlist, got nil")
	}
	if result.ID != 2 {
		t.Errorf("Expected ID 2, got %d", result.ID)
	}
}

func TestFindAdlistByAddress_NotFound(t *testing.T) {
	adlists := []Adlist{
		{ID: 1, Address: "https://example.com/a.txt"},
	}

	result := findAdlistByAddress(adlists, "https://example.com/missing.txt")
	if result != nil {
		t.Errorf("Expected nil for unknown address, got: %+v", result)
	}
}

func TestFindAdlistByAddress_Empty(t *testing.T) {
	result := findAdlistByAddress([]Adlist{}, "https://example.com/list.txt")
	if result != nil {
		t.Errorf("Expected nil for empty slice, got: %+v", result)
	}
}

func testAccPiholeAdlistDataSourceConfig_basic() string {
	return fmt.Sprintf(`
%s

resource "pihole_adlist" "seed" {
  address = "https://raw.githubusercontent.com/StevenBlack/hosts/master/hosts"
  comment = "lookup test"
}

data "pihole_adlist" "test" {
  address    = pihole_adlist.seed.address
  depends_on = [pihole_adlist.seed]
}
`, testAccPiholeProviderBlock())
}

func testAccPiholeAdlistDataSourceConfig_notFound() string {
	return fmt.Sprintf(`
%s

data "pihole_adlist" "test" {
  address = "https://does-not-exist.example.com/list.txt"
}
`, testAccPiholeProviderBlock())
}

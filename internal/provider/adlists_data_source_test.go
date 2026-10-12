package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccPiholeAdlistsDataSource_basic(t *testing.T) {
	testAccPreCheck(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPiholeAdlistsDataSourceConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.pihole_adlists.test", "id", "adlists"),
					resource.TestMatchResourceAttr("data.pihole_adlists.test", "adlists.#", regexp.MustCompile(`^\d+$`)),
				),
			},
		},
	})
}

func TestAccPiholeAdlistsDataSource_withExisting(t *testing.T) {
	testAccPreCheck(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPiholeAdlistsDataSourceConfig_withAdlist(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("pihole_adlist.test", "id"),
					resource.TestMatchResourceAttr("data.pihole_adlists.all", "adlists.#", regexp.MustCompile(`^[1-9]\d*$`)),
				),
			},
		},
	})
}

func TestPiholeAdlistsDataSource_Schema(t *testing.T) {
	ctx := testContext()
	req := testDataSourceSchemaRequest()
	resp := &testDataSourceSchemaResponse{}

	ds := NewAdlistsDataSource()
	ds.Schema(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema has errors: %v", resp.Diagnostics)
	}

	if resp.Schema.Attributes["id"] == nil {
		t.Error("Expected 'id' attribute in schema")
	}
	if resp.Schema.Attributes["adlists"] == nil {
		t.Error("Expected 'adlists' attribute in schema")
	}
	if !resp.Schema.Attributes["adlists"].IsComputed() {
		t.Error("Expected 'adlists' attribute to be computed")
	}
}

func TestAdlistsToDataSourceModels(t *testing.T) {
	adlists := []Adlist{
		{ID: 1, Address: "https://example.com/a.txt", Enabled: true, Comment: "first", Type: "block", Groups: []int{0}},
		{ID: 2, Address: "https://example.com/b.txt", Enabled: false, Comment: "second", Type: "allow", Groups: []int{0, 1}},
	}

	models, diags := adlistsToDataSourceModels(adlists)
	if diags.HasError() {
		t.Fatalf("adlistsToDataSourceModels returned errors: %v", diags)
	}
	if len(models) != 2 {
		t.Fatalf("Expected 2 models, got %d", len(models))
	}

	if models[0].ID.ValueString() != "1" {
		t.Errorf("Expected ID '1', got %q", models[0].ID.ValueString())
	}
	if models[0].Address.ValueString() != "https://example.com/a.txt" {
		t.Errorf("Unexpected address: %q", models[0].Address.ValueString())
	}
	if models[0].Enabled.ValueBool() != true {
		t.Error("Expected enabled=true for first model")
	}
	if models[1].Type.ValueString() != "allow" {
		t.Errorf("Expected type 'allow', got %q", models[1].Type.ValueString())
	}
	if len(models[1].Groups.Elements()) != 2 {
		t.Errorf("Expected 2 groups for second model, got %d", len(models[1].Groups.Elements()))
	}
}

func TestAdlistsToDataSourceModels_Empty(t *testing.T) {
	models, diags := adlistsToDataSourceModels([]Adlist{})
	if diags.HasError() {
		t.Fatalf("adlistsToDataSourceModels returned errors for empty input: %v", diags)
	}
	if len(models) != 0 {
		t.Errorf("Expected 0 models, got %d", len(models))
	}
}

func testAccPiholeAdlistsDataSourceConfig_basic() string {
	return fmt.Sprintf(`
%s

data "pihole_adlists" "test" {}
`, testAccPiholeProviderBlock())
}

func testAccPiholeAdlistsDataSourceConfig_withAdlist() string {
	return fmt.Sprintf(`
%s

resource "pihole_adlist" "test" {
  address = "https://raw.githubusercontent.com/StevenBlack/hosts/master/hosts"
  comment = "data source test"
}

data "pihole_adlists" "all" {
  depends_on = [pihole_adlist.test]
}
`, testAccPiholeProviderBlock())
}

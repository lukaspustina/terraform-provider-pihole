package provider

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccPiholeAdlist_basic(t *testing.T) {
	testAccPreCheck(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPiholeAdlistConfig("https://raw.githubusercontent.com/StevenBlack/hosts/master/hosts", "test adlist", "block", true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("pihole_adlist.test", "id"),
					resource.TestCheckResourceAttr("pihole_adlist.test", "address", "https://raw.githubusercontent.com/StevenBlack/hosts/master/hosts"),
					resource.TestCheckResourceAttr("pihole_adlist.test", "comment", "test adlist"),
					resource.TestCheckResourceAttr("pihole_adlist.test", "type", "block"),
					resource.TestCheckResourceAttr("pihole_adlist.test", "enabled", "true"),
				),
			},
			{
				ResourceName:      "pihole_adlist.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPiholeAdlistConfig("https://raw.githubusercontent.com/StevenBlack/hosts/master/hosts", "updated comment", "block", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pihole_adlist.test", "comment", "updated comment"),
					resource.TestCheckResourceAttr("pihole_adlist.test", "enabled", "false"),
				),
			},
		},
	})
}

func TestAccPiholeAdlist_allowList(t *testing.T) {
	testAccPreCheck(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPiholeAdlistConfig("https://raw.githubusercontent.com/nickcoutsos/dnsmasq-allowlist/master/allowlist.txt", "allow list", "allow", true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("pihole_adlist.test", "id"),
					resource.TestCheckResourceAttr("pihole_adlist.test", "type", "allow"),
				),
			},
		},
	})
}

func TestAccPiholeAdlist_disappears(t *testing.T) {
	testAccPreCheck(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPiholeAdlistConfig("https://raw.githubusercontent.com/StevenBlack/hosts/master/hosts", "disappears test", "block", true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPiholeAdlistExists("pihole_adlist.test"),
					testAccCheckPiholeAdlistDestroy("pihole_adlist.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccPiholeAdlistConfig(address, comment, listType string, enabled bool) string {
	return fmt.Sprintf(`
%s

resource "pihole_adlist" "test" {
  address = %[2]q
  comment = %[3]q
  type    = %[4]q
  enabled = %[5]t
}
`, testAccPiholeProviderBlock(), address, comment, listType, enabled)
}

func testAccCheckPiholeAdlistExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("adlist not found: %s", resourceName)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("adlist ID is not set")
		}
		return nil
	}
}

func testAccCheckPiholeAdlistDestroy(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("resource ID not set")
		}

		url := os.Getenv("PIHOLE_URL")
		password := os.Getenv("PIHOLE_PASSWORD")
		if url == "" || password == "" {
			return fmt.Errorf("PIHOLE_URL and PIHOLE_PASSWORD must be set for disappears test")
		}

		client, err := getOrCreateClient(url, password, ClientConfig{
			MaxConnections: 1,
			RequestDelayMs: 300,
			RetryAttempts:  3,
			RetryBackoffMs: 500,
		})
		if err != nil {
			return fmt.Errorf("failed to create client: %v", err)
		}

		id := 0
		if _, err := fmt.Sscanf(rs.Primary.ID, "%d", &id); err != nil {
			return fmt.Errorf("failed to parse adlist ID: %v", err)
		}

		return client.DeleteAdlist(id)
	}
}

// Unit tests

func TestAdlistResource_Schema(t *testing.T) {
	r := NewAdlistResource()

	schemaReq := fwresource.SchemaRequest{}
	schemaResp := &fwresource.SchemaResponse{}

	r.Schema(context.Background(), schemaReq, schemaResp)

	if schemaResp.Diagnostics.HasError() {
		t.Fatalf("Schema has errors: %v", schemaResp.Diagnostics.Errors())
	}

	attrs := schemaResp.Schema.Attributes

	for _, name := range []string{"id", "address", "enabled", "comment", "type", "groups"} {
		if attrs[name] == nil {
			t.Errorf("Schema should have %q attribute", name)
		}
	}

	if !attrs["address"].IsRequired() {
		t.Error("'address' attribute should be required")
	}
	if !attrs["id"].IsComputed() {
		t.Error("'id' attribute should be computed")
	}
	if !attrs["enabled"].IsOptional() {
		t.Error("'enabled' attribute should be optional")
	}
}

func TestAdlistResource_Metadata(t *testing.T) {
	r := NewAdlistResource()

	req := fwresource.MetadataRequest{ProviderTypeName: "pihole"}
	resp := &fwresource.MetadataResponse{}

	r.Metadata(context.Background(), req, resp)

	if resp.TypeName != "pihole_adlist" {
		t.Errorf("Expected TypeName 'pihole_adlist', got %q", resp.TypeName)
	}
}

func TestIntsToList(t *testing.T) {
	list, diags := intsToList([]int{0, 1, 2})
	if diags.HasError() {
		t.Fatalf("intsToList returned errors: %v", diags)
	}
	if len(list.Elements()) != 3 {
		t.Errorf("Expected 3 elements, got %d", len(list.Elements()))
	}
}

func TestIntsToList_Empty(t *testing.T) {
	list, diags := intsToList([]int{})
	if diags.HasError() {
		t.Fatalf("intsToList returned errors for empty slice: %v", diags)
	}
	if len(list.Elements()) != 0 {
		t.Errorf("Expected 0 elements, got %d", len(list.Elements()))
	}
	if list.IsNull() {
		t.Error("Expected non-null empty list, got null")
	}
}

func TestListToInts(t *testing.T) {
	list, _ := intsToList([]int{0, 1, 2})
	result := listToInts(list)

	if len(result) != 3 {
		t.Fatalf("Expected 3 ints, got %d", len(result))
	}
	for i, v := range []int{0, 1, 2} {
		if result[i] != v {
			t.Errorf("Index %d: expected %d, got %d", i, v, result[i])
		}
	}
}

func TestAdlistToModel(t *testing.T) {
	adlist := &Adlist{
		ID:      42,
		Address: "https://example.com/list.txt",
		Enabled: true,
		Comment: "test comment",
		Type:    "block",
		Groups:  []int{0, 1},
	}

	var data AdlistResourceModel
	var diags diag.Diagnostics
	adlistToModel(adlist, &data, &diags)

	if diags.HasError() {
		t.Fatalf("adlistToModel returned errors: %v", diags)
	}
	if data.ID.ValueString() != "42" {
		t.Errorf("Expected ID '42', got %q", data.ID.ValueString())
	}
	if data.Address.ValueString() != adlist.Address {
		t.Errorf("Expected address %q, got %q", adlist.Address, data.Address.ValueString())
	}
	if data.Enabled.ValueBool() != adlist.Enabled {
		t.Errorf("Expected enabled %v, got %v", adlist.Enabled, data.Enabled.ValueBool())
	}
	if data.Comment.ValueString() != adlist.Comment {
		t.Errorf("Expected comment %q, got %q", adlist.Comment, data.Comment.ValueString())
	}
	if data.Type.ValueString() != adlist.Type {
		t.Errorf("Expected type %q, got %q", adlist.Type, data.Type.ValueString())
	}
	if len(data.Groups.Elements()) != 2 {
		t.Errorf("Expected 2 group elements, got %d", len(data.Groups.Elements()))
	}
}

func TestAdlistToModel_EmptyGroups(t *testing.T) {
	adlist := &Adlist{
		ID:      1,
		Address: "https://example.com/list.txt",
		Groups:  []int{},
	}

	var data AdlistResourceModel
	var diags diag.Diagnostics
	adlistToModel(adlist, &data, &diags)

	if diags.HasError() {
		t.Fatalf("adlistToModel returned errors for empty groups: %v", diags)
	}
	if len(data.Groups.Elements()) != 0 {
		t.Errorf("Expected 0 group elements, got %d", len(data.Groups.Elements()))
	}
}

func TestAdlistResource_ImportState_InvalidID(t *testing.T) {
	r := NewAdlistResource().(fwresource.ResourceWithImportState)

	req := fwresource.ImportStateRequest{ID: "not-an-integer"}
	resp := &fwresource.ImportStateResponse{}

	r.ImportState(context.Background(), req, resp)

	if !resp.Diagnostics.HasError() {
		t.Error("Expected diagnostic error for non-integer import ID, got none")
	}
}

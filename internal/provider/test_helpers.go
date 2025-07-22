package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

// Test helper functions for unit tests

// testContext returns a basic context for testing
func testContext() context.Context {
	return context.Background()
}

// testDataSourceSchemaRequest returns a basic schema request for testing
func testDataSourceSchemaRequest() datasource.SchemaRequest {
	return datasource.SchemaRequest{}
}

// testDataSourceSchemaResponse represents a schema response for testing
type testDataSourceSchemaResponse = datasource.SchemaResponse

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccS3SourceResource(t *testing.T) {
	t.Parallel()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSourceResourceConfig(
					"acc", "test", "created by acceptance test", "SELECT * FROM _input", 3600),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"rockset_collection.test", "id", "acc.test"),
					resource.TestCheckResourceAttr(
						"rockset_collection.test", "name", "test"),
					resource.TestCheckResourceAttr(
						"rockset_collection.test", "workspace", "acc"),
					resource.TestCheckResourceAttr(
						"rockset_collection.test", "description", "created by acceptance test"),
					resource.TestCheckResourceAttr(
						"rockset_collection.test", "retention_secs", "3600"),
					resource.TestCheckResourceAttr(
						"rockset_collection.test", "storage_compression_type", "LZ4"),
					resource.TestCheckResourceAttr(
						"rockset_collection.test", "ingest_transformation", "SELECT * FROM _input"),
					// read only fields below
					resource.TestCheckResourceAttrSet(
						"rockset_collection.test", "created_at"),
					resource.TestCheckResourceAttrSet(
						"rockset_collection.test", "created_by"),
					resource.TestCheckResourceAttrSet(
						"rockset_collection.test", "read_only"),
					resource.TestCheckResourceAttrSet(
						"rockset_collection.test", "insert_only"),
					resource.TestCheckResourceAttrSet(
						"rockset_collection.test", "status"),
					resource.TestCheckResourceAttrSet(
						"rockset_collection.test", "created_by_apikey_name"),
				),
			},
			{
				Config: testAccCollectionResourceConfig(
					"acc", "test", "updated by acceptance test", "SELECT *, 'foo' AS foo FROM _input", 3600),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"rockset_collection.test", "description", "updated by acceptance test"),
				),
			},
		},
	})
}

func testAccSourceResourceConfig(workspace, name, desc, transform string, retention int) string {
	return fmt.Sprintf(`resource "rockset_collection" "test" {
	workspace = "%s"
	name = "%s"
	description = "%s"
	retention_secs = %d
	storage_compression_type = "LZ4"
	ingest_transformation = "%s"
}`, workspace, name, desc, retention, transform)
}

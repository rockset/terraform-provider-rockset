package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccExampleDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccExampleDataSourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.rockset_collection_source.test", "id", "eafa74f8-d59d-4d7e-9f7e-efa50810d8b8"),
					resource.TestCheckResourceAttr("data.rockset_collection_source.test", "collection", "snp"),
					resource.TestCheckResourceAttr("data.rockset_collection_source.test", "workspace", "persistent"),
					resource.TestCheckResourceAttr("data.rockset_collection_source.test", "resume_at", ""),
					resource.TestCheckResourceAttr("data.rockset_collection_source.test", "suspended_at", ""),
					// TODO setup a source which uses an integration, as the snp collection uses the public dataset,
					//  which doesn't require an integration
					resource.TestCheckResourceAttr("data.rockset_collection_source.test", "integration_name", ""),
					resource.TestCheckResourceAttr("data.rockset_collection_source.test", "status.state", "WATCHING"),
					resource.TestCheckResourceAttr("data.rockset_collection_source.test", "status.message", ""),
					resource.TestCheckResourceAttrSet("data.rockset_collection_source.test", "status.detected_size_bytes"),
					resource.TestCheckResourceAttrSet("data.rockset_collection_source.test", "status.last_processed_at"),
					resource.TestCheckResourceAttrSet("data.rockset_collection_source.test", "status.last_processed_item"),
					resource.TestCheckResourceAttrSet("data.rockset_collection_source.test", "status.total_processed_items"),
				),
			},
		},
	})
}

const testAccExampleDataSourceConfig = `
data "rockset_collection_source" "test" {
	workspace = "persistent"
	collection = "snp"
	id = "eafa74f8-d59d-4d7e-9f7e-efa50810d8b8"
}`

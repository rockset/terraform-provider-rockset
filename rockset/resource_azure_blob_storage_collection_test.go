package rockset

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/rockset/rockset-go-client/openapi"
)

func TestAccAzureBlobStorageCollection_Basic(t *testing.T) {
	var collection openapi.Collection

	name := randomName("azure_blob_storage")
	values := Values{
		Name:        name,
		Collection:  name,
		Workspace:   name,
		Description: description(),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRocksetCollectionDestroy, // Reused from base collection
		Steps: []resource.TestStep{
			{
				Config: getHCLTemplate("azure_blob_storage_collection.tf", values),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRocksetCollectionExists("rockset_azure_blob_storage_collection.test", &collection), // Reused from base collection
					resource.TestCheckResourceAttr("rockset_azure_blob_storage_collection.test", "name", values.Collection),
					resource.TestCheckResourceAttr("rockset_azure_blob_storage_collection.test", "workspace", values.Workspace),
					resource.TestCheckResourceAttr("rockset_azure_blob_storage_collection.test", "description", values.Description),
					resource.TestCheckResourceAttr("rockset_azure_blob_storage_collection.test", "retention_secs", "3600"),
				),
			},
		},
	})
}

func TestAccAzureBlobStorageCollection_Json(t *testing.T) {
	var collection openapi.Collection

	name := randomName("azure_blob_storage-json")
	values := Values{
		Name:        name,
		Collection:  name,
		Workspace:   name,
		Description: description(),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRocksetCollectionDestroy, // Reused from base collection
		Steps: []resource.TestStep{
			{
				Config: getHCLTemplate("azure_blob_storage_collection_json.tf", values),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRocksetCollectionExists("rockset_azure_blob_storage_collection.test", &collection), // Reused from base collection
					resource.TestCheckResourceAttr("rockset_azure_blob_storage_collection.test", "name", values.Name),
					resource.TestCheckResourceAttr("rockset_azure_blob_storage_collection.test", "workspace", values.Workspace),
					resource.TestCheckResourceAttr("rockset_azure_blob_storage_collection.test", "description", values.Description),
					resource.TestCheckResourceAttr("rockset_azure_blob_storage_collection.test", "retention_secs", "3600"),
					resource.TestCheckResourceAttr("rockset_azure_blob_storage_collection.test", "source.0.integration_name", values.Name),
					resource.TestCheckResourceAttr("rockset_azure_blob_storage_collection.test", "source.0.container", "sampledatasets"),
					resource.TestCheckResourceAttr("rockset_azure_blob_storage_collection.test", "source.0.pattern", "product.json"),
					resource.TestCheckResourceAttr("rockset_azure_blob_storage_collection.test", "source.0.format", "json"),
				),
			},
			{
				PreConfig: func() {
					triggerWriteAPISourceAdd(t, values.Workspace, values.Collection)
				},
				Config: getHCLTemplate("azure_blob_storage_collection.tf", values),
				Check: resource.ComposeTestCheckFunc(
					// check that we still just have two sources
					resource.TestCheckResourceAttr("rockset_azure_blob_storage_collection.test", "source.#", "2"),
				),
			},
		},
	})
}

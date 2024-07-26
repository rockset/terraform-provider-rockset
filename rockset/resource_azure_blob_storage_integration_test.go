package rockset

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/rockset/rockset-go-client"
	"github.com/rockset/rockset-go-client/openapi"
)

const ConnectionString = "BlobEndpoint=https://a.blob.core.windows.net/;SharedAccessSignature=sv=2022-11-02&ss=x=co&sp=x=2024-07-28T02:59:32Z&st=2024-07-26T18:59:32Z&spr=https&sig=x"

func TestAccAzureBlobStorageIntegration_Basic(t *testing.T) {
	var azureBlobStorageIntegration openapi.AzureBlobStorageIntegration

	name := randomName("azure_blob_storage")
	values := Values{
		Name: name,
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRocksetIntegrationDestroy("rockset_azure_blob_storage_integration"),
		Steps: []resource.TestStep{
			{
				Config: getHCLTemplate("azure_blob_storage_integration.tf", values),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRocksetAzureBlobStorageIntegrationExists("rockset_azure_blob_storage_integration.test", &azureBlobStorageIntegration),
					resource.TestCheckResourceAttr("rockset_azure_blob_storage_integration.test", "name", values.Name),
					resource.TestCheckResourceAttr("rockset_azure_blob_storage_integration.test", "connection_string", ConnectionString),
				),
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func testAccCheckRocksetAzureBlobStorageIntegrationExists(resource string,
	azureBlobStorageIntegration *openapi.AzureBlobStorageIntegration) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rc := testAccProvider.Meta().(*rockset.RockClient)

		rs, err := getResourceFromState(state, resource)
		if err != nil {
			return err
		}

		name := rs.Primary.ID
		resp, err := rc.GetIntegration(testCtx, name)
		if err != nil {
			return err
		}

		*azureBlobStorageIntegration = *resp.AzureBlobStorage

		return nil
	}
}

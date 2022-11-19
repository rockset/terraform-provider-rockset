package rockset

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/rockset/rockset-go-client"
	"github.com/rockset/rockset-go-client/openapi"
)

const testGCSIntegrationName = "terraform-provider-acceptance-test-gcs-integration"
const testGCSIntegrationDescription = "Terraform provider acceptance tests."

func TestAccGCSIntegration_Basic(t *testing.T) {
	var gcsIntegration openapi.GcsIntegration

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t, "TF_VAR_GCS_SERVICE_ACCOUNT_KEY") },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRocksetIntegrationDestroy("rockset_gcs_integration"),
		Steps: []resource.TestStep{
			{
				Config: getHCL("gcs_integration.tf"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRocksetGCSIntegrationExists("rockset_gcs_integration.test", &gcsIntegration),
					resource.TestCheckResourceAttr("rockset_gcs_integration.test",
						"name", testGCSIntegrationName),
					resource.TestCheckResourceAttr("rockset_gcs_integration.test",
						"description", testGCSIntegrationDescription),
				),
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func testAccCheckRocksetGCSIntegrationExists(resource string,
	gcsIntegration *openapi.GcsIntegration) resource.TestCheckFunc {
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

		*gcsIntegration = *resp.Gcs

		return nil
	}
}

package rockset

import (
	"errors"
	"os"
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
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckGCS(t)
		},
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRocksetGCSIntegrationDestroy,
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

func testAccPreCheckGCS(t *testing.T) {
	if v := os.Getenv("TF_VAR_GCS_SERVICE_ACCOUNT_KEY"); v == "" {
		t.Fatal("TF_VAR_GCS_SERVICE_ACCOUNT_KEY must be set for GCS acceptance tests")
	}
}

func testAccCheckRocksetGCSIntegrationDestroy(s *terraform.State) error {
	rc := testAccProvider.Meta().(*rockset.RockClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "rockset_gcs_integration" {
			continue
		}

		name := rs.Primary.ID
		_, err := rc.GetIntegration(testCtx, name)
		// An error could mean we didn't find the it, which is what we expect
		if err != nil {
			var re rockset.Error
			if errors.As(err, &re) {
				if re.IsNotFoundError() {
					return nil
				}
			}
			return err
		}
	}

	return nil
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

package rockset

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/rockset/rockset-go-client"
	"github.com/rockset/rockset-go-client/openapi"
)

const testS3IntegrationName = "terraform-provider-acceptance-test-s3-integration"
const testS3IntegrationDescription = "Terraform provider acceptance tests."
const testS3IntegrationRoleArn = "arn:aws:iam::469279130686:role/terraform-provider-rockset-tests"

func TestAccS3Integration_Basic(t *testing.T) {
	var s3Integration openapi.S3Integration

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRocksetS3IntegrationDestroy,
		Steps: []resource.TestStep{
			{
				Config: getHCL("s3_integration.tf"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRocksetS3IntegrationExists("rockset_s3_integration.test", &s3Integration),
					resource.TestCheckResourceAttr("rockset_s3_integration.test", "name", testS3IntegrationName),
					resource.TestCheckResourceAttr("rockset_s3_integration.test", "description", testS3IntegrationDescription),
					resource.TestCheckResourceAttr("rockset_s3_integration.test", "aws_role_arn", testS3IntegrationRoleArn),
				),
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func testAccCheckRocksetS3IntegrationDestroy(s *terraform.State) error {
	rc := testAccProvider.Meta().(*rockset.RockClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "rockset_s3_integration" {
			continue
		}

		name := rs.Primary.ID
		_, err := rc.GetIntegration(testCtx, name)
		// An error would mean we didn't find the it, we expect an error
		if err == nil {
			return err
		}
	}

	return nil
}

func testAccCheckRocksetS3IntegrationExists(resource string,
	s3Integration *openapi.S3Integration) resource.TestCheckFunc {
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

		*s3Integration = *resp.S3

		return nil
	}
}

package rockset

import (
	"fmt"
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
		PreCheck:     func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy: testAccCheckRocksetS3IntegrationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckS3IntegrationBasic(),
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

func testAccCheckS3IntegrationBasic() string {
	return fmt.Sprintf(`
resource rockset_s3_integration test {
	name = "%s"
	description = "%s"
	aws_role_arn = "%s"
}
`, testS3IntegrationName, testS3IntegrationDescription, testS3IntegrationRoleArn)
}

func testAccCheckRocksetS3IntegrationDestroy(s *terraform.State) error {
	rc := testAccProvider.Meta().(*rockset.RockClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "rockset_s3_integration" {
			continue
		}

		name := rs.Primary.ID
		// TODO: Change to convenience method
		getReq := rc.IntegrationsApi.GetIntegration(testCtx, name)
		_, _, err := getReq.Execute()
		// An error would mean we didn't find the it, we expect an error
		if err == nil {
			return err
		}
	}

	return nil
}

func testAccCheckRocksetS3IntegrationExists(resource string, s3Integration *openapi.S3Integration) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rc := testAccProvider.Meta().(*rockset.RockClient)

		rs, err := getResourceFromState(state, resource)
		if err != nil {
			return err
		}

		name := rs.Primary.ID
		// TODO: Change to convenience method
		getReq := rc.IntegrationsApi.GetIntegration(testCtx, name)
		resp, _, err := getReq.Execute()
		if err != nil {
			return err
		}

		*s3Integration = *resp.Data.S3

		return nil
	}
}

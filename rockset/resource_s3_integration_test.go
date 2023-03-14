package rockset

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/rockset/rockset-go-client"
	"github.com/rockset/rockset-go-client/openapi"
)

func TestAccS3Integration_Basic(t *testing.T) {
	var s3Integration openapi.S3Integration

	name := randomName("s3")
	values := Values{
		Name:        name,
		Collection:  name,
		Workspace:   name,
		Description: description(),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRocksetIntegrationDestroy("rockset_s3_integration"),
		Steps: []resource.TestStep{
			{
				Config: getHCLTemplate("s3_integration.tf", values),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRocksetS3IntegrationExists("rockset_s3_integration.test", &s3Integration),
					resource.TestCheckResourceAttr("rockset_s3_integration.test", "name", values.Name),
					resource.TestCheckResourceAttr("rockset_s3_integration.test", "description", values.Description),
					resource.TestCheckResourceAttr("rockset_s3_integration.test", "aws_role_arn", S3IntegrationRoleArn),
				),
				ExpectNonEmptyPlan: false,
			},
		},
	})
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

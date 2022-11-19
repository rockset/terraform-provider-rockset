package rockset

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/rockset/rockset-go-client"
	"github.com/rockset/rockset-go-client/openapi"
)

func TestAccDynamoDBIntegration_Basic(t *testing.T) {
	var dynamoDBIntegration openapi.DynamodbIntegration

	values := Values{
		Name:        randomName("integration"),
		Description: description(),
		Workspace:   "acc",
		Role:        "arn:aws:iam::469279130686:role/terraform-provider-rockset-tests-dynamo",
		Bucket:      "terraform-provider-rockset-tests",
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRocksetIntegrationDestroy("rockset_dynamodb_integration"),
		Steps: []resource.TestStep{
			{
				Config: getHCLTemplate("dynamodb_integration.tf", values),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRocksetDynamoDBIntegrationExists("rockset_dynamodb_integration.test", &dynamoDBIntegration),
					resource.TestCheckResourceAttr("rockset_dynamodb_integration.test", "name", values.Name),
					resource.TestCheckResourceAttr("rockset_dynamodb_integration.test", "description", values.Description),
					resource.TestCheckResourceAttr("rockset_dynamodb_integration.test", "aws_role_arn", values.Role),
					resource.TestCheckResourceAttr("rockset_dynamodb_integration.test", "s3_export_bucket_name", values.Bucket),
				),
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func testAccCheckRocksetDynamoDBIntegrationExists(resource string, dynamoDBIntegration *openapi.DynamodbIntegration) resource.TestCheckFunc {
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

		*dynamoDBIntegration = *resp.Dynamodb

		return nil
	}
}

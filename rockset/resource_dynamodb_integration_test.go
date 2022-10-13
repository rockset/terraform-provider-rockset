package rockset

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/rockset/rockset-go-client"
	"github.com/rockset/rockset-go-client/openapi"
)

const testDynamoDBIntegrationName = "terraform-provider-acceptance-test-dynamodb-integration"
const testDynamoDBIntegrationDescription = "Terraform provider acceptance tests."
const testDynamoDBIntegrationRoleArn = "arn:aws:iam::469279130686:role/terraform-provider-rockset-tests-dynamo"
const testDynamoDBIntegrationS3Bucket = "terraform-provider-rockset-tests"

func TestAccDynamoDBIntegration_Basic(t *testing.T) {
	var dynamoDBIntegration openapi.DynamodbIntegration

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRocksetIntegrationDestroy("rockset_dynamodb_integration"),
		Steps: []resource.TestStep{
			{
				Config: getHCL("dynamodb_integration.tf"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRocksetDynamoDBIntegrationExists("rockset_dynamodb_integration.test",
						&dynamoDBIntegration),
					resource.TestCheckResourceAttr("rockset_dynamodb_integration.test", "name",
						testDynamoDBIntegrationName),
					resource.TestCheckResourceAttr("rockset_dynamodb_integration.test", "description",
						testDynamoDBIntegrationDescription),
					resource.TestCheckResourceAttr("rockset_dynamodb_integration.test", "aws_role_arn",
						testDynamoDBIntegrationRoleArn),
					resource.TestCheckResourceAttr("rockset_dynamodb_integration.test",
						"s3_export_bucket_name", testDynamoDBIntegrationS3Bucket),
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

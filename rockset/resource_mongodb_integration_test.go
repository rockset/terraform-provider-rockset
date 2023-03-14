package rockset

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/rockset/rockset-go-client"
	"github.com/rockset/rockset-go-client/openapi"
)

const testMongoDBIntegrationName = "terraform-provider-acceptance-test-mongodb-integration"
const testMongoDBIntegrationDescription = "Terraform provider acceptance tests."

func TestAccMongoDBIntegration_Basic(t *testing.T) {
	t.Skip("mongodb needs to be reconfigured")
	var mongoDBIntegration openapi.MongoDbIntegration

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t, "TF_VAR_MONGODB_CONNECTION_URI") },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRocksetIntegrationDestroy("rockset_mongodb_integration"),
		Steps: []resource.TestStep{
			{
				Config: getHCL("mongodb_integration.tf"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRocksetMongoDBIntegrationExists("rockset_mongodb_integration.test",
						&mongoDBIntegration),
					resource.TestCheckResourceAttr("rockset_mongodb_integration.test", "name",
						testMongoDBIntegrationName),
					resource.TestCheckResourceAttr("rockset_mongodb_integration.test", "description",
						testMongoDBIntegrationDescription),
				),
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func testAccCheckRocksetMongoDBIntegrationExists(resource string, mongoDBIntegration *openapi.MongoDbIntegration) resource.TestCheckFunc {
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

		*mongoDBIntegration = *resp.Mongodb

		return nil
	}
}

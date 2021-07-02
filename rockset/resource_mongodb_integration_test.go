package rockset

import (
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/rockset/rockset-go-client"
	"github.com/rockset/rockset-go-client/openapi"
)

const testMongoDBIntegrationName = "terraform-provider-acceptance-test-mongodb-integration"
const testMongoDBIntegrationDescription = "Terraform provider acceptance tests."

/*
	Verifies necessary environment variables are set before running tests.
	Fails early if they are not set.
*/
func testAccPreCheckMongo(t *testing.T) {
	if v := os.Getenv("TF_VAR_MONGODB_CONNECTION_URI"); v == "" {
		t.Fatal("TF_VAR_MONGODB_CONNECTION_URI must be set for MongoDB acceptance tests")
	}
}

func TestAccMongoDBIntegration_Basic(t *testing.T) {
	var mongoDBIntegration openapi.MongoDbIntegration

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckMongo(t)
		},
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRocksetMongoDBIntegrationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckMongoDBIntegrationBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRocksetMongoDBIntegrationExists("rockset_mongodb_integration.test", &mongoDBIntegration),
					resource.TestCheckResourceAttr("rockset_mongodb_integration.test", "name", testMongoDBIntegrationName),
					resource.TestCheckResourceAttr("rockset_mongodb_integration.test", "description", testMongoDBIntegrationDescription),
				),
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func testAccCheckMongoDBIntegrationBasic() string {
	hclPath := filepath.Join("..", "testdata", "mongodb_integration.tf")
	hcl, err := getFileContents(hclPath)
	if err != nil {
		log.Fatalf("Unexpected error loading test data %s", hclPath)
	}

	return hcl
}

func testAccCheckRocksetMongoDBIntegrationDestroy(s *terraform.State) error {
	rc := testAccProvider.Meta().(*rockset.RockClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "rockset_mongodb_integration" {
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

func testAccCheckRocksetMongoDBIntegrationExists(resource string, mongoDBIntegration *openapi.MongoDbIntegration) resource.TestCheckFunc {
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

		*mongoDBIntegration = *resp.Data.Mongodb

		return nil
	}
}

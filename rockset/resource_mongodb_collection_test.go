package rockset

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/rockset/rockset-go-client/openapi"
	"github.com/stretchr/testify/assert"
)

const testMongoDBCollectionName = "terraform-provider-acceptance-tests-mongodb"
const testMongoDBCollectionWorkspace = "commons"
const testMongoDBCollectionDescription = "Terraform provider acceptance tests."

func TestAccMongoDBCollection_Basic(t *testing.T) {
	var collection openapi.Collection

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckMongo(t)
		},
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRocksetCollectionDestroy, // Reused from base collection
		Steps: []resource.TestStep{
			{
				Config: getHCL("mongodb_collection.tf"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRocksetCollectionExists("rockset_mongodb_collection.test", &collection), // Reused from base collection
					resource.TestCheckResourceAttr("rockset_mongodb_collection.test", "name", testMongoDBCollectionName),
					resource.TestCheckResourceAttr("rockset_mongodb_collection.test", "workspace", testMongoDBCollectionWorkspace),
					resource.TestCheckResourceAttr("rockset_mongodb_collection.test", "description", testMongoDBCollectionDescription),
					testAccCheckRetentionSecsMatches(&collection, 3600),
					testAccCheckMongoDBSourcesExpected(t, &collection),
				),
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func testAccCheckMongoDBSourcesExpected(t *testing.T, collection *openapi.Collection) resource.TestCheckFunc {
	assert := assert.New(t)

	return func(state *terraform.State) error {
		sources := collection.GetSources()
		assert.Equal(len(sources), 2)

		// With a set, order isn't considered
		var source1Index int
		var source2Index int
		if sources[0].Mongodb.GetCollectionName() == "accounts" {
			source1Index = 0
			source2Index = 1
		} else if sources[0].Mongodb.GetCollectionName() == "customers" {
			source1Index = 1
			source2Index = 0
		} else {
			return fmt.Errorf("Unexpected table name on first source.")
		}

		// Source 1
		assert.Equal(sources[source1Index].GetIntegrationName(), "terraform-provider-acceptance-test-mongodb-collection")
		assert.Equal(sources[source1Index].Mongodb.GetDatabaseName(), "sample_analytics")
		assert.Equal(sources[source1Index].Mongodb.GetCollectionName(), "accounts")

		// Source 2
		assert.Equal(sources[source2Index].GetIntegrationName(), "terraform-provider-acceptance-test-mongodb-collection")
		assert.Equal(sources[source2Index].Mongodb.GetDatabaseName(), "sample_analytics")
		assert.Equal(sources[source2Index].Mongodb.GetCollectionName(), "customers")

		return nil
	}
}

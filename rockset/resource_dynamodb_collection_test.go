package rockset

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/rockset/rockset-go-client/openapi"
	"github.com/stretchr/testify/assert"
)

const testDynamoDBCollectionName = "terraform-provider-acceptance-tests-dynamodb"
const testDynamoDBCollectionWorkspace = "commons"
const testDynamoDBCollectionDescription = "Terraform provider acceptance tests."

func TestAccDynamoDBCollection_Basic(t *testing.T) {
	var collection openapi.Collection

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRocksetCollectionDestroy, // Reused from base collection
		Steps: []resource.TestStep{
			{
				Config: getHCL("dynamodb_collection.tf"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRocksetCollectionExists("rockset_dynamodb_collection.test", &collection), // Reused from base collection
					resource.TestCheckResourceAttr("rockset_dynamodb_collection.test", "name", testDynamoDBCollectionName),
					resource.TestCheckResourceAttr("rockset_dynamodb_collection.test", "workspace", testDynamoDBCollectionWorkspace),
					resource.TestCheckResourceAttr("rockset_dynamodb_collection.test", "description", testDynamoDBCollectionDescription),
					testAccCheckRetentionSecsMatches(&collection, 3600),
					testAccCheckDynamoDBSourcesExpected(t, &collection),
				),
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func testAccCheckDynamoDBSourcesExpected(t *testing.T, collection *openapi.Collection) resource.TestCheckFunc {
	assert := assert.New(t)

	return func(state *terraform.State) error {
		sources := collection.GetSources()
		assert.Equal(len(sources), 2)

		// With a set, order isn't considered
		var source1Index int
		var source2Index int
		if sources[0].Dynamodb.GetTableName() == "terraform-provider-rockset-tests-1" {
			source1Index = 0
			source2Index = 1
		} else if sources[0].Dynamodb.GetTableName() == "terraform-provider-rockset-tests-2" {
			source1Index = 1
			source2Index = 0
		} else {
			return fmt.Errorf("Unexpected table name on first source.")
		}

		// Source 1
		assert.Equal(sources[source1Index].GetIntegrationName(), "terraform-provider-acceptance-test-dynamodb-collection-1")
		assert.Equal(sources[source1Index].Dynamodb.GetRcu(), int64(5))
		assert.Equal(sources[source1Index].Dynamodb.GetAwsRegion(), "us-west-2")
		assert.Equal(sources[source1Index].Dynamodb.GetTableName(), "terraform-provider-rockset-tests-1")
		assert.Equal(sources[source1Index].Dynamodb.Status.GetScanRecordsProcessed(), int64(1))

		// Source 2
		assert.Equal(sources[source2Index].GetIntegrationName(), "terraform-provider-acceptance-test-dynamodb-collection-1")
		assert.Equal(sources[source2Index].Dynamodb.GetRcu(), int64(5))
		assert.Equal(sources[source2Index].Dynamodb.GetAwsRegion(), "us-west-2")
		assert.Equal(sources[source2Index].Dynamodb.GetTableName(), "terraform-provider-rockset-tests-2")
		assert.Equal(sources[source2Index].Dynamodb.Status.GetScanRecordsProcessed(), int64(1))

		return nil
	}
}

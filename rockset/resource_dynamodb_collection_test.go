package rockset

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/rockset/rockset-go-client/openapi"
	"github.com/stretchr/testify/assert"
)

func TestAccDynamoDBCollection_Basic(t *testing.T) {
	var collection openapi.Collection
	rcu := 5
	values := Values{
		Name:        randomName("collection"),
		Collection:  randomName("collection"),
		Description: description(),
		Workspace:   "acc",
		Role:        "arn:aws:iam::469279130686:role/terraform-provider-rockset-tests-dynamo",
		Bucket:      "terraform-provider-rockset-tests",
		RCU:         &rcu,
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRocksetCollectionDestroy, // Reused from base collection
		Steps: []resource.TestStep{
			{
				Config: getHCLTemplate("dynamodb_collection.tf", values),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRocksetCollectionExists("rockset_dynamodb_collection.test",
						&collection), // Reused from base collection
					resource.TestCheckResourceAttr("rockset_dynamodb_collection.test", "name", values.Collection),
					resource.TestCheckResourceAttr("rockset_dynamodb_collection.test", "workspace", values.Workspace),
					resource.TestCheckResourceAttr("rockset_dynamodb_collection.test", "description", values.Description),
					testAccCheckRetentionSecsMatches(&collection, 3600),
					testAccCheckDynamoDBSourcesExpected(t, &collection, values),
				),
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccDynamoDBCollection_ScanTrue(t *testing.T) {
	var collection openapi.Collection
	useScanApi := true
	rcu := 5
	values := Values{
		Name:        randomName("collection"),
		Collection:  randomName("collection"),
		Description: description(),
		Workspace:   "acc",
		Role:        "arn:aws:iam::469279130686:role/terraform-provider-rockset-tests-dynamo",
		Bucket:      "terraform-provider-rockset-tests",
		UseScanApi:  &useScanApi,
		RCU:         &rcu,
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRocksetCollectionDestroy, // Reused from base collection
		Steps: []resource.TestStep{
			{
				Config: getHCLTemplate("dynamodb_collection.tf", values),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRocksetCollectionExists("rockset_dynamodb_collection.test",
						&collection), // Reused from base collection
					resource.TestCheckResourceAttr("rockset_dynamodb_collection.test", "name", values.Collection),
					resource.TestCheckResourceAttr("rockset_dynamodb_collection.test", "workspace", values.Workspace),
					resource.TestCheckResourceAttr("rockset_dynamodb_collection.test", "description", values.Description),
					testAccCheckRetentionSecsMatches(&collection, 3600),
					testAccCheckDynamoDBSourcesExpected(t, &collection, values),
				),
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccDynamoDBCollection_ScanFalse(t *testing.T) {
	var collection openapi.Collection
	useScanApi := false

	values := Values{
		Name:        randomName("collection"),
		Collection:  randomName("collection"),
		Description: description(),
		Workspace:   "acc",
		Role:        "arn:aws:iam::469279130686:role/terraform-provider-rockset-tests-dynamo",
		Bucket:      "terraform-provider-rockset-tests",
		UseScanApi:  &useScanApi,
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRocksetCollectionDestroy, // Reused from base collection
		Steps: []resource.TestStep{
			{
				Config: getHCLTemplate("dynamodb_collection.tf", values),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRocksetCollectionExists("rockset_dynamodb_collection.test",
						&collection), // Reused from base collection
					resource.TestCheckResourceAttr("rockset_dynamodb_collection.test", "name", values.Collection),
					resource.TestCheckResourceAttr("rockset_dynamodb_collection.test", "workspace", values.Workspace),
					resource.TestCheckResourceAttr("rockset_dynamodb_collection.test", "description", values.Description),
					testAccCheckRetentionSecsMatches(&collection, 3600),
					testAccCheckDynamoDBSourcesExpected(t, &collection, values),
				),
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccDynamoDBCollection_ScanNull(t *testing.T) {
	var collection openapi.Collection
	rcu := 5

	values := Values{
		Name:        randomName("collection"),
		Collection:  randomName("collection"),
		Description: description(),
		Workspace:   "acc",
		Role:        "arn:aws:iam::469279130686:role/terraform-provider-rockset-tests-dynamo",
		Bucket:      "terraform-provider-rockset-tests",
		RCU:         &rcu,
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRocksetCollectionDestroy, // Reused from base collection
		Steps: []resource.TestStep{
			{
				Config: getHCLTemplate("dynamodb_collection.tf", values),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRocksetCollectionExists("rockset_dynamodb_collection.test",
						&collection), // Reused from base collection
					resource.TestCheckResourceAttr("rockset_dynamodb_collection.test", "name", values.Collection),
					resource.TestCheckResourceAttr("rockset_dynamodb_collection.test", "workspace", values.Workspace),
					resource.TestCheckResourceAttr("rockset_dynamodb_collection.test", "description", values.Description),
					testAccCheckRetentionSecsMatches(&collection, 3600),
					testAccCheckDynamoDBSourcesExpected(t, &collection, values),
				),
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func testAccCheckDynamoDBSourcesExpected(t *testing.T, collection *openapi.Collection, values Values) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		sources := collection.GetSources()
		assert.Equal(t, len(sources), 2)

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
			return fmt.Errorf("unexpected table name on first source")
		}

		var rcu int64
		if values.RCU == nil {
			rcu = 0
		} else {
			rcu = int64(*values.RCU)
		}

		var useScanApi bool
		if values.UseScanApi == nil {
			useScanApi = true
		} else {
			useScanApi = *values.UseScanApi
		}

		// Source 1
		assert.Equal(t, sources[source1Index].GetIntegrationName(), values.Name)
		assert.Equal(t, sources[source1Index].Dynamodb.GetRcu(), rcu)
		assert.Equal(t, sources[source1Index].Dynamodb.GetAwsRegion(), "us-west-2")
		assert.Equal(t, sources[source1Index].Dynamodb.GetTableName(), "terraform-provider-rockset-tests-1")
		assert.Equal(t, sources[source1Index].Dynamodb.GetUseScanApi(), useScanApi)
		// assert.Equal(sources[source1Index].Dynamodb.Status.GetScanRecordsProcessed(), int64(1))

		// Source 2
		assert.Equal(t, sources[source2Index].GetIntegrationName(), values.Name)
		assert.Equal(t, sources[source2Index].Dynamodb.GetRcu(), rcu)
		assert.Equal(t, sources[source2Index].Dynamodb.GetAwsRegion(), "us-west-2")
		assert.Equal(t, sources[source2Index].Dynamodb.GetTableName(), "terraform-provider-rockset-tests-2")
		assert.Equal(t, sources[source2Index].Dynamodb.GetUseScanApi(), useScanApi)
		// assert.Equal(sources[source2Index].Dynamodb.Status.GetScanRecordsProcessed(), int64(1))

		return nil
	}
}

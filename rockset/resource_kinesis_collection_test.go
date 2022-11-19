package rockset

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/rockset/rockset-go-client/openapi"
	"github.com/stretchr/testify/assert"
)

const testKinesisCollectionName = "terraform-provider-acceptance-tests-kinesis"
const testKinesisCollectionWorkspace = "commons"
const testKinesisCollectionDescription = "Terraform provider acceptance tests."

func TestAccKinesisCollection_Basic(t *testing.T) {
	var collection openapi.Collection

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRocksetCollectionDestroy, // Reused from base collection
		Steps: []resource.TestStep{
			{
				Config: getHCL("kinesis_collection.tf"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRocksetCollectionExists("rockset_kinesis_collection.test", &collection), // Reused from base collection
					resource.TestCheckResourceAttr("rockset_kinesis_collection.test", "name", testKinesisCollectionName),
					resource.TestCheckResourceAttr("rockset_kinesis_collection.test", "workspace", testKinesisCollectionWorkspace),
					resource.TestCheckResourceAttr("rockset_kinesis_collection.test", "description", testKinesisCollectionDescription),
					testAccCheckRetentionSecsMatches(&collection, 3600),
					testAccCheckKinesisSourcesExpected(t, &collection),
				),
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func testAccCheckKinesisSourcesExpected(t *testing.T, collection *openapi.Collection) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		sources := collection.GetSources()

		assert.Equal(t, len(sources), 3)

		//With a set, order isn't considered
		var jsonSource, pgSource, mysqlSource openapi.Source
		var jsonIdx, pgIdx, mysqlIdx int
		for idx, source := range sources {
			if source.FormatParams == nil {
				jsonIdx = idx
			} else if source.FormatParams.GetMysqlDms() {
				mysqlIdx = idx
			} else if source.FormatParams.GetPostgresDms() {
				pgIdx = idx
			}
		}
		jsonSource = sources[jsonIdx]
		pgSource = sources[pgIdx]
		mysqlSource = sources[mysqlIdx]

		assert.Equal(t, jsonSource.GetIntegrationName(), "terraform-provider-acceptance-test-kinesis-collection")
		assert.Equal(t, pgSource.GetIntegrationName(), "terraform-provider-acceptance-test-kinesis-collection")
		assert.Equal(t, mysqlSource.GetIntegrationName(), "terraform-provider-acceptance-test-kinesis-collection")

		assert.Equal(t, jsonSource.Kinesis.GetStreamName(), "terraform-provider-rockset-tests-kinesis")
		assert.Equal(t, pgSource.Kinesis.GetStreamName(), "terraform-provider-rockset-tests-kinesis")
		assert.Equal(t, mysqlSource.Kinesis.GetStreamName(), "terraform-provider-rockset-tests-kinesis")

		assert.Empty(t, jsonSource.Kinesis.GetDmsPrimaryKey())
		assert.Equal(t, pgSource.Kinesis.GetDmsPrimaryKey()[0], "foo")
		assert.Equal(t, pgSource.Kinesis.GetDmsPrimaryKey()[1], "bar")
		assert.Equal(t, mysqlSource.Kinesis.GetDmsPrimaryKey()[0], "foo")
		assert.Equal(t, mysqlSource.Kinesis.GetDmsPrimaryKey()[1], "bar")

		return nil
	}
}

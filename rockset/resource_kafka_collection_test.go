package rockset

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/rockset/rockset-go-client/openapi"
)

func TestAccKafkaCollection_BasicV2(t *testing.T) {
	var collection openapi.Collection

	t.Skip("requires special setup")

	resource.ParallelTest(t, resource.TestCase{
		ExternalProviders: map[string]resource.ExternalProvider{
			"null": {
				VersionConstraint: "3.1.1",
				Source:            "hashicorp/null",
			},
		},
		PreCheck:          func() { testAccPreCheck(t, "TF_VAR_KAFKA_BROKER") },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRocksetCollectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: getHCL("kafka_collection.tf"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRocksetCollectionExists("rockset_kafka_collection.test", &collection),
					resource.TestCheckResourceAttr("rockset_kafka_collection.test", "name", "kafka-collection"),
					resource.TestCheckResourceAttr("rockset_kafka_collection.test", "workspace", "acc"),
					resource.TestCheckResourceAttr("rockset_kafka_collection.test", "description", "Terraform provider acceptance tests."),
					testAccCheckRetentionSecsMatches(&collection, 3600),
					testAccCheckKafkaSourceExpected(t, &collection),
				),
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccKafkaCollection_BasicV3(t *testing.T) {
	t.Skip("kafka needs to be reconfigured")
	var collection openapi.Collection

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRocksetCollectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: getHCL("kafka_collection_v3.tf"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRocksetCollectionExists("rockset_kafka_collection.test", &collection),
					resource.TestCheckResourceAttr("rockset_kafka_collection.test", "name", "kafka-collection"),
					resource.TestCheckResourceAttr("rockset_kafka_collection.test", "workspace", "acc"),
					resource.TestCheckResourceAttr("rockset_kafka_collection.test", "description", "Terraform provider acceptance tests."),
					testAccCheckRetentionSecsMatches(&collection, 3600),
					testAccCheckKafkaSourceExpected(t, &collection),
				),
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func testAccCheckKafkaSourceExpected(t *testing.T, collection *openapi.Collection) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		require.Len(t, collection.Sources, 1)
		k := collection.Sources[0].Kafka
		require.NotNil(t, k)
		assert.Equal(t, "test_json", k.GetKafkaTopicName())
		//assert.Equal(t, "EARLIEST", k.GetOffsetResetPolicy())

		return nil
	}
}

package rockset

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/rockset/rockset-go-client"
	"github.com/rockset/rockset-go-client/openapi"
)

const testKafkaIntegrationName = "terraform-provider-acceptance-test-kafka-integration"
const testKafkaIntegrationDescription = "Terraform provider acceptance tests."

func TestAccKafkaIntegration_BasicV2(t *testing.T) {
	var kafkaIntegration openapi.KafkaIntegration

	t.Skip("requires special setup")

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRocksetIntegrationDestroy("rockset_kafka_integration"),
		Steps: []resource.TestStep{
			{
				Config: getHCL("kafka_integration.tf"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRocksetKafkaIntegrationExists("rockset_kafka_integration.test",
						&kafkaIntegration),
					resource.TestCheckResourceAttr("rockset_kafka_integration.test", "name",
						testKafkaIntegrationName),
					resource.TestCheckResourceAttr("rockset_kafka_integration.test", "description",
						testKafkaIntegrationDescription),
					resource.TestCheckResourceAttr("rockset_kafka_integration.test", "kafka_data_format",
						"JSON"),
					resource.TestCheckResourceAttr("rockset_kafka_integration.test", "kafka_topic_names.0",
						"bar"),
					resource.TestCheckResourceAttr("rockset_kafka_integration.test", "kafka_topic_names.1",
						"foo"),
					testAccCheckKafkaConnectionString(t, &kafkaIntegration),
				),
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccKafkaIntegration_BasicV3(t *testing.T) {
	var kafkaIntegration openapi.KafkaIntegration

	t.Skip("test broken due to API change")

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t,
				"TF_VAR_CC_BOOTSTRAP_SERVERS", "TF_VAR_CC_SECRET", "TF_VAR_CC_KEY")
		},
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRocksetIntegrationDestroy("rockset_kafka_integration"),
		Steps: []resource.TestStep{
			{
				Config: getHCL("kafka_integration_v3.tf"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRocksetKafkaIntegrationExists("rockset_kafka_integration.test",
						&kafkaIntegration),
					resource.TestCheckResourceAttr("rockset_kafka_integration.test", "name",
						testKafkaIntegrationName),
					resource.TestCheckResourceAttr("rockset_kafka_integration.test", "description",
						testKafkaIntegrationDescription),
				),
				ExpectNonEmptyPlan: false,
			},
			{
				// validate that changing the secret forces a new resource
				Config:             getHCL("kafka_integration_v3b.tf"),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckRocksetKafkaIntegrationExists(resource string, kafkaIntegration *openapi.KafkaIntegration) resource.TestCheckFunc {
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

		*kafkaIntegration = *resp.Kafka

		return nil
	}
}

func testAccCheckKafkaConnectionString(t *testing.T, collection *openapi.KafkaIntegration) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		require.NotNil(t, collection)
		assert.Regexp(t, `kafka://.*@api\.usw2a1\.rockset\.com`, collection.GetConnectionString())
		return nil
	}
}

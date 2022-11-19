package rockset

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/rockset/rockset-go-client/openapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testGCSCollectionName = "terraform-provider-acceptance-tests-gcs"
const testGCSCollectionWorkspace = "commons"
const testGCSCollectionDescription = "Terraform provider acceptance tests."

func TestAccGCSCollection_Basic(t *testing.T) {
	var collection openapi.Collection

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t, "TF_VAR_GCS_SERVICE_ACCOUNT_KEY") },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRocksetCollectionDestroy, // Reused from base collection
		Steps: []resource.TestStep{
			{
				Config: getHCL("gcs_collection_csv.tf"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRocksetCollectionExists("rockset_gcs_collection.test", &collection), // Reused from base collection
					resource.TestCheckResourceAttr("rockset_gcs_collection.test", "name", testGCSCollectionName),
					resource.TestCheckResourceAttr("rockset_gcs_collection.test", "workspace", testGCSCollectionWorkspace),
					resource.TestCheckResourceAttr("rockset_gcs_collection.test", "description", testGCSCollectionDescription),
					testAccCheckRetentionSecsMatches(&collection, 3600),
					testAccCheckGCSSourcesExpected(t, &collection),
				),
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func testAccCheckGCSSourcesExpected(t *testing.T, collection *openapi.Collection) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		sources := collection.GetSources()
		require.Len(t, sources, 1)
		source := sources[0]

		assert.Equal(t, source.GetIntegrationName(), "terraform-provider-acceptance-tests-gcs-collection")
		require.Len(t, collection.GetFieldMappings(), 2)
		return nil
	}
}

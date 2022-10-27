package rockset

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/rockset/rockset-go-client/openapi"
	"testing"
)

func TestAccSampleCollection_Basic(t *testing.T) {
	var collection openapi.Collection

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRocksetCollectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: getHCL("sample_collection.tf"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRocksetCollectionExists("rockset_sample_collection.test", &collection),
					resource.TestCheckResourceAttr("rockset_sample_collection.test", "dataset", "cities"),
				),
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

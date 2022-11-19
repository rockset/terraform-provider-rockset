package rockset

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/rockset/rockset-go-client/openapi"
)

func TestAccS3Collection_Basic(t *testing.T) {
	var collection openapi.Collection

	name := randomName("s3")
	values := Values{
		Name:        name,
		Collection:  name,
		Workspace:   name,
		Description: description(),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRocksetCollectionDestroy, // Reused from base collection
		Steps: []resource.TestStep{
			{
				Config: getHCLTemplate("s3_collection.tf", values),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRocksetCollectionExists("rockset_s3_collection.test", &collection), // Reused from base collection
					resource.TestCheckResourceAttr("rockset_s3_collection.test", "name", values.Collection),
					resource.TestCheckResourceAttr("rockset_s3_collection.test", "workspace", values.Workspace),
					resource.TestCheckResourceAttr("rockset_s3_collection.test", "description", values.Description),
					testAccCheckRetentionSecsMatches(&collection, 3600),
				),
				Destroy:            false,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccS3Collection_Json(t *testing.T) {
	var collection openapi.Collection

	name := randomName("s3-json")
	values := Values{
		Name:        name,
		Collection:  name,
		Workspace:   name,
		Description: description(),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRocksetCollectionDestroy, // Reused from base collection
		Steps: []resource.TestStep{
			{
				Config: getHCLTemplate("s3_collection_json.tf", values),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRocksetCollectionExists("rockset_s3_collection.test", &collection), // Reused from base collection
					resource.TestCheckResourceAttr("rockset_s3_collection.test", "name", values.Name),
					resource.TestCheckResourceAttr("rockset_s3_collection.test", "workspace", values.Workspace),
					resource.TestCheckResourceAttr("rockset_s3_collection.test", "description", values.Description),
					testAccCheckRetentionSecsMatches(&collection, 3600),
				),
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

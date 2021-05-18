package rockset

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/rockset/rockset-go-client"
)

const testCollectionName = "terraform-provider-acceptance-tests-1"
const testCollectionWorkspace = "commons"
const testCollectionDescription = "Terraform provider acceptance tests."

func TestAccCollection_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRocksetCollectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCollectionBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRocksetCollectionExists("rockset_collection.test"),
					resource.TestCheckResourceAttr("rockset_collection.test", "name", testCollectionName),
					resource.TestCheckResourceAttr("rockset_collection.test", "workspace", testCollectionWorkspace),
					resource.TestCheckResourceAttr("rockset_collection.test", "description", testCollectionDescription),
				),
				ExpectNonEmptyPlan: false,
			},
			{
				Config: testAccCheckCollectionUpdateForceRecreate(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRocksetCollectionExists("rockset_collection.test"),
					resource.TestCheckResourceAttr("rockset_collection.test", "name", fmt.Sprintf("%s-updated", testCollectionName)),
					resource.TestCheckResourceAttr("rockset_collection.test", "workspace", testCollectionWorkspace),
					resource.TestCheckResourceAttr("rockset_collection.test", "description", testCollectionDescription),
				),
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func testAccCheckCollectionBasic() string {
	return fmt.Sprintf(`
resource rockset_collection test {
	name        = "%s"
	workspace   = "%s"
	description = "%s"
}
`, testCollectionName, testCollectionWorkspace, testCollectionDescription)
}

func testAccCheckCollectionUpdateForceRecreate() string {
	return fmt.Sprintf(`
resource rockset_collection test {
	name        = "%s-updated"
	workspace   = "%s"
	description = "%s"
}`, testCollectionName, testCollectionWorkspace, testCollectionDescription)
}

func testAccCheckRocksetCollectionDestroy(s *terraform.State) error {
	rc := testAccProvider.Meta().(*rockset.RockClient)
	ctx := context.TODO()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "rockset_collection" {
			continue
		}

		workspace, name := workspaceAndNameFromID(rs.Primary.ID)

		_, err := rc.GetCollection(ctx, workspace, name)

		// An error would mean we didn't find the key, we expect an error
		if err == nil {
			// We did not get an error, so we failed to delete the key.
			return fmt.Errorf("Collection %s still exists.", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckRocksetCollectionExists(resource string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rc := testAccProvider.Meta().(*rockset.RockClient)
		ctx := context.TODO()

		rs, err := getResourceFromState(state, resource)
		if err != nil {
			return err
		}

		workspace, name := workspaceAndNameFromID(rs.Primary.ID)
		_, err = rc.GetCollection(ctx, workspace, name)
		if err != nil {
			return err
		}

		return nil
	}
}

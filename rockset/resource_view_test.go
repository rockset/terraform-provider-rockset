package rockset

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/rockset/rockset-go-client"
	"github.com/rockset/rockset-go-client/openapi"
	"testing"
)

const testViewName = "terraform-provider-acceptance-tests"
const testViewWorkspace = "commons"
const testViewQuery = "select * from commons._events where _events.kind = 'COLLECTION'"
const testViewDescription = "Terraform provider acceptance tests"

func TestAccView_Basic(t *testing.T) {
	var view openapi.View

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRocksetViewDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckViewBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRocksetViewExists("rockset_view.test", &view),
					resource.TestCheckResourceAttr("rockset_view.test", "name", testViewName),
					resource.TestCheckResourceAttr("rockset_view.test", "query", testViewQuery),
					resource.TestCheckResourceAttr("rockset_view.test", "description", testViewDescription),
					resource.TestCheckResourceAttrSet("rockset_view.test", "created_by"),
				),
				ExpectNonEmptyPlan: false,
			},
			{
				Config: testAccCheckViewUpdateName(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRocksetViewExists("rockset_view.test", &view),
					resource.TestCheckResourceAttr("rockset_view.test", "name", fmt.Sprintf("%s-updated", testViewName)),
					resource.TestCheckResourceAttr("rockset_view.test", "query", testViewQuery),
					resource.TestCheckResourceAttr("rockset_view.test", "description", testViewDescription),
					resource.TestCheckResourceAttrSet("rockset_view.test", "created_by"),
				),
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func testAccCheckViewBasic() string {
	return fmt.Sprintf(`
resource rockset_view test {
	workspace   = "%s"
	name        = "%s"
	query       = "%s"
	description	= "%s"
}
`, testViewWorkspace, testViewName, testViewQuery, testViewDescription)
}

func testAccCheckViewUpdateName() string {
	return fmt.Sprintf(`
resource rockset_view test {
	workspace   = "%s"
	name        = "%s-updated"
	query       = "%s"
	description	= "%s"
}
`, testViewWorkspace, testViewName, testViewQuery, testViewDescription)
}

func testAccCheckRocksetViewDestroy(s *terraform.State) error {
	rc := testAccProvider.Meta().(*rockset.RockClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "rockset_view" {
			continue
		}

		workspace, name := workspaceAndNameFromID(rs.Primary.ID)
		_, err := rc.GetView(testCtx, workspace, name)

		// An error would mean we didn't find the key, we expect an error
		if err == nil {
			// We did not get an error, so we failed to delete the key.
			return fmt.Errorf("view %s.%s still exists", workspace, name)
		}
	}

	return nil
}

func testAccCheckRocksetViewExists(resource string, view *openapi.View) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rc := testAccProvider.Meta().(*rockset.RockClient)

		rs, err := getResourceFromState(state, resource)
		if err != nil {
			return err
		}

		workspace, name := workspaceAndNameFromID(rs.Primary.ID)

		resp, err := rc.GetView(testCtx, workspace, name)
		if err != nil {
			return err
		}

		*view = resp

		return nil
	}
}

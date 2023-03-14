package rockset

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/rockset/rockset-go-client"
	"github.com/rockset/rockset-go-client/openapi"
)

func TestAccView_Basic(t *testing.T) {
	var view openapi.View

	name := randomName("view")
	v1 := Values{
		Name:        name,
		Collection:  name,
		Workspace:   name,
		Description: description(),
		SQL:         "select * from commons._events",
	}
	v2 := v1
	v2.SQL = "select * from commons._events where _events.kind = 'COLLECTION'"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRocksetViewDestroy,
		Steps: []resource.TestStep{
			{
				Config: getHCLTemplate("view_basic.tf", v1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRocksetViewExists("rockset_view.test", &view),
					resource.TestCheckResourceAttr("rockset_view.test", "name", v1.Name),
					resource.TestCheckResourceAttr("rockset_view.test", "query", v1.SQL),
					resource.TestCheckResourceAttr("rockset_view.test", "description", v1.Description),
					resource.TestCheckResourceAttrSet("rockset_view.test", "created_by"),
				),
				ExpectNonEmptyPlan: false,
			},
			{
				Config: getHCLTemplate("view_basic.tf", v2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRocksetViewExists("rockset_view.test", &view),
					resource.TestCheckResourceAttr("rockset_view.test", "name", v2.Name),
					resource.TestCheckResourceAttr("rockset_view.test", "query", v2.SQL),
					resource.TestCheckResourceAttr("rockset_view.test", "description", v2.Description),
					resource.TestCheckResourceAttrSet("rockset_view.test", "created_by"),
				),
				ExpectNonEmptyPlan: false,
			},
		},
	})
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

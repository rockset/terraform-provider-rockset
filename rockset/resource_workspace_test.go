package rockset

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/rockset/rockset-go-client"
	"github.com/rockset/rockset-go-client/openapi"
)

const testWorkspaceName = "terraform-provider-acceptance-tests"
const testWorkspaceDescription = "Terraform provider acceptance tests"

func TestAccWorkspace_Basic(t *testing.T) {
	var workspace openapi.Workspace

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRocksetWorkspaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckWorkspaceBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRocksetWorkspaceExists("rockset_workspace.test", &workspace),
					resource.TestCheckResourceAttr("rockset_workspace.test", "name", testWorkspaceName),
					resource.TestCheckResourceAttr("rockset_workspace.test", "description", testWorkspaceDescription),
					resource.TestCheckResourceAttrSet("rockset_workspace.test", "created_by"),
				),
				ExpectNonEmptyPlan: false,
			},
			{
				Config: testAccCheckWorkspaceUpdateName(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRocksetWorkspaceExists("rockset_workspace.test", &workspace),
					resource.TestCheckResourceAttr("rockset_workspace.test", "name", fmt.Sprintf("%s-updated", testWorkspaceName)),
					resource.TestCheckResourceAttr("rockset_workspace.test", "description", testWorkspaceDescription),
					resource.TestCheckResourceAttrSet("rockset_workspace.test", "created_by"),
				),
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func testAccCheckWorkspaceBasic() string {
	return fmt.Sprintf(`
resource rockset_workspace test {
	name        = "%s"
	description	= "%s"
}
`, testWorkspaceName, testWorkspaceDescription)
}

func testAccCheckWorkspaceUpdateName() string {
	return fmt.Sprintf(`
resource rockset_workspace test {
	name        = "%s-updated"
	description	= "%s"
}
`, testWorkspaceName, testWorkspaceDescription)
}

func testAccCheckRocksetWorkspaceDestroy(s *terraform.State) error {
	rc := testAccProvider.Meta().(*rockset.RockClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "rockset_api_key" {
			continue
		}

		name := rs.Primary.ID
		_, err := rc.GetWorkspace(testCtx, name)

		// An error would mean we didn't find the key, we expect an error
		if err == nil {
			// We did not get an error, so we failed to delete the key.
			return fmt.Errorf("Workspace %s still exists.", name)
		}
	}

	return nil
}

func testAccCheckRocksetWorkspaceExists(resource string, workspace *openapi.Workspace) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rc := testAccProvider.Meta().(*rockset.RockClient)

		rs, err := getResourceFromState(state, resource)
		if err != nil {
			return err
		}

		name := rs.Primary.ID

		resp, err := rc.GetWorkspace(testCtx, name)
		if err != nil {
			return err
		}

		*workspace = resp

		return nil
	}
}

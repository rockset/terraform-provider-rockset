package rockset

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/rockset/rockset-go-client"
	"github.com/rockset/rockset-go-client/openapi"
)

func TestAccWorkspace_Basic(t *testing.T) {
	var workspace openapi.Workspace

	type values struct {
		Name        string
		Description string
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRocksetWorkspaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: getHCLTemplate("workspace_basic.tf", values{"acc-ws", "Terraform provider acceptance tests"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRocksetWorkspaceExists("rockset_workspace.test", &workspace),
					resource.TestCheckResourceAttr("rockset_workspace.test", "name", "acc-ws"),
					resource.TestCheckResourceAttr("rockset_workspace.test", "description", "Terraform provider acceptance tests"),
					resource.TestCheckResourceAttrSet("rockset_workspace.test", "created_by"),
				),
				ExpectNonEmptyPlan: false,
			},
			{
				Config: getHCLTemplate("workspace_basic.tf", values{"acc-ws-updated", "Terraform provider acceptance tests"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRocksetWorkspaceExists("rockset_workspace.test", &workspace),
					resource.TestCheckResourceAttr("rockset_workspace.test", "name", "acc-ws-updated"),
					resource.TestCheckResourceAttr("rockset_workspace.test", "description", "Terraform provider acceptance tests"),
					resource.TestCheckResourceAttrSet("rockset_workspace.test", "created_by"),
				),
				ExpectNonEmptyPlan: false,
			},
		},
	})
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
			return fmt.Errorf("workspace %s still exists", name)
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

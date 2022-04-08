package rockset

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/rockset/rockset-go-client"
	"github.com/rockset/rockset-go-client/openapi"
	"reflect"
	"regexp"
	"strings"
	"testing"
)

const testRoleName = "terraform-provider-acceptance-tests"
const testRoleDescription = "Terraform provider acceptance tests"

func TestAccRole_Basic(t *testing.T) {
	var role openapi.Role

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRocksetRoleDestroy,
		Steps: []resource.TestStep{
			{
				Config: getHCL("role.tf"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRocksetRoleExists("rockset_role.test", &role),
					resource.TestCheckResourceAttr("rockset_role.test", "name", testRoleName),
					resource.TestCheckResourceAttr("rockset_role.test", "description", testRoleDescription),
					testAccCheckRocksetRolePrivileges(&role, []openapi.Privilege{
						{
							Action:       openapi.PtrString("QUERY_DATA_WS"),
							ResourceName: openapi.PtrString("common"),
							Cluster:      openapi.PtrString("*ALL*"),
						},
						{
							Action:       openapi.PtrString("CREATE_COLLECTION_INTEGRATION"),
							ResourceName: openapi.PtrString("dummy"),
						},
						{
							Action: openapi.PtrString("GET_METRICS_GLOBAL"),
						},
					}),
				),
				ExpectNonEmptyPlan: false,
			},
			{
				Config: getHCL("role_bad_global.tf"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRocksetRoleExists("rockset_role.test", &role),
					resource.TestCheckResourceAttr("rockset_role.test", "name", testRoleName),
					resource.TestCheckResourceAttr("rockset_role.test", "description", testRoleDescription),
				),
				ExpectError: regexp.MustCompile("can't specify resource_name for UPDATE_VI_GLOBAL action"),
			},
			{
				Config: getHCL("role_bad_integration.tf"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRocksetRoleExists("rockset_role.test", &role),
					resource.TestCheckResourceAttr("rockset_role.test", "name", testRoleName),
					resource.TestCheckResourceAttr("rockset_role.test", "description", testRoleDescription),
				),
				ExpectError: regexp.MustCompile("can't specify cluster for CREATE_COLLECTION_INTEGRATION action"),
			},
		},
	})
}

func testAccCheckRocksetRolePrivileges(role *openapi.Role, privs []openapi.Privilege) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		if !reflect.DeepEqual(role.GetPrivileges(), privs) {
			var b strings.Builder
			b.WriteString("\nexpected")

			for _, p := range privs {
				b.WriteString(fmt.Sprintf("action: %s\n", p.GetAction()))
				b.WriteString(fmt.Sprintf("resource name: %s\n", p.GetResourceName()))
				b.WriteString(fmt.Sprintf("cluster: %s\n", p.GetCluster()))
			}

			b.WriteString("\nactual\n")

			for _, p := range role.Privileges {
				b.WriteString(fmt.Sprintf("action: %s\n", p.GetAction()))
				b.WriteString(fmt.Sprintf("resource name: %s\n", p.GetResourceName()))
				b.WriteString(fmt.Sprintf("cluster: %s\n", p.GetCluster()))
			}

			return fmt.Errorf(b.String())
		}

		return nil
	}
}

func testAccCheckRocksetRoleDestroy(s *terraform.State) error {
	rc := testAccProvider.Meta().(*rockset.RockClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "rockset_role" {
			continue
		}

		_, err := rc.GetRole(testCtx, rs.Primary.ID)

		// An error would mean we didn't find the key, we expect an error
		if err == nil {
			// We did not get an error, so we failed to delete the key.
			return fmt.Errorf("role %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckRocksetRoleExists(resource string, role *openapi.Role) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rc := testAccProvider.Meta().(*rockset.RockClient)

		rs, err := getResourceFromState(state, resource)
		if err != nil {
			return err
		}

		resp, err := rc.GetRole(testCtx, rs.Primary.ID)
		if err != nil {
			return err
		}

		*role = resp

		return nil
	}
}

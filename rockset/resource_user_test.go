package rockset

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/rockset/rockset-go-client"
	"github.com/rockset/rockset-go-client/openapi"
	"reflect"
	"testing"
)

func TestAccUser_Basic(t *testing.T) {
	var user openapi.User

	type values struct {
		Email     string
		FirstName string
		LastName  string
		Roles     []string
	}
	var first = values{
		Email: "acc@rockset.com",
		Roles: []string{"read-only"},
	}
	var second = values{
		Email:     "acc@rockset.com",
		Roles:     []string{"member"},
		FirstName: "Acceptance",
		LastName:  "Testing",
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRocksetUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: getHCLTemplate("user_basic.tftpl", first),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRocksetUserExists("rockset_user.test", &user),
					resource.TestCheckResourceAttr("rockset_user.test", "email", "acc@rockset.com"),
					testAccUserRoleListMatches(&user, []string{"read-only"}),
				),
				ExpectNonEmptyPlan: false,
			},
			{
				Config: getHCLTemplate("user_basic.tftpl", second),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRocksetUserExists("rockset_user.test", &user),
					resource.TestCheckResourceAttr("rockset_user.test", "email", "acc@rockset.com"),
					resource.TestCheckResourceAttr("rockset_user.test", "first_name", second.FirstName),
					resource.TestCheckResourceAttr("rockset_user.test", "last_name", second.LastName),
					testAccUserRoleListMatches(&user, []string{"member"}),
				),
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func testAccCheckRocksetUserDestroy(s *terraform.State) error {
	rc := testAccProvider.Meta().(*rockset.RockClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "rockset_user" {
			continue
		}

		email := rs.Primary.ID
		_, err := rc.GetUser(testCtx, email)

		// An error would mean we didn't find the key, we expect an error
		if err == nil {
			// We did not get an error, so we failed to delete the key.
			return fmt.Errorf("user %s still exists", email)
		}
	}

	return nil
}

func testAccCheckRocksetUserExists(resource string, user *openapi.User) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rc := testAccProvider.Meta().(*rockset.RockClient)

		rs, err := getResourceFromState(state, resource)
		if err != nil {
			return err
		}

		email := rs.Primary.ID
		resp, err := rc.GetUser(testCtx, email)
		if err != nil {
			return err
		}

		*user = resp

		return nil
	}
}

func testAccUserRoleListMatches(user *openapi.User, expectedRoles []string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		if !reflect.DeepEqual(user.GetRoles(), expectedRoles) {
			return fmt.Errorf("expected %s roles, got %s", expectedRoles, user.GetRoles())
		}

		return nil
	}
}

package rockset

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/rockset/rockset-go-client"
	"github.com/rockset/rockset-go-client/openapi"
)

const testUserEmail = "terraform-provider-acceptance-tests@rockset.com"
const testUserRole1 = "read-only"
const testUserRole2 = "member"

func TestAccUser_Basic(t *testing.T) {
	var user openapi.User

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRocksetUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckUserBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRocksetUserExists("rockset_user.test", &user),
					resource.TestCheckResourceAttr("rockset_user.test", "email", testUserEmail),
					testAccUserRoleListMatches(&user, []string{testUserRole1}),
				),
				ExpectNonEmptyPlan: false,
			},
			{
				Config: testAccCheckUserTwoRoles(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRocksetUserExists("rockset_user.test", &user),
					resource.TestCheckResourceAttr("rockset_user.test", "email", testUserEmail),
					testAccUserRoleListMatches(&user, []string{testUserRole1, testUserRole2}),
				),
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func testAccCheckUserBasic() string {
	return fmt.Sprintf(`
resource rockset_user test {
	email        = "%s"
	roles				 = ["%s"]
}
`, testUserEmail, testUserRole1)
}

func testAccCheckUserTwoRoles() string {
	return fmt.Sprintf(`
resource rockset_user test {
	email        = "%s"
	roles				 = ["%s", "%s"]
}
`, testUserEmail, testUserRole1, testUserRole2)
}

func testAccCheckRocksetUserDestroy(s *terraform.State) error {
	rc := testAccProvider.Meta().(*rockset.RockClient)
	ctx := context.TODO()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "rockset_user" {
			continue
		}

		email := rs.Primary.ID
		_, err := getUserByEmail(ctx, rc, email)

		// An error would mean we didn't find the key, we expect an error
		if err == nil {
			// We did not get an error, so we failed to delete the key.
			return fmt.Errorf("User %s still exists.", email)
		}
	}

	return nil
}

func testAccCheckRocksetUserExists(resource string, user *openapi.User) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rc := testAccProvider.Meta().(*rockset.RockClient)
		ctx := context.TODO()

		rs, err := getResourceFromState(state, resource)
		if err != nil {
			return err
		}

		email := rs.Primary.ID
		resp, err := getUserByEmail(ctx, rc, email)
		if err != nil {
			return err
		}

		*user = *resp

		return nil
	}
}

func testAccUserRoleListMatches(user *openapi.User, expectedRoles []string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		if !reflect.DeepEqual(user.GetRoles(), expectedRoles) {
			return fmt.Errorf("Expected %s collections, got %s.", expectedRoles, user.GetRoles())
		}

		return nil
	}
}

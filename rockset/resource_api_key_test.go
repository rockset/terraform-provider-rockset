package rockset

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/rockset/rockset-go-client"
)

const testApiKeyName = "terraform-provider-acceptance-tests"
const testApiKeyUser = "john@rockset.com" // TODO, replace with user we create!

func TestAccApiKey_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRocksetApiKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckApiKeyBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRocksetApiKeyExists("rockset_api_key.test"),
					resource.TestCheckResourceAttr("rockset_api_key.test", "name", testApiKeyName),
					resource.TestCheckNoResourceAttr("rockset_api_key.test", "user"),
					resource.TestCheckResourceAttrSet("rockset_api_key.test", "key"),
				),
				ExpectNonEmptyPlan: false,
			},
			{
				Config: testAccCheckApiKeyUpdateName(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRocksetApiKeyExists("rockset_api_key.test"),
					resource.TestCheckResourceAttr("rockset_api_key.test", "name", fmt.Sprintf("%s-updated", testApiKeyName)),
					resource.TestCheckNoResourceAttr("rockset_api_key.test", "user"),
					resource.TestCheckResourceAttrSet("rockset_api_key.test", "key"),
				),
				ExpectNonEmptyPlan: false,
			},
			{
				Config: testAccCheckApiKeyUpdateUser(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRocksetApiKeyExists("rockset_api_key.test"),
					resource.TestCheckResourceAttr("rockset_api_key.test", "name", fmt.Sprintf("%s-updated", testApiKeyName)),
					resource.TestCheckResourceAttr("rockset_api_key.test", "user", testApiKeyUser),
					resource.TestCheckResourceAttrSet("rockset_api_key.test", "key"),
				),
				ExpectNonEmptyPlan: false,
			},
			{
				// Back to basic, will change name AND api key
				Config: testAccCheckApiKeyBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRocksetApiKeyExists("rockset_api_key.test"),
					resource.TestCheckResourceAttr("rockset_api_key.test", "name", testApiKeyName),
					resource.TestCheckNoResourceAttr("rockset_api_key.test", "user"),
					resource.TestCheckResourceAttrSet("rockset_api_key.test", "key"),
				),
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func testAccCheckApiKeyBasic() string {
	return fmt.Sprintf(`
resource rockset_api_key test {
	name        = "%s"
}
`, testApiKeyName)
}

func testAccCheckApiKeyUpdateName() string {
	return fmt.Sprintf(`
resource rockset_api_key test {
	name        = "%s-updated"
}
`, testApiKeyName)
}

func testAccCheckApiKeyUpdateUser() string {
	return fmt.Sprintf(`
resource rockset_api_key test {
	name        = "%s-updated"
	user				= "%s" 
}
`, testApiKeyName, testApiKeyUser)
}

func testAccCheckRocksetApiKeyDestroy(s *terraform.State) error {
	rc := testAccProvider.Meta().(*rockset.RockClient)
	ctx := context.TODO()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "rockset_api_key" {
			continue
		}

		name, user, err := idToUserAndName(rs.Primary.ID)
		if err != nil {
			return err
		}
		_, err = getApiKey(ctx, rc, name, user)

		// An error would mean we didn't find the key, we expect an error
		if err == nil {
			// We did not get an error, so we failed to delete the key.
			return fmt.Errorf("Api Key %s still exists.", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckRocksetApiKeyExists(resource string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rc := testAccProvider.Meta().(*rockset.RockClient)

		rs, err := getResourceFromState(state, resource)
		if err != nil {
			return err
		}

		name, user, err := idToUserAndName(rs.Primary.ID)
		if err != nil {
			return err
		}

		ctx := context.TODO()
		_, err = getApiKey(ctx, rc, name, user)
		if err != nil {
			return err
		}

		return nil
	}
}

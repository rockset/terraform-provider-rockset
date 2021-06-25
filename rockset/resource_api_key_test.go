package rockset

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/rockset/rockset-go-client"
	"github.com/rockset/rockset-go-client/openapi"
)

const testApiKeyName = "terraform-provider-acceptance-tests"
const testApiKeyUser = "terraform-provider-tests-apikey-user@rockset.com"

func TestAccApiKey_Basic(t *testing.T) {
	var apiKey openapi.ApiKey

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy: testAccCheckRocksetApiKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckApiKeyBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRocksetApiKeyExists("rockset_api_key.test", &apiKey),
					resource.TestCheckResourceAttr("rockset_api_key.test", "name", testApiKeyName),
					resource.TestCheckNoResourceAttr("rockset_api_key.test", "user"),
					resource.TestCheckResourceAttrSet("rockset_api_key.test", "key"),
				),
				ExpectNonEmptyPlan: false,
			},
			{
				Config: testAccCheckApiKeyUpdateName(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRocksetApiKeyExists("rockset_api_key.test", &apiKey),
					resource.TestCheckResourceAttr("rockset_api_key.test", "name", fmt.Sprintf("%s-updated", testApiKeyName)),
					resource.TestCheckNoResourceAttr("rockset_api_key.test", "user"),
					resource.TestCheckResourceAttrSet("rockset_api_key.test", "key"),
				),
				ExpectNonEmptyPlan: false,
			},
			{
				Config: testAccCheckApiKeyUpdateUser(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRocksetApiKeyExists("rockset_api_key.test", &apiKey),
					resource.TestCheckResourceAttr("rockset_api_key.test", "name", fmt.Sprintf("%s-updated", testApiKeyName)),
					resource.TestCheckResourceAttrSet("rockset_api_key.test", "key"),
				),
				ExpectNonEmptyPlan: false,
			},
			{
				// Back to basic, will change name AND api key
				Config: testAccCheckApiKeyBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRocksetApiKeyExists("rockset_api_key.test", &apiKey),
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
resource rockset_user test {
	email        = "%s"
	roles				 = ["read-only"]
}
resource rockset_api_key test {
	name        = "%s-updated"
}
`, testApiKeyUser, testApiKeyName)
}

func testAccCheckRocksetApiKeyDestroy(s *terraform.State) error {
	rc := testAccProvider.Meta().(*rockset.RockClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "rockset_api_key" {
			continue
		}

		name, user, err := idToUserAndName(rs.Primary.ID)
		if err != nil {
			return err
		}
		_, err = getApiKey(testCtx, rc, name, user)

		// An error would mean we didn't find the key, we expect an error
		if err == nil {
			// We did not get an error, so we failed to delete the key.
			return fmt.Errorf("Api Key %s still exists.", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckRocksetApiKeyExists(resource string, apiKey *openapi.ApiKey) resource.TestCheckFunc {
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

		resp, err := getApiKey(testCtx, rc, name, user)
		if err != nil {
			return err
		}

		*apiKey = *resp

		return nil
	}
}

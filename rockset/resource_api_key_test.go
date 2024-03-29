package rockset

import (
	"errors"
	"fmt"
	rockerr "github.com/rockset/rockset-go-client/errors"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/rockset/rockset-go-client"
	"github.com/rockset/rockset-go-client/openapi"
	"github.com/rockset/rockset-go-client/option"
)

func TestAccApiKey_Basic(t *testing.T) {
	var apiKey openapi.ApiKey
	var keyValueOnCreation string

	v1 := Values{Name: randomName("apikey")}
	v2 := Values{Name: randomName("apikey")}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRocksetApiKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: getHCLTemplate("apikey_basic.tf", v1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRocksetApiKeyExists("rockset_api_key.test", &apiKey),
					resource.TestCheckResourceAttr("rockset_api_key.test", "name", v1.Name),
					resource.TestCheckNoResourceAttr("rockset_api_key.test", "user"),
					resource.TestCheckResourceAttrSet("rockset_api_key.test", "key"),
					// store the created key for comparison in another test later
					resource.TestCheckResourceAttrWith("rockset_api_key.test", "key", func(value string) error {
						keyValueOnCreation = value
						return nil
					}),
				),
				ExpectNonEmptyPlan: false,
			},
			{
				// Re-apply the same configuration, to verify that nothing changes.
				Config: getHCLTemplate("apikey_basic.tf", v1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("rockset_api_key.test", "key"),
					// Check that key in the stored resource has not changed its value.
					// We perform this specific test because the GET api key endpoint returns the key
					// value, but it is obfuscated and storing that might lead to unexpected results
					// if the user then writes that obfuscated value somewhere else like in a key vault.
					resource.TestCheckResourceAttrWith("rockset_api_key.test", "key", func(value string) error {
						if value != keyValueOnCreation {
							return fmt.Errorf("key has changed since creation, that likely means it is being overriden in the state with the obfuscated value")
						}
						return nil
					}),
				),
				ExpectNonEmptyPlan: false,
			},
			{
				Config: getHCLTemplate("apikey_basic.tf", v2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRocksetApiKeyExists("rockset_api_key.test", &apiKey),
					resource.TestCheckResourceAttr("rockset_api_key.test", "name", v2.Name),
					resource.TestCheckNoResourceAttr("rockset_api_key.test", "user"),
					resource.TestCheckResourceAttrSet("rockset_api_key.test", "key"),
				),
				ExpectNonEmptyPlan: false,
			},
			{
				// Back to basic, will change name AND api key
				Config: getHCLTemplate("apikey_basic.tf", v1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRocksetApiKeyExists("rockset_api_key.test", &apiKey),
					resource.TestCheckResourceAttr("rockset_api_key.test", "name", v1.Name),
					resource.TestCheckNoResourceAttr("rockset_api_key.test", "user"),
					resource.TestCheckResourceAttrSet("rockset_api_key.test", "key"),
				),
				ExpectNonEmptyPlan: false,
			},
		},
	})
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

		var options []option.APIKeyOption
		if user != "" {
			options = append(options, option.ForUser(user))
		}
		_, err = rc.GetAPIKey(testCtx, name, options...)

		// An error would mean we didn't find the key, we expect an error
		if err == nil {
			// We did not get an error, so we failed to delete the key.
			return fmt.Errorf("api Key %s still exists", rs.Primary.ID)
		}

		var re rockerr.Error
		if errors.As(err, &re) {
			if re.IsNotFoundError() {
				continue
			}
		}
		return err
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

		var options []option.APIKeyOption
		if user != "" {
			options = append(options, option.ForUser(user))
		}
		resp, err := rc.GetAPIKey(testCtx, name, options...)
		if err != nil {
			return err
		}

		*apiKey = resp

		return nil
	}
}

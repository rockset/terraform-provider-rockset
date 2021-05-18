package rockset

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var testAccProviders map[string]*schema.Provider
var testAccProvider *schema.Provider

func init() {
	testAccProvider = Provider()
	testAccProviders = map[string]*schema.Provider{
		"rockset": testAccProvider,
	}
}

func TestProvider(t *testing.T) {
	// InternalValidate should be called to validate the structure
	// of the provider.
	// This should be called in a unit test for any provider to verify
	// before release
	if err := Provider().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("ROCKSET_APIKEY"); v == "" {
		t.Fatal("ROCKSET_APIKEY must be set for acceptance tests")
	}
	if v := os.Getenv("ROCKSET_APISERVER"); v == "" {
		t.Fatal("ROCKSET_APISERVER must be set for acceptance tests")
	}
}

func getResourceFromState(state *terraform.State, resource string) (*terraform.ResourceState, error) {
	rs, ok := state.RootModule().Resources[resource]
	if !ok {
		return rs, fmt.Errorf("Not found: %s", resource)
	}
	if rs.Primary.ID == "" {
		return rs, fmt.Errorf("No Record ID is set")
	}

	return rs, nil
}

func testAccCheckResourceListMatches(resource string, resourceAttribute string, attributeValueCheck []string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, err := getResourceFromState(state, resource)
		if err != nil {
			return err
		}

		// Lists show up as the attribute name and index in one key.
		// E.g. roles=["foo", "bar"] will be {"roles.0"="foo", "roles.1"="bar"}
		// So we build a list of the values.
		var currentValue []string
		for k, v := range rs.Primary.Attributes {
			// Attributes will have a final string with the number in the list
			// E.g. roles.0=foo, roles.1=bar, roles.#=2
			// Skip the #
			if strings.HasPrefix(k, fmt.Sprintf("%s.", resourceAttribute)) &&
				!strings.HasPrefix(k, fmt.Sprintf("%s.#", resourceAttribute)) {
				currentValue = append(currentValue, v)
			}
		}

		if reflect.DeepEqual(currentValue, attributeValueCheck) {
			return nil
		} else {
			return fmt.Errorf("Resource attribute %s.%s did not match expected value. Found %s expected %s.",
				resource, resourceAttribute, currentValue, attributeValueCheck)
		}
	}
}

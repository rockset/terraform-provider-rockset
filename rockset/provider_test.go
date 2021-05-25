package rockset

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

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

func getFileContents(path string) (string, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

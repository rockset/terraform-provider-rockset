package rockset

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/rs/zerolog"
)

var testAccProviders map[string]*schema.Provider
var testAccProvider *schema.Provider
var testCtx context.Context

func init() {
	testAccProvider = Provider()
	testAccProviders = map[string]*schema.Provider{
		"rockset": testAccProvider,
	}

	testCtx = createTestContext()
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

/*
	Verifies necessary environment variables are set before running tests.
	Fails early if they are not set.
*/
func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("ROCKSET_APIKEY"); v == "" {
		t.Fatal("ROCKSET_APIKEY must be set for acceptance tests")
	}
	if v := os.Getenv("ROCKSET_APISERVER"); v == "" {
		t.Fatal("ROCKSET_APISERVER must be set for acceptance tests")
	}
}

/*
	Finds a specific resource by identifier (e.g. rockset_collection.test) from terraform state
	and returns the resource state.
*/
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

/*
	Gets a file's contents and returns them as a string.
	Intended to be used for test data.
*/
func getFileContents(path string) (string, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

/*
	Creates a context with debug logging for use in tests.
*/
func createTestContext() context.Context {
	console := zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}
	log := zerolog.New(console).Level(zerolog.TraceLevel).With().Timestamp().Logger()

	return log.WithContext(context.Background())
}

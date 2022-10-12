package rockset

import (
	"bytes"
	"context"
	"fmt"
	"github.com/rockset/rockset-go-client"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"
	"text/template"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/rs/zerolog"
)

var testAccProviderFactories map[string]func() (*schema.Provider, error)

var testAccProvider *schema.Provider
var testCtx context.Context

func init() {
	testAccProvider = Provider()
	testAccProviderFactories = map[string]func() (*schema.Provider, error){
		"rockset": func() (*schema.Provider, error) {
			return testAccProvider, nil
		},
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

// testAccPreCheck verifies required environment variables are set before running tests.
// It always requires ROCKSET_APIKEY and ROCKSET_APISERVER.
// Fails early if they are not set.
func testAccPreCheck(t *testing.T, env ...string) {
	env = append(env, "ROCKSET_APIKEY", "ROCKSET_APISERVER")
	for _, e := range env {
		if _, found := os.LookupEnv(e); !found {
			t.Fatalf("%s must be set for acceptance tests", e)
		}
	}
}

/*
	Finds a specific resource by identifier (e.g. rockset_collection.test) from terraform state
	and returns the resource state.
*/
func getResourceFromState(state *terraform.State, resource string) (*terraform.ResourceState, error) {
	rs, ok := state.RootModule().Resources[resource]
	if !ok {
		return rs, fmt.Errorf("not found: %s", resource)
	}
	if rs.Primary.ID == "" {
		return rs, fmt.Errorf("no Record ID is set")
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

func getHCL(filename string) string {
	hclPath := filepath.Join("..", "testdata", filename)
	hcl, err := getFileContents(hclPath)
	if err != nil {
		log.Fatalf("Unexpected error loading test data %s", hclPath)
	}

	return hcl
}

// getHCLTemplate returns a rendered HCL config
func getHCLTemplate(filename string, data any) string {
	hclPath := filepath.Join("..", "testdata", filename)
	hcl, err := getFileContents(hclPath)
	if err != nil {
		log.Fatalf("Unexpected error loading test data %s", hclPath)
	}

	var buf []byte
	rendered := bytes.NewBuffer(buf)
	t := template.Must(template.New("hcl").Parse(hcl))
	if err = t.Execute(rendered, data); err != nil {
		log.Fatalf("failed to render %s: %v", filename, err)
	}

	return rendered.String()
}

/*
	Creates a context with debug logging for use in tests.
*/
func createTestContext() context.Context {
	console := zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}
	l := zerolog.New(console).Level(zerolog.TraceLevel).With().Timestamp().Logger()

	return l.WithContext(context.Background())
}

// testAccCheckRocksetIntegrationDestroy checks that an integration has been destroyed
func testAccCheckRocksetIntegrationDestroy(resource string) func(*terraform.State) error {
	return func(s *terraform.State) error {
		rc := testAccProvider.Meta().(*rockset.RockClient)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != resource {
				continue
			}

			name := rs.Primary.ID
			_, err := rc.GetIntegration(testCtx, name)
			// An error would mean we didn't find it, we expect an error
			if err == nil {
				return err
			}
		}

		return nil
	}
}

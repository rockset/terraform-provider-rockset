package rockset

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"text/template"
	"time"

	"github.com/rockset/rockset-go-client"
	rockerr "github.com/rockset/rockset-go-client/errors"
	"github.com/rockset/rockset-go-client/openapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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

func TestToBoolPtrNilIfEmpty_Nil(t *testing.T) {
	var v interface{}
	var expected *bool
	actual := toBoolPtrNilIfEmpty(v)
	assert.Equal(t, expected, actual)
}

func TestToBoolPtrNilIfEmpty_True(t *testing.T) {
	var v interface{}
	v = true
	expected := true
	actual := toBoolPtrNilIfEmpty(v)
	assert.Equal(t, &expected, actual)
}

func TestToBoolPtrNilIfEmpty_False(t *testing.T) {
	var v interface{}
	v = false
	expected := false
	actual := toBoolPtrNilIfEmpty(v)
	assert.Equal(t, &expected, actual)
}

func TestToBoolPtrNilIfEmpty_Error(t *testing.T) {
	var v interface{}
	v = "string"
	assert.Panics(t, assert.PanicTestFunc(func() {
		toBoolPtrNilIfEmpty(v)
	}))
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
	return getHCLTemplateWithFn(filename, data, nil)
}

func getHCLTemplateWithFn(filename string, data any, funcMap template.FuncMap) string {
	hclPath := filepath.Join("..", "testdata", filename)
	hcl, err := getFileContents(hclPath)
	if err != nil {
		log.Fatalf("Unexpected error loading test data %s", hclPath)
	}

	var buf []byte
	rendered := bytes.NewBuffer(buf)
	t := template.New("hcl")
	if funcMap != nil {
		t = t.Funcs(funcMap)
	}
	t = template.Must(t.Parse(hcl))

	if err = t.Execute(rendered, data); err != nil {
		log.Fatalf("failed to render %s: %v", filename, err)
	}

	return rendered.String()
}

const buildNum = "CIRCLE_BUILD_NUM"
const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))

// StringWithCharset creates a random string with length and charset
func stringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

// String creates a random string with length
func randomString(length int) string {
	return stringWithCharset(length, charset)
}

func randomName(prefix string) string {
	num, found := os.LookupEnv(buildNum)
	if !found {
		if user, found := os.LookupEnv("USER"); found {
			num = strings.ToLower(user)
		} else {
			num = "dev"
		}
	}

	return fmt.Sprintf("tf_%s_%s_%s", num, prefix, randomString(6))
}

func description() string {
	num, found := os.LookupEnv(buildNum)
	if !found {
		num = "dev"
	}
	return fmt.Sprintf("created by terraform integration test run %s", num)
}

type Values struct {
	Name                   string
	Alias                  string
	Collection             string
	Description            string
	Workspace              string
	SQL                    string
	Email                  string
	FirstName              string
	LastName               string
	Roles                  []string
	Role                   string
	Bucket                 string
	Tag                    string
	Retention              int
	IngestTransformation   string
	CreateTimeout          string
	UseScanApi             *bool
	RCU                    *int
	StorageCompressionType string
}

const S3IntegrationRoleArn = "arn:aws:iam::469279130686:role/terraform-provider-rockset-tests"

// createTestContext creates a context with debug logging for use in tests.
func createTestContext() context.Context {
	lvl := zerolog.WarnLevel
	if os.Getenv("ROCKSET_DEBUG") != "" {
		lvl = zerolog.TraceLevel
	}
	console := zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}
	l := zerolog.New(console).Level(lvl).With().Timestamp().Logger()

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

func TestDiagFromErr(t *testing.T) {
	err := rockerr.NewWithStatusCode(errors.New("apierr"), &http.Response{StatusCode: http.StatusConflict})

	var re rockerr.Error
	require.True(t, errors.As(err, &re))

	re.ErrorModel = &openapi.ErrorModel{
		Message: openapi.PtrString("api error"),
		TraceId: openapi.PtrString("foobar"),
		QueryId: openapi.PtrString("queryid"),
		ErrorId: openapi.PtrString("errorid"),
		Line:    openapi.PtrInt32(42),
		Column:  openapi.PtrInt32(42),
	}

	tests := []struct {
		name string
		err  error
		want string
	}{
		{"plain error", errors.New("plain error"), ""},
		{
			"rockset error with http code",
			re,
			"api error: HTTP status code (409) Conflict, Trace ID: foobar, Error ID: errorid, " +
				"Query ID: queryid, Line: 42, Column: 42",
		},
	}
	for _, tst := range tests {
		t.Run(tst.name, func(t *testing.T) {
			d := DiagFromErr(tst.err)
			require.Len(t, d, 1)
			assert.Equal(t, tst.want, d[0].Detail)
		})
	}
}

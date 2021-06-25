package rockset

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/rockset/rockset-go-client"
)

type Config struct {
	APIKey    string
	APIServer string
}

func Provider() *schema.Provider {
	return &schema.Provider{
		ResourcesMap: map[string]*schema.Resource{
			"rockset_alias":          resourceAlias(),
			"rockset_api_key":        resourceApiKey(),
			"rockset_collection":     resourceCollection(),
			"rockset_query_lambda":   resourceQueryLambda(),
			"rockset_s3_collection":  resourceS3Collection(),
			"rockset_s3_integration": resourceS3Integration(),
			"rockset_user":           resourceUser(),
			"rockset_workspace":      resourceWorkspace(),
			// "rockset_query_lambda": resourceQueryLambda(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"rockset_account": dataSourceRocksetAccount(),
		},
		Schema: map[string]*schema.Schema{
			"api_key": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "The API key used to access Rockset",
				Sensitive:   true,
			},
			"api_server": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "api.rs2.usw2.rockset.com",
				Description: "The API server for accessing Rockset",
			},
		},
		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	config := Config{
		APIKey:    d.Get("api_key").(string),
		APIServer: d.Get("api_server").(string),
	}

	return config.Client()
}

func (c *Config) Client() (interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics
	var opts []rockset.RockOption

	if c.APIKey != "" {
		opts = append(opts, rockset.WithAPIKey(c.APIKey), rockset.WithAPIServer(c.APIServer))
	}

	rc, err := rockset.NewClient(opts...)
	if err != nil {
		return nil, diag.FromErr(err)
	}

	return rc, diags
}

var nameRegexp = regexp.MustCompile("^[[:alnum:]][[:alnum:]-_]*$")

func rocksetNameValidator(val interface{}, key string) ([]string, []error) {
	s := val.(string)
	if nameRegexp.MatchString(s) {
		return nil, nil
	}
	return nil, []error{fmt.Errorf("%s must start with alphanumeric, the rest can be alphanumeric, -, or _", key)}
}

/*
	Returns an id of format <workspace>.<collection>.
	This is how collections are referenced in Rockset.
*/
func toID(workspace, name string) string {
	// The provider will be configured for 1 account.
	// This should be universally unique within the account.
	return fmt.Sprintf("%s.%s", workspace, name)
}

func workspaceAndNameFromID(id string) (string, string) {
	tokens := strings.SplitN(id, ".", 2)
	if len(tokens) != 2 {
		log.Printf("unparsable id: %s", id)
		return "", ""
	}
	return tokens[0], tokens[1]
}

// convert an array of interface{} to an array of string
func toStringArray(a []interface{}) []string {
	r := make([]string, len(a))
	for i, v := range a {
		r[i] = v.(string)
	}
	return r
}

/*

 */
func toStringPtrNilIfEmpty(v string) *string {
	if v == "" {
		return nil
	}

	return &v
}

func toStringArrayPtr(v []string) *[]string {
	return &v
}

func mergeSchemas(mergeOnto map[string]*schema.Schema, toMerge map[string]*schema.Schema) map[string]*schema.Schema {
	for k, v := range toMerge {
		mergeOnto[k] = v
	}

	return mergeOnto
}

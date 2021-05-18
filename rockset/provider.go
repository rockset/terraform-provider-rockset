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
			"rockset_user":           resourceUser(),
			"rockset_s3_integration": resourceS3Integration(),
			// "rockset_workspace":      resourceWorkspace(),
			// "rockset_s3_collection":  resourceS3Collection(),
			// "rockset_query_lambda": resourceQueryLambda(),
			// "rockset_collection": resourceCollection(),
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
		// TODO: WithAPIKey no longer in go client, re-add this once it's added back
		//opts = append(opts, rockset.WithAPIKey(c.APIKey), rockset.WithAPIServer(c.APIServer))
	} else {
		opts = append(opts, rockset.FromEnv())
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

func toID(workspace, name string) string {
	// TODO if there are multiple accounts which all have the same workspace and name
	// this ID wont't work, so perhaps the account name should be included in the id?
	return fmt.Sprintf("%s:%s", workspace, name)
}

func workspaceAndNameFromID(id string) (string, string) {
	tokens := strings.SplitN(id, ":", 2)
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

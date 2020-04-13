package rockset

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/rockset/rockset-go-client"
	"regexp"
)

type Config struct {
	APIKey    string
	APIServer string
}

func Provider() *schema.Provider {
	provider := &schema.Provider{
		ResourcesMap: map[string]*schema.Resource{
			"rockset_workspace": resourceWorkspace(),
			"rockset_s3_integration": resourceS3Integration(),
			"rockset_s3_collection":  resourceS3Collection(),
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
	}

	provider.ConfigureFunc = func(d *schema.ResourceData) (interface{}, error) {
		config := Config{
			APIKey:    d.Get("api_key").(string),
			APIServer: d.Get("api_server").(string),
		}

		return config.Client()
	}

	return provider
}

func (c *Config) Client() (interface{}, error) {
	var opts []rockset.RockOption

	if c.APIKey != "" {
		opts = append(opts, rockset.WithAPIKey(c.APIKey), rockset.WithAPIServer(c.APIServer))
	} else {
		opts = append(opts, rockset.FromEnv())
	}

	return rockset.NewClient(opts...)
}

var nameRegexp = regexp.MustCompile("^[[:alnum:]][[:alnum:]-_]*$")

func rocksetNameValidator(val interface{}, key string) ([]string, []error) {
	s := val.(string)
	if nameRegexp.MatchString(s) {
		return nil, nil
	}
	return nil, []error{fmt.Errorf("%s must start with alphanumeric, the rest can be alphanumeric, -, or _", key)}
}

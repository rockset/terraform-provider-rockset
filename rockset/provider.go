package rockset

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/rockset/rockset-go-client"
	rockerr "github.com/rockset/rockset-go-client/errors"
)

type Config struct {
	APIKey    string
	APIServer string
	OrgID     string
}

func Provider() *schema.Provider {
	schema.DescriptionKind = schema.StringMarkdown
	return &schema.Provider{
		ResourcesMap: map[string]*schema.Resource{
			"rockset_alias":                 resourceAlias(),
			"rockset_api_key":               resourceApiKey(),
			"rockset_autoscaling_policy":    resourceAutoScalingPolicy(),
			"rockset_collection":            resourceCollection(),
			"rockset_collection_mount":      resourceCollectionMount(),
			"rockset_dynamodb_collection":   resourceDynamoDBCollection(),
			"rockset_dynamodb_integration":  resourceDynamoDBIntegration(),
			"rockset_gcs_collection":        resourceGCSCollection(),
			"rockset_gcs_integration":       resourceGCSIntegration(),
			"rockset_kafka_collection":      resourceKafkaCollection(),
			"rockset_kafka_integration":     resourceKafkaIntegration(),
			"rockset_kinesis_collection":    resourceKinesisCollection(),
			"rockset_kinesis_integration":   resourceKinesisIntegration(),
			"rockset_mongodb_collection":    resourceMongoDBCollection(),
			"rockset_mongodb_integration":   resourceMongoDBIntegration(),
			"rockset_query_lambda":          resourceQueryLambda(),
			"rockset_query_lambda_tag":      resourceQueryLambdaTag(),
			"rockset_query_lambda_auto_tag": resourceQueryLambdaAutoTag(),
			"rockset_role":                  resourceRole(),
			"rockset_s3_collection":         resourceS3Collection(),
			"rockset_s3_integration":        resourceS3Integration(),
			"rockset_user":                  resourceUser(),
			"rockset_view":                  resourceView(),
			"rockset_virtual_instance":      resourceVirtualInstance(),
			"rockset_workspace":             resourceWorkspace(),
			"rockset_scheduled_lambda":      resourceScheduledLambda(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"rockset_account":          dataSourceRocksetAccount(),
			"rockset_query_lambda":     dataSourceRocksetQueryLambda(),
			"rockset_query_lambda_tag": dataSourceRocksetQueryLambdaTag(),
			"rockset_user":             dataSourceRocksetUser(),
			"rockset_virtual_instance": dataSourceRocksetVirtualInstance(),
			"rockset_workspace":        dataSourceRocksetWorkspace(),
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
				Default:     "",
				Description: "The API server for accessing Rockset",
			},
			"organization_id": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
				Description: "The ID of the organization to connect to. " +
					"If this is set, the provider will validate that the organization_id matches the organization_id " +
					"of the api key. If it does not match, the provider will return an error.\n",
			},
		},
		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(_ context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	config := Config{
		APIKey:    d.Get("api_key").(string),
		APIServer: d.Get("api_server").(string),
		OrgID:     d.Get("organization_id").(string),
	}

	return config.Client()
}

const providerUserAgent = "terraform-provider-rockset"

func (c *Config) Client() (interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics
	var opts = []rockset.RockOption{
		rockset.WithUserAgent(fmt.Sprintf("%s/%s", providerUserAgent, Version)),
	}

	if c.APIKey != "" {
		opts = append(opts, rockset.WithAPIKey(c.APIKey))
	}

	if c.APIServer != "" {
		opts = append(opts, rockset.WithAPIServer(c.APIServer))
	}

	if debug := os.Getenv("ROCKSET_DEBUG"); debug == "true" {
		opts = append(opts, rockset.WithHTTPDebug())
	}

	rc, err := rockset.NewClient(opts...)
	if err != nil {
		return nil, DiagFromErr(err)
	}

	// if we have an org id in the config, validate that it matches the org id of the api key
	if c.OrgID != "" {
		org, err := rc.GetOrganization(context.Background())
		if err != nil {
			return nil, DiagFromErr(err)
		}

		if org.GetId() != c.OrgID {
			return nil, diag.Errorf(
				"the organization configured in the provider `%s` does not match the organization of the api key: `%s`",
				c.OrgID, org.GetId())
		}
	}

	return rc, diags
}

const nameRe = "[[:alnum:]][[:alnum:]-_]*"

var nameRegexp = regexp.MustCompile(fmt.Sprintf("^%s$", nameRe))

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
	// TODO refactor this call to return the error
	ws, name, _ := split(id, ".")
	return ws, name
}

func mountToID(collectionPath, id string) string {
	return collectionPath + ":" + id
}

func idToMount(id string) (string, string, error) {
	return split(id, ":")
}

func split(id, sep string) (string, string, error) {
	tokens := strings.SplitN(id, sep, 2)
	if len(tokens) != 2 {
		return "", "", fmt.Errorf("could not locate separator %s in %s", sep, id)
	}
	return tokens[0], tokens[1], nil
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

func toBoolPtrNilIfEmpty(v interface{}) *bool {
	var res *bool
	if v == nil {
		return res
	} else {
		vB := v.(bool)
		res = &vB
	}
	return res
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

// checkForNotFoundError check is the error is a Rockset NotFoundError, and then clears the id which makes
// terraform create the resource, but if it isn't a NotFoundError it will return the error wrapped in diag.Diagnostics
func checkForNotFoundError(d *schema.ResourceData, err error) diag.Diagnostics {
	var re rockerr.Error
	if !errors.As(err, &re) {
		return DiagFromErr(err)
	}

	if !re.IsNotFoundError() {
		return DiagFromErr(err)
	}

	d.SetId("")

	return diag.Diagnostics{}
}

func DiagFromErr(err error) diag.Diagnostics {
	if err == nil {
		return nil
	}
	d := diag.Diagnostic{
		Severity: diag.Error,
		Summary:  err.Error(),
	}

	var re rockerr.Error
	if errors.As(err, &re) {
		var sb strings.Builder
		var msgs []string

		sb.WriteString(re.GetMessage())
		sb.WriteString(": ")

		if t, ok := re.GetTypeOk(); ok {
			msgs = append(msgs, fmt.Sprintf("Error Type: %s", *t))
		}
		if re.StatusCode != 0 {
			msgs = append(msgs, fmt.Sprintf("HTTP status code (%d) %s", re.StatusCode, http.StatusText(re.StatusCode)))
		}
		if re.GetTraceId() != "" {
			msgs = append(msgs, fmt.Sprintf("Trace ID: %s", re.GetTraceId()))
		}
		if re.GetErrorId() != "" {
			msgs = append(msgs, fmt.Sprintf("Error ID: %s", re.GetErrorId()))
		}
		if re.GetQueryId() != "" {
			msgs = append(msgs, fmt.Sprintf("Query ID: %s", re.GetQueryId()))
		}
		if re.HasLine() {
			msgs = append(msgs, fmt.Sprintf("Line: %d", re.GetLine()))
		}
		if re.HasColumn() {
			msgs = append(msgs, fmt.Sprintf("Column: %d", re.GetColumn()))
		}

		sb.WriteString(strings.Join(msgs, ", "))

		d.Detail = sb.String()
	}

	return diag.Diagnostics{d}
}

package rockset

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/rockset/rockset-go-client"
	"github.com/rockset/rockset-go-client/openapi"
)

func kinesisCollectionSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"source": {
			Description: "Defines a source for this collection.",
			Type:        schema.TypeSet,
			ForceNew:    true,
			Optional:    true,
			MinItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"integration_name": {
						Description:  "The name of the Rockset Kinesis integration.",
						Type:         schema.TypeString,
						ForceNew:     true,
						Required:     true,
						ValidateFunc: rocksetNameValidator,
					},
					"stream_name": {
						Description: "Name of Kinesis stream.",
						Type:        schema.TypeString,
						ForceNew:    true,
						Required:    true,
					},
					"aws_region": {
						Description: "AWS region name for the Kinesis stream, by default us-west-2 is used",
						Type:        schema.TypeString,
						ForceNew:    true,
						Optional:    true,
						Default:     "us-west-2",
					},
					"format": {
						Type:     schema.TypeString,
						ForceNew: true,
						Required: true,
						ValidateFunc: validation.StringMatch(
							regexp.MustCompile("^(json|mysql|postgres)$"), "only 'json', 'mysql', or 'postgres' is supported"),
						Description: "Format of the data. One of: json, mysql, postgres. dms_primary_keys list can only be set for mysql or postgres. ",
					},
					"dms_primary_key": {
						Description: "Set of fields that correspond to a DMS primary key. Can only be set if format is mysql or postgres.",
						Type:        schema.TypeList,
						ForceNew:    true,
						Optional:    true,
						MinItems:    1,
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
					},
				},
			},
		},
	} // End schema return
} // End func

func resourceKinesisCollection() *schema.Resource {
	return &schema.Resource{
		Description: "Manages a collection with an Kinesis source attached.",

		CreateContext: resourceKinesisCollectionCreate,
		ReadContext:   resourceKinesisCollectionRead,
		DeleteContext: resourceCollectionDelete, // No change from base collection delete

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		// This schema will use the base collection schema as a foundation
		// And layer on just the necessary fields for a Kinesis collection
		Schema: mergeSchemas(baseCollectionSchema(), kinesisCollectionSchema()),
	}
}

func resourceKinesisCollectionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rc := meta.(*rockset.RockClient)
	var diags diag.Diagnostics
	var err error

	name := d.Get("name").(string)
	workspace := d.Get("workspace").(string)

	// Add all base schema fields
	params := createBaseCollectionRequest(d)
	// Add fields for Kinesis
	params.Sources = makeKinesisSourceParams(d.Get("source"))

	_, err = rc.CreateCollection(ctx, workspace, name, params)
	if err != nil {
		return diag.FromErr(err)
	}

	err = rc.WaitUntilCollectionReady(ctx, workspace, name)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(toID(workspace, name))

	return diags
}

func resourceKinesisCollectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rc := meta.(*rockset.RockClient)
	var diags diag.Diagnostics
	var err error

	workspace, name := workspaceAndNameFromID(d.Id())

	collection, err := rc.GetCollection(ctx, workspace, name)
	if err != nil {
		return diag.FromErr(err)
	}

	// Gets all the fields any generic collection has
	err = parseBaseCollection(&collection, d)
	if err != nil {
		return diag.FromErr(err)
	}

	// Gets all the fields relevant to a Kinesis collection
	err = parseKinesisCollection(&collection, d)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

/*
	Takes in a collection returned from the api.
	Parses the fields relevant to an Kinesis source and
	puts them into the schema object.
*/
func parseKinesisCollection(collection *openapi.Collection, d *schema.ResourceData) error {

	var err error

	sourcesList := *collection.Sources
	sourcesCount := len(sourcesList)
	if sourcesCount < 1 {
		return fmt.Errorf("expected %s to have at least 1 source", collection.GetName())
	}

	err = d.Set("source", flattenKinesisSourceParams(&sourcesList))
	if err != nil {
		return err
	}

	return nil // No errors
}

func flattenKinesisSourceParams(sources *[]openapi.Source) []interface{} {
	convertedList := make([]interface{}, 0, len(*sources))
	for _, source := range *sources {
		m := make(map[string]interface{})
		m["integration_name"] = source.IntegrationName
		m["stream_name"] = source.Kinesis.StreamName
		m["aws_region"] = source.Kinesis.GetAwsRegion()

		isJson, jsonOk := source.FormatParams.GetJsonOk()
		isMysql, mysqlOk := source.FormatParams.GetMysqlDmsOk()
		isPostgres, postgresOk := source.FormatParams.GetPostgresDmsOk()
		if jsonOk && *isJson {
			m["format"] = "json"
		} else if mysqlOk && *isMysql {
			m["format"] = "mysql"
		} else if postgresOk && *isPostgres {
			m["format"] = "postgres"
		} else {
			// There's a bug in the API currently
			// format_params is null if the format is JSON
			// TODO: Once fixed, this else path should be an error
			m["format"] = "json"
		}

		m["dms_primary_key"] = source.Kinesis.GetDmsPrimaryKey()

		convertedList = append(convertedList, m)
	}

	return convertedList
}

func makeKinesisSourceParams(in interface{}) *[]openapi.Source {
	sources := make([]openapi.Source, 0, in.(*schema.Set).Len())

	for _, i := range in.(*schema.Set).List() {
		if val, ok := i.(map[string]interface{}); ok {
			source := openapi.Source{}
			source.Kinesis = openapi.NewSourceKinesisWithDefaults()
			format := openapi.FormatParams{}
			source.FormatParams = &format
			// API returns an empty list if this isn't set...
			// So we have to default it to an empty list here.
			source.Kinesis.DmsPrimaryKey = toStringArrayPtr(make([]string, 0))

			for k, v := range val {
				switch k {
				case "integration_name":
					source.IntegrationName = v.(string)
				case "stream_name":
					source.Kinesis.StreamName = v.(string)
				case "aws_region":
					source.Kinesis.AwsRegion = toStringPtrNilIfEmpty(v.(string))
				case "format":
					formatType := v.(string)
					switch formatType {
					case "json":
						source.FormatParams.Json = openapi.PtrBool(true)
					case "mysql":
						source.FormatParams.MysqlDms = openapi.PtrBool(true)
					case "postgres":
						source.FormatParams.PostgresDms = openapi.PtrBool(true)
					}
				case "dms_primary_key":
					source.Kinesis.DmsPrimaryKey = toStringArrayPtr(toStringArray(v.([]interface{})))
				}
			}
			sources = append(sources, source)
		}
	}

	return &sources
}

package rockset

import (
	"context"
	"fmt"
	"github.com/rockset/rockset-go-client/option"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/rockset/rockset-go-client"
	"github.com/rockset/rockset-go-client/openapi"
)

func dynamoDBCollectionSchema() map[string]*schema.Schema {
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
						Description:  "The name of the Rockset DynamoDB integration.",
						Type:         schema.TypeString,
						ForceNew:     true,
						Required:     true,
						ValidateFunc: rocksetNameValidator,
					},
					"table_name": {
						Description: "Name of DynamoDB table containing data.",
						Type:        schema.TypeString,
						ForceNew:    true,
						Required:    true,
					},
					"aws_region": {
						Description: "AWS region name of DynamoDB table, by default us-west-2 is used.",
						Type:        schema.TypeString,
						ForceNew:    true,
						Optional:    true,
					},
					"rcu": {
						Description: "Max RCU usage for scan.",
						Type:        schema.TypeInt,
						ForceNew:    true,
						Optional:    true,
					},
					"scan_start_time": {
						Description: "DynamoDB scan start time.",
						Type:        schema.TypeString,
						Computed:    true,
					},
					"scan_end_time": {
						Description: "DynamoDB scan end time.",
						Type:        schema.TypeString,
						Computed:    true,
					},
					"scan_records_processed": {
						Description: "Number of records inserted using scan.",
						Type:        schema.TypeInt,
						Computed:    true,
					},
					"scan_total_records": {
						Description: "Number of records in DynamoDB table at time of scan.",
						Type:        schema.TypeInt,
						Computed:    true,
					},
					"state": {
						Description: "State of current ingest for this table.",
						Type:        schema.TypeString,
						Computed:    true,
					},
					"stream_last_processed_at": {
						Description: "ISO-8601 date when source was last processed.",
						Type:        schema.TypeString,
						Computed:    true,
					},
					"use_scan_api": {
						Description: "Whether the initial table scan should use the DynamoDB scan API. If false, export will be performed using an S3 bucket.",
						Type:        schema.TypeBool,
						Optional:    true,
						Default:     true,
					},
				},
			},
		},
	} // End schema return
} // End func

func resourceDynamoDBCollection() *schema.Resource {
	return &schema.Resource{
		Description: "Manages a collection with an DynamoDB source attached.",

		CreateContext: resourceDynamoDBCollectionCreate,
		ReadContext:   resourceDynamoDBCollectionRead,
		DeleteContext: resourceCollectionDelete, // No change from base collection delete

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		// This schema will use the base collection schema as a foundation
		// And layer on just the necessary fields for a dynamodb collection
		Schema: mergeSchemas(baseCollectionSchema(), dynamoDBCollectionSchema()),
	}
}

func resourceDynamoDBCollectionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rc := meta.(*rockset.RockClient)
	var diags diag.Diagnostics
	var err error

	name := d.Get("name").(string)
	workspace := d.Get("workspace").(string)

	// Add all base schema fields
	params := createBaseCollectionRequest(d)
	// Add fields for DynamoDB
	params.Sources = makeSourceParams(d.Get("source"))

	_, err = rc.CreateCollection(ctx, workspace, name, option.WithCollectionRequest(*params))
	if err != nil {
		return diag.FromErr(err)
	}

	if err = waitForCollectionAndDocuments(ctx, rc, d, workspace, name); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(toID(workspace, name))

	return diags
}

func resourceDynamoDBCollectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	// Gets all the fields relevant to a DynamoDB collection
	err = parseDynamoDBCollection(&collection, d)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

/*
Takes in a collection returned from the api.
Parses the fields relevant to an DynamoDB source and
puts them into the schema object.
*/
func parseDynamoDBCollection(collection *openapi.Collection, d *schema.ResourceData) error {

	var err error

	sourcesList := collection.Sources
	sourcesCount := len(sourcesList)
	if sourcesCount < 1 {
		return fmt.Errorf("expected %s to have at least 1 source", collection.GetName())
	}

	err = d.Set("source", flattenSourceParams(&sourcesList))
	if err != nil {
		return err
	}

	return nil // No errors
}

func flattenSourceParams(sources *[]openapi.Source) []interface{} {
	convertedList := make([]interface{}, 0, len(*sources))
	for _, source := range *sources {
		m := make(map[string]interface{})
		m["integration_name"] = source.IntegrationName
		m["table_name"] = source.Dynamodb.TableName
		m["aws_region"] = source.Dynamodb.GetAwsRegion()
		m["rcu"] = source.Dynamodb.GetRcu()
		m["scan_start_time"] = source.Dynamodb.Status.GetScanStartTime()
		m["scan_end_time"] = source.Dynamodb.Status.GetScanEndTime()
		m["scan_records_processed"] = source.Dynamodb.Status.GetScanRecordsProcessed()
		m["scan_total_records"] = source.Dynamodb.Status.GetScanTotalRecords()
		m["state"] = source.Dynamodb.Status.GetState()
		m["stream_last_processed_at"] = source.Dynamodb.Status.GetStreamLastProcessedAt()
		m["use_scan_api"] = source.Dynamodb.UseScanApi

		convertedList = append(convertedList, m)
	}

	return convertedList
}

func makeSourceParams(in interface{}) []openapi.Source {
	sources := make([]openapi.Source, 0, in.(*schema.Set).Len())

	for _, i := range in.(*schema.Set).List() {
		if val, ok := i.(map[string]interface{}); ok {
			source := openapi.Source{}
			source.Dynamodb = openapi.NewSourceDynamoDbWithDefaults()
			for k, v := range val {
				switch k {
				case "integration_name":
					source.IntegrationName = toStringPtrNilIfEmpty(v.(string))
				case "table_name":
					source.Dynamodb.TableName = v.(string)
				case "aws_region":
					source.Dynamodb.AwsRegion = toStringPtrNilIfEmpty(v.(string))
				case "rcu":
					source.Dynamodb.Rcu = openapi.PtrInt64(int64(v.(int)))
				case "use_scan_api":
					source.Dynamodb.UseScanApi = toBoolPtrNilIfEmpty(v)
				}
			}

			sources = append(sources, source)
		}
	}

	return sources
}

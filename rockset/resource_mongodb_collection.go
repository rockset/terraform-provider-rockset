package rockset

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/rockset/rockset-go-client"
	"github.com/rockset/rockset-go-client/openapi"
)

func mongoDBCollectionSchema() map[string]*schema.Schema {
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
						Description:  "The name of the Rockset MongoDB integration.",
						Type:         schema.TypeString,
						ForceNew:     true,
						Required:     true,
						ValidateFunc: rocksetNameValidator,
					},
					"database_name": {
						Description: "MongoDB database name containing the target collection.",
						Type:        schema.TypeString,
						ForceNew:    true,
						Required:    true,
					},
					"collection_name": {
						Description: "MongoDB collection name of the target collection.",
						Type:        schema.TypeString,
						ForceNew:    true,
						Required:    true,
					},
					"scan_start_time": {
						Description: "MongoDB scan start time.",
						Type:        schema.TypeString,
						Computed:    true,
					},
					"scan_end_time": {
						Description: "MongoDB scan end time.",
						Type:        schema.TypeString,
						Computed:    true,
					},
					"scan_records_processed": {
						Description: "Number of records inserted using scan.",
						Type:        schema.TypeInt,
						Computed:    true,
					},
					"scan_total_records": {
						Description: "Number of records in MongoDB table at time of scan.",
						Type:        schema.TypeInt,
						Computed:    true,
					},
					"state": {
						Description: "State of current ingest for this table.",
						Type:        schema.TypeString,
						Computed:    true,
					},
					"stream_last_insert_processed_at": {
						Description: "ISO-8601 date when new insert from source was last processed.",
						Type:        schema.TypeString,
						Computed:    true,
					},
					"stream_last_update_processed_at": {
						Description: "ISO-8601 date when update from source was last processed.",
						Type:        schema.TypeString,
						Computed:    true,
					},
					"stream_last_delete_processed_at": {
						Description: "ISO-8601 date when delete from source was last processed.",
						Type:        schema.TypeString,
						Computed:    true,
					},
					"stream_records_inserted": {
						Description: "Number of new records inserted using stream.",
						Type:        schema.TypeInt,
						Computed:    true,
					},
					"stream_records_updated": {
						Description: "Number of new records updated using stream.",
						Type:        schema.TypeInt,
						Computed:    true,
					},
					"stream_records_deleted": {
						Description: "Number of new records deleted using stream.",
						Type:        schema.TypeInt,
						Computed:    true,
					},
				},
			},
		},
	} // End schema return
} // End func

func resourceMongoDBCollection() *schema.Resource {
	return &schema.Resource{
		Description: "Manages a collection with an MongoDB source attached.",

		CreateContext: resourceMongoDBCollectionCreate,
		ReadContext:   resourceMongoDBCollectionRead,
		DeleteContext: resourceCollectionDelete, // No change from base collection delete

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		// This schema will use the base collection schema as a foundation
		// And layer on just the necessary fields for a MongoDB collection
		Schema: mergeSchemas(baseCollectionSchema(), mongoDBCollectionSchema()),
	}
}

func resourceMongoDBCollectionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rc := meta.(*rockset.RockClient)
	var diags diag.Diagnostics
	var err error

	name := d.Get("name").(string)
	workspace := d.Get("workspace").(string)

	// Add all base schema fields
	params := createBaseCollectionRequest(d)
	// Add fields for MongoDB
	addMongoDBParams(d, params)

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

func resourceMongoDBCollectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	// Gets all the fields relevant to a MongoDB collection
	err = parseMongoDBCollection(&collection, d)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

/*
	Takes in a collection returned from the api.
	Parses the fields relevant to an MongoDB source and
	puts them into the schema object.
*/
func parseMongoDBCollection(collection *openapi.Collection, d *schema.ResourceData) error {

	var err error

	sourcesList := *collection.Sources
	sourcesCount := len(sourcesList)
	if sourcesCount < 1 {
		return fmt.Errorf("expected %s to have at least 1 source", collection.GetName())
	}

	err = d.Set("source", flattenMongoDBSourceParams(&sourcesList))
	if err != nil {
		return err
	}

	return nil // No errors
}

/*
	Adds the MongoDB sources data to the create collection request
*/
func addMongoDBParams(d *schema.ResourceData, params *openapi.CreateCollectionRequest) {
	params.Sources = makeMongoDBSourceParams(d.Get("source"))
}

func flattenMongoDBSourceParams(sources *[]openapi.Source) []interface{} {
	convertedList := make([]interface{}, 0, len(*sources))
	for _, source := range *sources {
		m := make(map[string]interface{})
		m["integration_name"] = source.IntegrationName
		m["database_name"] = source.Mongodb.DatabaseName
		m["collection_name"] = source.Mongodb.CollectionName
		m["scan_start_time"] = source.Mongodb.Status.GetScanStartTime()
		m["scan_end_time"] = source.Mongodb.Status.GetScanEndTime()
		m["scan_records_processed"] = source.Mongodb.Status.GetScanRecordsProcessed()
		m["scan_total_records"] = source.Mongodb.Status.GetScanTotalRecords()
		m["state"] = source.Mongodb.Status.GetState()
		m["stream_last_insert_processed_at"] = source.Mongodb.Status.GetStreamLastInsertProcessedAt()
		m["stream_last_update_processed_at"] = source.Mongodb.Status.GetStreamLastUpdateProcessedAt()
		m["stream_last_delete_processed_at"] = source.Mongodb.Status.GetStreamLastDeleteProcessedAt()
		m["stream_records_inserted"] = source.Mongodb.Status.GetStreamRecordsInserted()
		m["stream_records_updated"] = source.Mongodb.Status.GetStreamRecordsUpdated()
		m["stream_records_deleted"] = source.Mongodb.Status.GetStreamRecordsDeleted()

		convertedList = append(convertedList, m)
	}

	return convertedList
}

func makeMongoDBSourceParams(in interface{}) *[]openapi.Source {
	sources := make([]openapi.Source, 0, in.(*schema.Set).Len())

	for _, i := range in.(*schema.Set).List() {
		if val, ok := i.(map[string]interface{}); ok {
			source := openapi.Source{}
			source.Mongodb = openapi.NewSourceMongoDbWithDefaults()
			for k, v := range val {
				switch k {
				case "integration_name":
					source.IntegrationName = v.(string)
				case "database_name":
					source.Mongodb.DatabaseName = v.(string)
				case "collection_name":
					source.Mongodb.CollectionName = v.(string)
				}
			}
			sources = append(sources, source)
		}
	}

	return &sources
}

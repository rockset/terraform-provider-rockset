package rockset

import (
	"context"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/rockset/rockset-go-client"
	"github.com/rockset/rockset-go-client/option"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/rockset/rockset-go-client/openapi"
)

// The base collection schema will be the foundation
// of each <type>_collection schema
// It will implement all arguments except sources,
// even though many of these won't likely be used
// for just a write api collection.
func baseCollectionSchema() map[string]*schema.Schema { //nolint:funlen
	return map[string]*schema.Schema{
		"description": {
			Description: "Text describing the collection.",
			Type:        schema.TypeString,
			Default:     "created by Rockset terraform provider",
			Optional:    true,
		},
		"ingest_transformation": {
			Description: `Ingest transformation SQL query. Turns the collection into insert_only mode.

When inserting data into Rockset, you can transform the data by providing a single SQL query, 
that contains all of the desired data transformations. 
This is referred to as the collection’s ingest transformation or, historically, its field mapping query.

For more information see https://rockset.com/docs/ingest-transformation/`,
			Type:     schema.TypeString,
			Optional: true,
		},
		"name": {
			Description:  "Unique identifier for the collection. Can contain alphanumeric or dash characters.",
			Type:         schema.TypeString,
			ForceNew:     true,
			Required:     true,
			ValidateFunc: rocksetNameValidator,
		},
		"retention_secs": {
			Description:  "Number of seconds after which data is purged. Based on event time.",
			Type:         schema.TypeInt,
			ForceNew:     true,
			Optional:     true,
			ValidateFunc: validation.IntAtLeast(0),
		},
		"wait_for_collection": {
			Description:  "Wait until the collection is ready.",
			Type:         schema.TypeBool,
			Optional:     true,
			Default:      true,
			ForceNew:     true,
			RequiredWith: []string{"wait_for_documents"},
		},
		"wait_for_documents": {
			Description:  "Wait until the collection has documents. The default is to wait for 0 documents, which means it doesn't wait.",
			Type:         schema.TypeInt,
			Optional:     true,
			Default:      0,
			ForceNew:     true,
			ValidateFunc: validation.IntAtLeast(0),
		},
		"workspace": {
			Description:  "The name of the workspace.",
			Type:         schema.TypeString,
			ForceNew:     true,
			Required:     true,
			ValidateFunc: rocksetNameValidator,
		},
	} // End schema return
} // End func

// parseBaseCollection takes in a collection returned from the api, parses the base fields any collection has,
// and puts them into the schema object.
func parseBaseCollection(collection *openapi.Collection, d *schema.ResourceData) error {
	var err error

	err = d.Set("name", collection.GetName())
	if err != nil {
		return err
	}

	err = d.Set("workspace", collection.GetWorkspace())
	if err != nil {
		return err
	}

	err = d.Set("description", collection.GetDescription())
	if err != nil {
		return err
	}

	err = d.Set("retention_secs", collection.GetRetentionSecs())
	if err != nil {
		return err
	}

	err = d.Set("ingest_transformation", collection.GetFieldMappingQuery().Sql)
	if err != nil {
		return err
	}

	return nil // No errors
}

func createBaseCollectionRequest(d *schema.ResourceData) *openapi.CreateCollectionRequest {
	/*
		Parses resource data and returns a create collection request
		with all the base fields a basic collection will have.
		Per-source terraform resources can add to the collection request
		to implement sources and other fields related to the source.
	*/
	name := d.Get("name").(string)
	description := d.Get("description").(string)

	params := openapi.NewCreateCollectionRequest()
	params.Name = &name
	params.SetDescription(description)

	if v, ok := d.GetOk("retention_secs"); ok {
		retentionSecondsDuration := time.Duration(v.(int)) * time.Second
		retentionSeconds := int64(retentionSecondsDuration.Seconds())
		params.RetentionSecs = &retentionSeconds
	}

	if v, ok := d.GetOk("ingest_transformation"); ok {
		fmq := v.(string)
		params.FieldMappingQuery = &openapi.FieldMappingQuery{Sql: &fmq}
	}

	return params
}

const defaultCollectionTimeout = 20 * time.Minute

func resourceCollection() *schema.Resource {
	return &schema.Resource{
		Description: "Manages a basic collection with no sources. Usually used for the write api.",

		CreateContext: resourceCollectionCreate,
		ReadContext:   resourceCollectionRead,
		UpdateContext: resourceCollectionUpdate,
		DeleteContext: resourceCollectionDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: baseCollectionSchema(),
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(defaultCollectionTimeout),
		},
	}
}

func resourceCollectionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rc := meta.(*rockset.RockClient)
	var diags diag.Diagnostics

	name := d.Get("name").(string)
	workspace := d.Get("workspace").(string)

	params := createBaseCollectionRequest(d)
	_, err := rc.CreateCollection(ctx, workspace, name, option.WithCollectionRequest(*params))
	if err != nil {
		return diag.FromErr(err)
	}

	if err = waitForCollectionAndDocuments(ctx, rc, d, workspace, name); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(toID(workspace, name))

	return diags
}

func resourceCollectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rc := meta.(*rockset.RockClient)
	var diags diag.Diagnostics
	var err error

	workspace, name := workspaceAndNameFromID(d.Id())

	collection, err := rc.GetCollection(ctx, workspace, name)
	if err != nil {
		return checkForNotFoundError(d, err)
	}

	err = parseBaseCollection(&collection, d)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceCollectionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rc := meta.(*rockset.RockClient)
	var diags diag.Diagnostics
	var err error

	workspace, name := workspaceAndNameFromID(d.Id())

	err = rc.DeleteCollection(ctx, workspace, name)
	if err != nil {
		return diag.FromErr(err)
	}

	err = rc.WaitUntilCollectionGone(ctx, workspace, name)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceCollectionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rc := meta.(*rockset.RockClient)
	var diags diag.Diagnostics
	var err error

	workspace, name := workspaceAndNameFromID(d.Id())

	var options []option.CollectionOption
	if desc := d.Get("description"); desc != nil {
		options = append(options, option.WithCollectionDescription(desc.(string)))
	}
	if it := d.Get("ingest_transformation"); it != nil {
		options = append(options, option.WithIngestTransformation(it.(string)))
	}

	_, err = rc.UpdateCollection(ctx, workspace, name, options...)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func waitForCollectionAndDocuments(ctx context.Context, rc *rockset.RockClient, d *schema.ResourceData, workspace, name string) error {
	if wait := d.Get("wait_for_collection").(bool); wait {
		tflog.Debug(ctx, "waiting for collection", map[string]interface{}{
			"workspace": workspace,
			"name":      name,
		})
		if err := rc.WaitUntilCollectionReady(ctx, workspace, name); err != nil {
			return err
		}
	}

	if nDocs := d.Get("wait_for_documents").(int); nDocs > 0 {
		tflog.Debug(ctx, "waiting for collection documents", map[string]interface{}{
			"workspace": workspace,
			"name":      name,
		})
		if err := rc.WaitUntilCollectionHasDocuments(ctx, workspace, name, int64(nDocs)); err != nil {
			return err
		}
	}

	return nil
}

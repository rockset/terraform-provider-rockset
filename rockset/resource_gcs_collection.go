package rockset

import (
	"context"
	"github.com/rockset/rockset-go-client/option"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/rockset/rockset-go-client"
)

func resourceGCSCollection() *schema.Resource {
	return &schema.Resource{
		Description: "Manages a collection with an GCS source attached.",

		CreateContext: resourceGCSCollectionCreate,
		ReadContext:   resourceGCSCollectionRead,
		DeleteContext: resourceCollectionDelete, // No change from base collection delete

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		// This schema will use the base collection schema as a foundation
		// And layer on just the necessary fields for an gcs collection
		Schema: mergeSchemas(baseCollectionSchema(), gcsCollectionSchema()),
	}
}

func gcsCollectionSchema() map[string]*schema.Schema {
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
						Description:  "The name of the Rockset GCS integration.",
						Type:         schema.TypeString,
						ForceNew:     true,
						Required:     true,
						ValidateFunc: rocksetNameValidator,
					},
					"prefix": {
						Type:        schema.TypeString,
						ForceNew:    true,
						Optional:    true,
						Default:     nil,
						Description: "Simple path prefix to GCS key.",
					},
					"bucket": {
						Type:        schema.TypeString,
						ForceNew:    true,
						Required:    true,
						Description: "GCS bucket containing the target data.",
					},
					"format": formatSchema(),
					"csv":    csvSchema(),
					"xml":    xmlSchema(),
					"field_mapping_query": {
						Type:        schema.TypeString,
						ForceNew:    true,
						Optional:    true,
						Description: "Field mapping SQL query.",
					},
				},
			},
		},
	}
}

func resourceGCSCollectionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rc := meta.(*rockset.RockClient)
	var diags diag.Diagnostics
	var err error

	name := d.Get("name").(string)
	workspace := d.Get("workspace").(string)

	// Add all base schema fields
	params := createBaseCollectionRequest(d)
	// Add fields for gcs
	sources, err := makeBucketSourceParams("gcs", d.Get("source"))
	if err != nil {
		return diag.FromErr(err)
	}
	params.Sources = sources

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

func resourceGCSCollectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rc := meta.(*rockset.RockClient)
	var diags diag.Diagnostics
	var err error

	workspace, name := workspaceAndNameFromID(d.Id())

	collection, err := rc.GetCollection(ctx, workspace, name)
	if err != nil {
		return checkForNotFoundError(d, err)
	}

	// Gets all the fields any generic collection has
	err = parseBaseCollection(&collection, d)
	if err != nil {
		return diag.FromErr(err)
	}

	// Gets all the fields relevant to an s3 collection
	err = parseBucketCollection("gcs", &collection, d)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

package rockset

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/rockset/rockset-go-client/openapi"
	"github.com/rockset/rockset-go-client/option"

	"github.com/rockset/rockset-go-client"
)

func s3CollectionSchema() map[string]*schema.Schema {
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
						Description: "The name of the Rockset S3 integration. If no S3 integration is provided " +
							"only data in public S3 buckets are accessible.",
						Type:         schema.TypeString,
						ForceNew:     true,
						Required:     true,
						ValidateFunc: rocksetNameValidator,
					},
					"prefix": {
						Type:        schema.TypeString,
						ForceNew:    true,
						Optional:    true,
						Deprecated:  "use pattern instead",
						Description: "Simple path prefix to S3 keys.",
					},
					"pattern": {
						Type:        schema.TypeString,
						ForceNew:    true,
						Optional:    true,
						Description: "Regex path pattern to S3 keys.",
					},
					"bucket": {
						Type:        schema.TypeString,
						ForceNew:    true,
						Required:    true,
						Description: "S3 bucket containing the target data.",
					},
					"format": formatSchema(),
					"csv":    csvSchema(),
					"xml":    xmlSchema(),
				},
			},
		},
	} // End schema return
} // End func

func resourceS3Collection() *schema.Resource {
	return &schema.Resource{
		Description: "Manages a collection with on or more S3 sources attached. " +
			"Uses an S3 integration to access the S3 bucket. If no integration is provided, " +
			"only data in public buckets are accessible.\n\n",

		CreateContext: resourceS3CollectionCreate,
		ReadContext:   resourceS3CollectionRead,
		UpdateContext: resourceCollectionUpdate, // No change from base collection update
		DeleteContext: resourceCollectionDelete, // No change from base collection delete

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		// This schema will use the base collection schema as a foundation
		// And layer on just the necessary fields for an s3 collection
		Schema: mergeSchemas(baseCollectionSchema(), s3CollectionSchema()),
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(defaultCollectionTimeout),
		},
	}
}

func resourceS3CollectionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rc := meta.(*rockset.RockClient)
	var diags diag.Diagnostics
	var err error

	name := d.Get("name").(string)
	workspace := d.Get("workspace").(string)

	// Add all base schema fields
	params := createBaseCollectionRequest(d)
	// Add fields for s3
	sources, err := makeBucketSourceParams("s3", d.Get("source"))
	if err != nil {
		return DiagFromErr(err)
	}
	params.Sources = sources

	c, err := rc.CreateCollection(ctx, workspace, name, option.WithCollectionRequest(*params))
	if err != nil {
		return DiagFromErr(err)
	}
	tflog.Trace(ctx, "created Rockset collection", map[string]interface{}{"workspace": workspace, "name": name},
		sourcesToTraceInfo(c.Sources))

	if err = waitForCollectionAndDocuments(ctx, rc, d, workspace, name); err != nil {
		return DiagFromErr(err)
	}

	d.SetId(toID(workspace, name))

	return diags
}

func resourceS3CollectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rc := meta.(*rockset.RockClient)
	var diags diag.Diagnostics
	var err error

	workspace, name := workspaceAndNameFromID(d.Id())

	collection, err := rc.GetCollection(ctx, workspace, name)
	if err != nil {
		return checkForNotFoundError(d, err)
	}
	tflog.Trace(ctx, "read Rockset collection", map[string]interface{}{"workspace": workspace, "name": name},
		sourcesToTraceInfo(collection.Sources))

	// Gets all the fields any generic collection has
	err = parseBaseCollection(&collection, d)
	if err != nil {
		return DiagFromErr(err)
	}

	// Gets all the fields relevant to a s3 collection
	err = parseBucketCollection("s3", &collection, d)
	if err != nil {
		return DiagFromErr(err)
	}

	return diags
}

func sourcesToTraceInfo(sources []openapi.Source) map[string]any {
	trace := make(map[string]any)
	for i, source := range sources {
		trace["source_"+strconv.Itoa(i)] = source
	}
	return trace
}

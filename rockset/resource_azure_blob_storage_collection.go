package rockset

import (
	"context"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/rockset/rockset-go-client/option"

	"github.com/rockset/rockset-go-client"
)

func azureBlobStorageCollectionSchema() map[string]*schema.Schema {
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
						Description: "The name of the Rockset Azure Blob Store integration. If no Azure Blob Store integration is provided " +
							"only data in public Azure Blob Store buckets are accessible.",
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
						Description: "Simple path prefix to Azure Blob Storage container keys.",
					},
					"pattern": {
						Type:        schema.TypeString,
						ForceNew:    true,
						Optional:    true,
						Description: "Regex path pattern to Azure Blob Storage keys.",
					},
					"container": {
						Type:        schema.TypeString,
						ForceNew:    true,
						Required:    true,
						Description: "Azure Blob Store container containing the target data.",
					},
					"format": formatSchema(),
					"csv":    csvSchema(),
					"xml":    xmlSchema(),
				},
			},
		},
	} // End schema return
} // End func

func resourceAzureBlobStorageCollection() *schema.Resource {
	return &schema.Resource{
		Description: "Manages a collection with on or more Azure Blob Storage sources attached. " +
			"Uses an Azure Blob Storage integration to access the Azure Blob Storage container. If no integration is provided, " +
			"only data in public storage accounts and containers are accessible.\n\n",

		CreateContext: resourceAzureBlobStorageCollectionCreate,
		ReadContext:   resourceAzureBlobStorageCollectionRead,
		UpdateContext: resourceCollectionUpdate, // No change from base collection update
		DeleteContext: resourceCollectionDelete, // No change from base collection delete

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		// This schema will use the base collection schema as a foundation
		// And layer on just the necessary fields for an s3 collection
		Schema: mergeSchemas(baseCollectionSchema(), azureBlobStorageCollectionSchema()),
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(defaultCollectionTimeout),
		},
	}
}

func resourceAzureBlobStorageCollectionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rc := meta.(*rockset.RockClient)
	var diags diag.Diagnostics
	var err error

	name := d.Get("name").(string)
	workspace := d.Get("workspace").(string)

	// Add all base schema fields
	params := createBaseCollectionRequest(d)
	// Add fields for Azure Blob Storage
	sources, err := makeBucketSourceParams("azure_blob_storage", d.Get("source"))
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

func resourceAzureBlobStorageCollectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
	err = parseBucketCollection(ctx, "azure_blob_storage", &collection, d)
	if err != nil {
		return DiagFromErr(err)
	}

	return diags
}

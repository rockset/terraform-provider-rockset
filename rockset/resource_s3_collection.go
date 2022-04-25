package rockset

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
					"field_mapping_query": {
						Type:        schema.TypeString,
						ForceNew:    true,
						Optional:    true,
						Description: "Field mapping SQL query.",
					},
				},
			},
		},
	} // End schema return
} // End func

func resourceS3Collection() *schema.Resource {
	return &schema.Resource{
		Description: "Manages a collection with on or more S3 sources attached. " +
			"Uses an S3 integration to access the S3 bucket. If no integration is provided, " +
			"only data in public buckets are accessible.\n\n" +
			"```hcl\n" +
			// TODO use go:generate to include this file
			//   examples/resources/rockset_s3_collection/resource.tf
			`resource rockset_workspace sample {
  name = "sample"
  description = "sample datasets"
}

resource rockset_s3_integration public {
  name = "rockset-public-collections"
  description = "Integration to access Rockset's public datasets"
  aws_role_arn = "arn:aws:iam::469279130686:role/rockset-public-datasets"
}

resource rockset_s3_collection cities {
  workspace = rockset_workspace.sample.name
  name = "cities"
  integration_name = rockset_s3_integration.public.name
  source = {
    bucket = "rockset-public-datasets"
    prefix = "partial-cities"
    format = "json"
  }
}


` +
			"```",

		CreateContext: resourceS3CollectionCreate,
		ReadContext:   resourceS3CollectionRead,
		DeleteContext: resourceCollectionDelete, // No change from base collection delete

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		// This schema will use the base collection schema as a foundation
		// And layer on just the necessary fields for an s3 collection
		Schema: mergeSchemas(baseCollectionSchema(), s3CollectionSchema()),
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
		return diag.FromErr(err)
	}
	params.Sources = sources

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

func resourceS3CollectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
	err = parseBucketCollection("s3", &collection, d)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

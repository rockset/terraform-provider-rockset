package rockset

import (
	"context"
	"fmt"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/rockset/rockset-go-client"
	"github.com/rockset/rockset-go-client/dataset"
	"github.com/rockset/rockset-go-client/openapi"
	"github.com/rockset/rockset-go-client/option"
	"strings"
)

func sampleCollectionSchema() map[string]*schema.Schema {
	var datasets []string
	for _, ds := range dataset.All() {
		datasets = append(datasets, string(ds))
	}

	return map[string]*schema.Schema{
		"dataset": {
			Description: fmt.Sprintf("Name of the sample dataset, must be one of: %s",
				strings.Join(datasets, ", ")),
			Type:     schema.TypeString,
			ForceNew: true,
			Required: true,
			ValidateDiagFunc: func(i interface{}, path cty.Path) diag.Diagnostics {
				s := i.(string)
				ds := dataset.Sample(s)
				m := dataset.Lookup(ds)
				if m == "" {
					return diag.Errorf("%s is not a known sample dataset", s)
				}

				return diag.Diagnostics{}
			},
		},
	}
} // End func

func resourceSampleCollection() *schema.Resource {
	return &schema.Resource{
		Description: "Manages a collection created from one of the Rockset sample datasets.",

		CreateContext: resourceSampleCollectionCreate,
		ReadContext:   resourceSampleCollectionRead,
		DeleteContext: resourceCollectionDelete, // No change from base collection delete

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		// This schema will use the base collection schema as a foundation
		// And layer on just the necessary fields for this collection
		Schema: mergeSchemas(baseCollectionSchema(), sampleCollectionSchema()),
	}
}

func resourceSampleCollectionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rc := meta.(*rockset.RockClient)
	var diags diag.Diagnostics
	var err error

	name := d.Get("name").(string)
	workspace := d.Get("workspace").(string)

	// Add all base schema fields
	params := createBaseCollectionRequest(d)

	ds := dataset.Sample(d.Get("dataset").(string))
	pattern := dataset.Lookup(ds)
	params.Sources = []openapi.Source{{S3: &openapi.SourceS3{
		Pattern: &pattern,
		Region:  openapi.PtrString("us-west-2"),
		Bucket:  dataset.RocksetPublicDatasets,
	}}}

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

func resourceSampleCollectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
	err = parseBucketCollection("XXX", &collection, d)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

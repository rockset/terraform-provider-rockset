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

func s3CollectionSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"integration_name": {
			Type:     schema.TypeString,
			ForceNew: true,
			Required: true,
		},
		"prefix": {
			Type:        schema.TypeString,
			ForceNew:    true,
			Optional:    true,
			Default:     nil,
			Description: "Simple path prefix to s3 key.",
		},
		"pattern": {
			Type:        schema.TypeString,
			ForceNew:    true,
			Optional:    true,
			Default:     nil,
			Description: "Regex path prefix to s3 key.",
		},
		"bucket": {
			Type:        schema.TypeString,
			ForceNew:    true,
			Required:    true,
			Description: "S3 bucket containing the target data.",
		},
		"format": {
			Type:     schema.TypeString,
			ForceNew: true,
			Required: true,
			ValidateFunc: validation.StringMatch(
				regexp.MustCompile("^(json|csv)$"), "only 'json' or 'csv' is supported"),
			Description: "Format of the data. One of: json, csv. csv block cannot be set if format is json.",
		},
		"csv": {
			Type:     schema.TypeSet,
			ForceNew: true,
			Optional: true,
			MinItems: 0,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"first_line_as_column_names": {
						Type:     schema.TypeBool,
						ForceNew: true,
						Optional: true,
						Default:  true,
					},
					"separator": {
						Type:     schema.TypeString,
						ForceNew: true,
						Optional: true,
						Default:  ",",
					},
					"encoding": {
						Type:     schema.TypeString,
						ForceNew: true,
						Optional: true,
						Default:  "UTF-8",
						ValidateFunc: validation.StringMatch(
							regexp.MustCompile("^(UTF-8|UTF-16|ISO_8859_1)$"), "must be either 'UTF-8', 'UTF-16' or 'ISO_8859_1'"),
					},
					"column_names": {
						Type:     schema.TypeList,
						ForceNew: true,
						Optional: true,
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
					},
					"column_types": {
						Type:     schema.TypeList,
						ForceNew: true,
						Optional: true,
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
					},
					"quote_char": {
						Type:     schema.TypeString,
						ForceNew: true,
						Optional: true,
						Default:  `"`,
					},
					"escape_char": {
						Type:     schema.TypeString,
						ForceNew: true,
						Optional: true,
						Default:  `\`,
					},
				},
			},
		},
	} // End schema return
} // End func

func resourceS3Collection() *schema.Resource {
	return &schema.Resource{
		Description: "Manages a collection with an s3 source attached.",

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
	err = addS3Params(d, params)
	if err != nil {
		return diag.FromErr(err)
	}

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
		return diag.FromErr(err)
	}

	// Gets all the fields any generic collection has
	err = parseBaseCollection(&collection, d)
	if err != nil {
		return diag.FromErr(err)
	}

	// Gets all the fields relevant to an s3 collection
	err = parseS3Collection(&collection, d)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

/*
	Takes in a collection returned from the api.
	Parses the fields relevant to an s3 source and
	puts them into the schema object.
*/
func parseS3Collection(collection *openapi.Collection, d *schema.ResourceData) error {

	var err error

	sourcesList := *collection.Sources
	sourcesCount := len(sourcesList)
	if sourcesCount != 1 {
		return fmt.Errorf("expected %s to have exactly one source, got %d", collection.GetName(), sourcesCount)
	}

	s3Source := sourcesList[0]
	formatParams := s3Source.FormatParams
	if *formatParams.Json {
		err = d.Set("format", "json")
		if err != nil {
			return err
		}
	} else if formatParams.Csv != nil {
		err = d.Set("format", "csv")
		if err != nil {
			return err
		}

		err = d.Set("csv", flattenCsvParams(formatParams.Csv))
		if err != nil {
			return err
		}
	}

	err = d.Set("prefix", s3Source.S3.GetPrefix())
	if err != nil {
		return err
	}

	err = d.Set("pattern", s3Source.S3.GetPattern())
	if err != nil {
		return err
	}

	err = d.Set("bucket", s3Source.S3.GetBucket())
	if err != nil {
		return err
	}

	err = d.Set("integration_name", s3Source.GetIntegrationName())
	if err != nil {
		return err
	}

	return nil // No errors
}

func addS3Params(d *schema.ResourceData, params *openapi.CreateCollectionRequest) error {
	/*
		Adds the s3 sources data to the create collection request
	*/
	var format = openapi.FormatParams{}
	switch d.Get("format").(string) {
	case "json":
		format.Json = openapi.PtrBool(true)
		if d.Get("csv") != nil {
			return fmt.Errorf("can't define csv block with json format")
		}
	case "csv":
		format.Csv = makeCsvParams(d.Get("csv"))
	}

	sources := []openapi.Source{
		{
			FormatParams:    &format,
			IntegrationName: d.Get("integration_name").(string),
			S3: &openapi.SourceS3{
				Prefix:  toStringPtrNilIfEmpty(d.Get("prefix").(string)),
				Pattern: toStringPtrNilIfEmpty(d.Get("pattern").(string)),
				Bucket:  d.Get("bucket").(string),
			},
		},
	}
	params.Sources = &sources

	return nil
}

func flattenCsvParams(params *openapi.CsvParams) []interface{} {
	m := make(map[string]interface{})

	m["first_line_as_column_names"] = *params.FirstLineAsColumnNames
	m["separator"] = *params.Separator
	m["encoding"] = *params.Encoding
	m["escape_char"] = *params.EscapeChar
	m["quote_char"] = *params.QuoteChar
	m["column_names"] = *params.ColumnNames
	m["column_types"] = *params.ColumnTypes

	return []interface{}{m}
}

func makeCsvParams(in interface{}) *openapi.CsvParams {
	m := openapi.CsvParams{}

	for _, i := range in.(*schema.Set).List() {
		if val, ok := i.(map[string]interface{}); ok {
			for k, v := range val {
				switch k {
				case "first_line_as_column_names":
					m.FirstLineAsColumnNames = openapi.PtrBool(v.(bool))
				case "separator":
					m.Separator = toStringPtrNilIfEmpty(v.(string))
				case "encoding":
					m.Encoding = toStringPtrNilIfEmpty(v.(string))
				case "quote_char":
					m.QuoteChar = toStringPtrNilIfEmpty(v.(string))
				case "escape_char":
					m.EscapeChar = toStringPtrNilIfEmpty(v.(string))
				case "column_names":
					m.ColumnNames = toStringArrayPtr(toStringArray(v.([]interface{})))
				case "column_types":
					m.ColumnTypes = toStringArrayPtr(toStringArray(v.([]interface{})))
				}
			}
		}
	}

	return &m
}

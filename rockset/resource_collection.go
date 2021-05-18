package rockset

import (
	"context"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/rockset/rockset-go-client"
	"github.com/rockset/rockset-go-client/openapi"
)

// The base collection schema will be the foundation
// of each <type>_collection schema
// It will implement all arguments except sources,
// even though many of these won't likely be used
// for just a write api collection.
func baseCollectionSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Description:  "Unique identifier for collection. Can contain alphanumeric or dash characters.",
			Type:         schema.TypeString,
			ForceNew:     true,
			Required:     true,
			ValidateFunc: rocksetNameValidator,
		},
		"description": {
			Description: "Text describing the collection.",
			Type:        schema.TypeString,
			Default:     "created by Rockset terraform provider",
			ForceNew:    true,
			Optional:    true,
		},
		"workspace": {
			Type:     schema.TypeString,
			ForceNew: true,
			Required: true,
		},
		"retention_secs": {
			Description:  "Number of seconds after which data is purged. Based on event time.",
			Type:         schema.TypeInt,
			ForceNew:     true,
			Optional:     true,
			ValidateFunc: validation.IntAtLeast(1),
		},
		"event_time_info": {
			Description: "Configuration for event data.",
			Type:        schema.TypeSet,
			ForceNew:    true,
			Optional:    true,
			MinItems:    0,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"field": {
						Description: "name of the field containing event time",
						Type:        schema.TypeString,
						ForceNew:    true,
						Required:    true,
					},
					"format": {
						Description: "format of time field",
						Type:        schema.TypeString,
						ForceNew:    true,
						Required:    true,
						ValidateFunc: validation.StringInSlice(
							[]string{"milliseconds_since_epoch", "seconds_since_epoch"},
							false), // Ignore case false, must do exact match
					},
					"time_zone": {
						Description: "default time zone, in standard IANA format",
						Type:        schema.TypeString,
						ForceNew:    true,
						Required:    true,
					},
				},
			},
		}, // End event_time_info
		"field_mapping": {
			Description: "List of field mappings.",
			Type:        schema.TypeList,
			ForceNew:    true,
			Optional:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"name": {
						Type:     schema.TypeString,
						ForceNew: true,
						Required: true,
					},
					"input_fields": {
						Type:     schema.TypeList,
						ForceNew: true,
						Optional: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"field_name": {
									Type:     schema.TypeString,
									ForceNew: true,
									Required: true,
								},
								"param": {
									Type:     schema.TypeString,
									ForceNew: true,
									Required: true,
								},
								"if_missing": {
									Type:     schema.TypeString,
									ForceNew: true,
									Required: true,
									ValidateFunc: validation.StringMatch(
										regexp.MustCompile("^(PASS|SKIP)$"), "must be either 'PASS' or 'SKIP'"),
								},
								"is_drop": {
									Type:     schema.TypeBool,
									ForceNew: true,
									Required: true,
								},
							},
						},
					},
					"output_field": {
						Type:     schema.TypeSet,
						ForceNew: true,
						Required: true,
						MinItems: 1,
						MaxItems: 1,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"field_name": {
									Type:     schema.TypeString,
									ForceNew: true,
									Required: true,
								},
								"sql": {
									Type:     schema.TypeString,
									ForceNew: true,
									Required: true,
								},
								"on_error": {
									Type:     schema.TypeString,
									ForceNew: true,
									Required: true,
									ValidateFunc: validation.StringMatch(
										regexp.MustCompile("^(FAIL|SKIP)$"), "must be either 'FAIL' or 'SKIP'"),
								},
							},
						},
					},
				},
			},
		}, // End field_mapping
		"clustering_key": {
			Description: "List of clustering fields.",
			Type:        schema.TypeList,
			ForceNew:    true,
			Optional:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"field_name": {
						Description: "The name of a field. Parsed as a SQL qualified name.",
						Type:        schema.TypeString,
						ForceNew:    true,
						Required:    true,
					},
					"type": {
						Description: "The type of partitions on a field.",
						Type:        schema.TypeString,
						ForceNew:    true,
						Required:    true,
					},
					"keys": {
						Description: "The values for partitioning of a field.",
						Type:        schema.TypeList,
						ForceNew:    true,
						Required:    true,
						MinItems:    1,
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
					},
				},
			},
		}, // End clustering_key
		"field_schemas": {
			Description: "List of field schemas.",
			Type:        schema.TypeList,
			ForceNew:    true,
			Optional:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"field_name": {
						Description: "The name of a field. Parsed as a SQL qualified name.",
						Type:        schema.TypeString,
						ForceNew:    true,
						Required:    true,
					},
					"index_mode": {
						Description: "Whether to have index or no_index.",
						Type:        schema.TypeString,
						ForceNew:    true,
						Required:    true,
						ValidateFunc: validation.StringInSlice(
							[]string{"index", "no_index"},
							false), // Ignore case false, must do exact match
					},
					"range_index_mode": {
						Description: "Whether to have v1_index or no_index.",
						Type:        schema.TypeString,
						ForceNew:    true,
						Required:    true,
						ValidateFunc: validation.StringInSlice(
							[]string{"v1_index", "no_index"},
							false), // Ignore case false, must do exact match
					},
					"type_index_mode": {
						Description: "Whether to have index or no_index.",
						Type:        schema.TypeString,
						ForceNew:    true,
						Required:    true,
						ValidateFunc: validation.StringInSlice(
							[]string{"index", "no_index"},
							false), // Ignore case false, must do exact match
					},
					"column_index_mode": {
						Description: "Whether to have store or no_store.",
						Type:        schema.TypeString,
						ForceNew:    true,
						Required:    true,
						ValidateFunc: validation.StringInSlice(
							[]string{"store", "no_store"},
							false), // Ignore case false, must do exact match
					},
				},
			},
		}, // End field_schemas
		"inverted_index_group_encoding_options": {
			Description: "Inverted index group encoding options.",
			Type:        schema.TypeSet,
			ForceNew:    true,
			Optional:    true,
			MinItems:    0,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"group_size": {
						Description: "Group size.",
						Type:        schema.TypeInt,
						ForceNew:    true,
						Required:    true,
					},
					"restart_length": {
						Description: "Restart length.",
						Type:        schema.TypeInt,
						ForceNew:    true,
						Required:    true,
					},
					"event_time_codec": {
						Description: "Event time codec.",
						Type:        schema.TypeString,
						ForceNew:    true,
						Required:    true,
					},
					"doc_id_codec": {
						Description: "Doc id codec.",
						Type:        schema.TypeString,
						ForceNew:    true,
						Required:    true,
					},
				},
			},
		}, // End inverted_index_group_encoding_options
	} // End schema return
} // End func

func resourceCollection() *schema.Resource {
	return &schema.Resource{
		Description: "Manages a basic collection with no resources. Usually used for the write api.",

		CreateContext: resourceCollectionCreate,
		ReadContext:   resourceCollectionRead,
		DeleteContext: resourceCollectionDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: baseCollectionSchema(),
	}
}

func resourceCollectionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rc := meta.(*rockset.RockClient)
	var diags diag.Diagnostics

	workspace := d.Get("workspace").(string)
	name := d.Get("name").(string)
	description := d.Get("description").(string)

	req := rc.CollectionsApi.CreateCollection(ctx, workspace)
	params := openapi.NewCreateCollectionRequest(name)
	params.SetDescription(description)

	_, _, err := req.Body(*params).Execute()
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

func parseBaseCollection(collection *openapi.Collection, d *schema.ResourceData) error {
	/*
		Takes in a collection returned from the api.
		Parses the base fields any collection has and
		puts them into the schema object.
	*/
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

	return nil // No errors
}

func resourceCollectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rc := meta.(*rockset.RockClient)
	var diags diag.Diagnostics
	var err error

	workspace, name := workspaceAndNameFromID(d.Id())

	collection, err := rc.GetCollection(ctx, workspace, name)
	if err != nil {
		return diag.FromErr(err)
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

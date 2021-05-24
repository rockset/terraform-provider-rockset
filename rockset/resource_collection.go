package rockset

import (
	"context"
	"log"
	"regexp"
	"time"

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
						Optional:    true,
					},
					"keys": {
						Description: "The values for partitioning of a field.",
						Type:        schema.TypeList,
						ForceNew:    true,
						Optional:    true,
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

/*
	Takes in a collection returned from the api.
	Parses the base fields any collection has and
	puts them into the schema object.
*/
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

	err = d.Set("field_mapping", flattenFieldMappings(collection.GetFieldMappings()))
	if err != nil {
		return err
	}

	err = d.Set("clustering_key", flattenClusteringKeys(collection.GetClusteringKey()))
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

	params := openapi.NewCreateCollectionRequest(name)
	params.SetDescription(description)

	if v, ok := d.GetOk("field_mapping"); ok && len(v.([]interface{})) > 0 {
		mappings := makeFieldMappings(v.([]interface{}))
		params.SetFieldMappings(*mappings)
	}

	if v, ok := d.GetOk("retention_secs"); ok {
		retentionSecondsDuration := time.Duration(v.(int)) * time.Second
		retentionSeconds := int64(retentionSecondsDuration.Seconds())
		params.RetentionSecs = &retentionSeconds
	}

	if v, ok := d.GetOk("clustering_key"); ok && len(v.([]interface{})) > 0 {
		// The api and the go client use the singular 'ClusteringKey'
		// But the value is in fact a list.
		clusteringKeys := makeClusteringKeys(v.([]interface{}))
		params.ClusteringKey = clusteringKeys
	}

	return params
}

func resourceCollection() *schema.Resource {
	return &schema.Resource{
		Description: "Manages a basic collection with no sources. Usually used for the write api.",

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

	name := d.Get("name").(string)
	workspace := d.Get("workspace").(string)

	req := rc.CollectionsApi.CreateCollection(ctx, workspace)
	params := createBaseCollectionRequest(d)

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

func makeClusteringKeys(v []interface{}) *[]openapi.FieldPartition {
	clusteringKeys := make([]openapi.FieldPartition, 0, len(v))
	for _, raw := range v {
		fp := openapi.FieldPartition{}
		cfg := raw.(map[string]interface{})

		if v, ok := cfg["field_name"]; ok {
			fieldName := v.(string)
			fp.FieldName = &fieldName
		}

		if v, ok := cfg["type"]; ok {
			partitionType := v.(string)
			fp.Type = &partitionType
		}

		if v, ok := cfg["keys"]; ok {
			partitionKeys := toStringArray(v.([]interface{}))
			fp.Keys = &partitionKeys
		}

		clusteringKeys = append(clusteringKeys, fp)
	}

	return &clusteringKeys
}

func makeFieldMappings(v []interface{}) *[]openapi.FieldMappingV2 {
	mappings := make([]openapi.FieldMappingV2, 0, len(v))
	for _, raw := range v {
		fm := openapi.FieldMappingV2{}
		cfg := raw.(map[string]interface{})

		if v, ok := cfg["name"]; ok {
			fieldMappingName := v.(string)
			fm.Name = &fieldMappingName
		}

		if v, ok := cfg["output_field"]; ok {
			log.Printf("output_field(%T): %+v", v, v)
			fm.OutputField = makeOutputField(v)
		}

		if v, ok := cfg["input_fields"]; ok {
			fm.InputFields = makeInputFields(v)
		}

		mappings = append(mappings, fm)
	}

	return &mappings
}

func makeOutputField(in interface{}) *openapi.OutputField {
	of := openapi.OutputField{}

	for _, i := range in.(*schema.Set).List() {
		if val, ok := i.(map[string]interface{}); ok {
			for k, v := range val {
				switch k {
				case "field_name":
					field := v.(string)
					of.FieldName = &field
				case "on_error":
					field := v.(string)
					of.OnError = &field
				case "sql":
					field := v.(string)
					of.Value = &openapi.SqlExpression{Sql: &field}
				}
			}
		}
	}

	return &of
}

func makeInputFields(in interface{}) *[]openapi.InputField {
	fields := make([]openapi.InputField, 0)
	log.Printf("in: %T", in)

	if arr, ok := in.([]interface{}); ok {
		for _, a := range arr {
			cfg, ok := a.(map[string]interface{})
			if !ok {
				log.Printf("failed to cast %+v to map[string]interface{}", a)
				continue
			}

			i := openapi.InputField{}

			if v, ok := cfg["field_name"]; ok {
				field := v.(string)
				i.FieldName = &field
			}

			if v, ok := cfg["param"]; ok {
				field := v.(string)
				i.Param = &field
			}

			if v, ok := cfg["if_missing"]; ok {
				field := v.(string)
				i.IfMissing = &field
			}

			if v, ok := cfg["is_drop"]; ok {
				field := v.(bool)
				i.IsDrop = &field
			}

			fields = append(fields, i)
		}
	}

	return &fields
}

func flattenFieldMappings(fieldMappings []openapi.FieldMappingV2) []interface{} {
	var out = make([]interface{}, 0, len(fieldMappings))

	for _, f := range fieldMappings {
		m := make(map[string]interface{})

		m["name"] = f.Name
		m["output_field"] = flattenOutputField(*f.OutputField)
		m["input_fields"] = flattenInputFields(*f.InputFields)

		out = append(out, m)
	}

	return out
}

func flattenOutputField(outputField openapi.OutputField) []interface{} {
	m := make(map[string]interface{})

	m["field_name"] = outputField.FieldName
	m["on_error"] = outputField.OnError
	m["sql"] = outputField.Value.Sql

	return []interface{}{m}
}

func flattenInputFields(inputFields []openapi.InputField) []interface{} {
	var out = make([]interface{}, 0, len(inputFields))

	for _, i := range inputFields {
		m := make(map[string]interface{})
		m["field_name"] = i.FieldName
		m["if_missing"] = i.IfMissing
		m["is_drop"] = i.IsDrop
		m["param"] = i.Param
		out = append(out, m)
	}

	return out
}

func flattenClusteringKeys(clusteringKeys []openapi.FieldPartition) []interface{} {
	var out = make([]interface{}, 0, len(clusteringKeys))

	for _, fieldPartition := range clusteringKeys {
		m := make(map[string]interface{})

		m["field_name"] = fieldPartition.FieldName
		m["type"] = fieldPartition.Type
		m["keys"] = fieldPartition.Keys

		out = append(out, m)
	}

	return out
}

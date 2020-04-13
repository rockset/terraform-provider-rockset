package rockset

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/rockset/rockset-go-client"
	models "github.com/rockset/rockset-go-client/lib/go"
	"log"
	"regexp"
	"strings"
)

func resourceS3Collection() *schema.Resource {
	return &schema.Resource{
		Create: resourceS3CollectionCreate,
		Read:   resourceS3CollectionRead,
		Delete: resourceS3CollectionDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				ForceNew:     true,
				Required:     true,
				ValidateFunc: rocksetNameValidator,
			},
			"description": {
				Type:     schema.TypeString,
				Default:  "created by Rockset terraform provider",
				ForceNew: true,
				Optional: true,
			},
			"workspace": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"retention": {
				Type:        schema.TypeInt,
				ForceNew:    true,
				Required:    true,
				Description: "retention period in seconds",
			},
			"integration_name": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"prefix": {
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
				Default:  "",
			},
			"pattern": {
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
				Default:  "",
			},
			"bucket": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"format": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
				ValidateFunc: validation.StringMatch(
					regexp.MustCompile("^(json|csv)$"), "only 'json' or 'csv' is supported"),
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
			"field_mapping": {
				Type:     schema.TypeList,
				ForceNew: true,
				Optional: true,
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
			},
		},
	}
}

func resourceS3CollectionCreate(d *schema.ResourceData, m interface{}) error {
	rc := m.(*rockset.RockClient)

	workspace := d.Get("workspace").(string)
	name := d.Get("name").(string)

	var format = models.FormatParams{}
	switch d.Get("format").(string) {
	case "json":
		format.Json = true
		if d.Get("csv") != nil {
			return fmt.Errorf("can't define csv block with json format")
		}
	case "csv":
		format.Csv = makeCsvParams(d.Get("csv"))
	}

	request := models.CreateCollectionRequest{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		Sources: []models.Source{
			{
				IntegrationName: d.Get("integration_name").(string),
				S3: &models.SourceS3{
					Prefix:  d.Get("prefix").(string),
					Pattern: d.Get("pattern").(string),
					Bucket:  d.Get("bucket").(string),
				},
				FormatParams: &format,
			},
		},
	}

	if v, ok := d.GetOk("field_mapping"); ok && len(v.([]interface{})) > 0 {
		mappings := make([]models.FieldMappingV2, 0)
		for _, raw := range v.([]interface{}) {
			fm := models.FieldMappingV2{}
			cfg := raw.(map[string]interface{})

			if v, ok := cfg["name"]; ok {
				fm.Name = v.(string)
			}

			if v, ok := cfg["output_field"]; ok {
				fm.OutputField = makeOutputField(v)
			}

			if v, ok := cfg["input_fields"]; ok {
				fm.InputFields = makeInputFields(v)
			}

			mappings = append(mappings, fm)
		}
		request.FieldMappings = mappings
	}

	_, _, err := rc.Collection.Create(workspace, request)
	if err != nil {
		return asSwaggerMessage(err)
	}

	d.SetId(toWorkspaceID(workspace, name))

	return resourceS3CollectionRead(d, m)
}

func resourceS3CollectionRead(d *schema.ResourceData, m interface{}) error {
	rc := m.(*rockset.RockClient)

	workspace, name := workspaceID(d.Id())

	res, _, err := rc.Collection.Get(workspace, name)
	if err != nil {
		return asSwaggerMessage(err)
	}

	if len(res.Data.Sources) != 1 {
		return fmt.Errorf("expected %s to have exactly one source, got %d", name, len(res.Data.Sources))
	}

	format := res.Data.Sources[0].FormatParams
	if format.Json {
		d.Set("format", "json")
	} else if format.Csv != nil {
		d.Set("csv", flattenCsvParams(format.Csv))
	}

	if res.Data.Sources[0].S3 == nil {
		return fmt.Errorf("expected %s to a S3 collection", name)
	}
	s3 := res.Data.Sources[0].S3

	d.Set("name", res.Data.Name)
	d.Set("description", res.Data.Description)
	d.Set("prefix", s3.Prefix)
	d.Set("pattern", s3.Pattern)
	d.Set("bucket", s3.Bucket)

	if res.Data.FieldMappings != nil {
		if err = d.Set("field_mapping", flattenFieldMappings(res.Data.FieldMappings)); err != nil {
			return err
		}
	}

	d.SetId(toWorkspaceID(workspace, name))

	return nil
}

func resourceS3CollectionDelete(d *schema.ResourceData, m interface{}) error {
	rc := m.(*rockset.RockClient)

	workspace := d.Get("workspace").(string)
	name := d.Get("name").(string)

	_, _, err := rc.Collection.Delete(workspace, name)
	if err != nil {
		return asSwaggerMessage(err)
	}

	d.SetId("")

	// loop until the collection is gone as the deletion is asynchronous
	err = resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		log.Printf("checking if %s in workspace %s still exist", name, workspace)
		_, _, err = rc.Collection.Get(workspace, name)
		if err == nil {
			return resource.RetryableError(fmt.Errorf("collection %s still exist", name))
		}
		// TODO we need a better way to check the error
		if err.Error() == "404 Not Found" {
			return nil
		}
		return resource.NonRetryableError(err)
	})
	if err != nil {
		log.Println(err)
	}

	return err
}

func toWorkspaceID(workspace, name string) string {
	// TODO if there are multiple accounts which all have the same workspace and collection name
	// this ID wont't work
	return fmt.Sprintf("%s:%s", workspace, name)
}

func workspaceID(id string) (string, string) {
	tokens := strings.SplitN(id, ":", 2)
	if len(tokens) != 2 {
		log.Printf("unparsable id: %s", id)
		return "", ""
	}
	return tokens[0], tokens[1]
}

func makeOutputField(in interface{}) *models.OutputField {
	of := models.OutputField{}

	if list, ok := in.([]interface{}); ok {
		for _, v := range list {
			if cfg, ok := v.(map[string]interface{}); ok {
				if v, ok := cfg["field_name"]; ok {
					of.FieldName = v.(string)
				}
				if v, ok := cfg["on_error"]; ok {
					of.OnError = v.(string)
				}
				if v, ok := cfg["sql"]; ok {
					of.Value = &models.SqlExpression{Sql: v.(string)}
				}
			}
		}
	} else {
		log.Println("no match")
	}

	return &of
}

func makeInputFields(in interface{}) []models.InputField {
	fields := make([]models.InputField, 0)
	log.Printf("in: %T", in)

	if arr, ok := in.([]interface{}); ok {
		for _, a := range arr {
			if cfg, ok := a.(map[string]interface{}); ok {
				i := models.InputField{}

				if v, ok := cfg["field_name"]; ok {
					i.FieldName = v.(string)
				}

				if v, ok := cfg["param"]; ok {
					i.Param = v.(string)
				}

				if v, ok := cfg["if_missing"]; ok {
					i.IfMissing = v.(string)
				}

				if v, ok := cfg["is_drop"]; ok {
					i.IsDrop = v.(bool)
				}

				fields = append(fields, i)
			} else {
				log.Printf("failed to cast %+v to map[string]interface{}", a)
			}
		}
	}

	return fields
}

func makeCsvParams(in interface{}) *models.CsvParams {
	m := models.CsvParams{}

	for _, i := range in.(*schema.Set).List() {
		if val, ok := i.(map[string]interface{}); ok {
			for k, v := range val {
				log.Printf("%s: %T %v", k, v, v)
				switch k {
				case "first_line_as_column_names":
					m.FirstLineAsColumnNames = v.(bool)
				case "separator":
					m.Separator = v.(string)
				case "encoding":
					m.Encoding = v.(string)
				case "quote_char":
					m.QuoteChar = v.(string)
				case "escape_char":
					m.EscapeChar = v.(string)
				case "column_names":
					m.ColumnNames = toStringArray(v.([]interface{}))
				case "column_types":
					m.ColumnTypes = toStringArray(v.([]interface{}))
				}
			}
		}
	}

	return &m
}

// convert an array of interface{} to an array of string
func toStringArray(a []interface{}) []string {
	r := make([]string, len(a))
	for i, v := range a {
		r[i] = v.(string)
	}
	return r
}

func flattenCsvParams(params *models.CsvParams) []interface{} {
	m := make(map[string]interface{})

	m["first_line_as_column_names"] = params.FirstLineAsColumnNames
	m["separator"] = params.Separator
	m["encoding"] = params.Encoding
	m["escape_char"] = params.EscapeChar
	m["quote_char"] = params.QuoteChar
	m["column_names"] = params.ColumnNames
	m["column_types"] = params.ColumnTypes

	return []interface{}{m}
}

func flattenFieldMappings(fieldMappings []models.FieldMappingV2) []interface{} {
	var out = make([]interface{}, 0, 0)

	for _, f := range fieldMappings {
		m := make(map[string]interface{})

		m["name"] = f.Name
		m["output_field"] = flattenOutputField(*f.OutputField)
		m["input_fields"] = flattenInputFields(f.InputFields)

		out = append(out, m)
	}

	return out
}

func flattenOutputField(outputField models.OutputField) []interface{} {
	m := make(map[string]interface{})

	m["field_name"] = outputField.FieldName
	m["on_error"] = outputField.OnError
	m["sql"] = outputField.Value.Sql

	return []interface{}{m}
}

func flattenInputFields(inputFields []models.InputField) []interface{} {
	var out = make([]interface{}, 0, 0)

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

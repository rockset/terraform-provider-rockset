package rockset

import (
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/rockset/rockset-go-client/openapi"
)

/*
	Takes in a collection returned from the api.
	Parses the fields relevant to an s3 source and
	puts them into the schema object.
*/
func parseBucketCollection(sourceType string, collection *openapi.Collection, d *schema.ResourceData) error {
	var err error

	sourcesList := *collection.Sources
	sourcesCount := len(sourcesList)
	if sourcesCount < 1 {
		return fmt.Errorf("expected %s to have at least 1 source", collection.GetName())
	}

	err = d.Set("source", flattenBucketSourceParams(sourceType, &sourcesList))
	if err != nil {
		return err
	}

	return nil // No errors
}

func flattenBucketSourceParams(sourceType string, sources *[]openapi.Source) []interface{} {
	convertedList := make([]interface{}, 0, len(*sources))
	for _, source := range *sources {
		m := make(map[string]interface{})
		formatParams := source.FormatParams

		if *formatParams.Json {
			m["format"] = "json"
		} else if formatParams.Csv != nil {
			m["format"] = "csv"
			m["csv"] = flattenCsvParams(formatParams.Csv)
		} else if formatParams.Xml != nil {
			m["format"] = "xml"
			m["xml"] = flattenXmlParams(formatParams.Xml)
		}

		m["integration_name"] = source.IntegrationName
		switch sourceType {
		case "gcs":
			m["prefix"] = source.Gcs.Prefix
			m["bucket"] = source.Gcs.Bucket
		case "s3":
			m["prefix"] = source.S3.Prefix
			m["pattern"] = source.S3.Pattern
			m["bucket"] = source.S3.Bucket
		default:
			panic("unknown source type " + sourceType)
		}
		convertedList = append(convertedList, m)
	}

	return convertedList
}

func makeBucketSourceParams(sourceType string, in interface{}) (*[]openapi.Source, error) {
	sources := make([]openapi.Source, 0, in.(*schema.Set).Len())

	for _, i := range in.(*schema.Set).List() {
		if val, ok := i.(map[string]interface{}); ok {
			source := openapi.Source{}
			format := openapi.FormatParams{}
			source.FormatParams = &format
			source.IntegrationName = val["integration_name"].(string)

			switch sourceType {
			case "gcs":
				source.Gcs = openapi.NewSourceGcsWithDefaults()
				source.Gcs.Prefix = toStringPtrNilIfEmpty(val["prefix"].(string))
				bucket := val["bucket"].(string)
				source.Gcs.Bucket = &bucket
			case "s3":
				source.S3 = openapi.NewSourceS3WithDefaults()
				source.S3.Prefix = toStringPtrNilIfEmpty(val["prefix"].(string))
				source.S3.Pattern = toStringPtrNilIfEmpty(val["pattern"].(string))
				source.S3.Bucket = val["bucket"].(string)
			default:
				panic("unknown source type " + sourceType)
			}

			csvBlock, csvBlockOK := val["csv"] // optional fields
			xmlBlock, xmlBlockOK := val["xml"] // optional fields

			xmlBlockIsSet := xmlBlockOK && xmlBlock.(*schema.Set).Len() != 0
			csvBlockIsSet := csvBlockOK && csvBlock.(*schema.Set).Len() != 0

			switch val["format"].(string) {
			case "json":
				if csvBlockIsSet {
					return nil, fmt.Errorf("can't define csv block with json format")
				}
				if xmlBlockIsSet {
					return nil, fmt.Errorf("can't define xml block with json format")
				}
			case "csv":
				if xmlBlockIsSet {
					return nil, fmt.Errorf("can't define xml block with csv format")
				}
				format.Csv = makeCsvParams(csvBlock)
			case "xml":
				if csvBlockIsSet {
					return nil, fmt.Errorf("can't define csv block with xml format")
				}
				format.Xml = makeXmlParams(xmlBlock)
			}

			sources = append(sources, source)
		}
	}

	return &sources, nil
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

func flattenXmlParams(params *openapi.XmlParams) []interface{} {
	m := make(map[string]interface{})
	m["root_tag"] = *params.RootTag
	m["encoding"] = *params.Encoding
	m["doc_tag"] = *params.DocTag
	m["value_tag"] = *params.ValueTag
	m["attribute_prefix"] = *params.AttributePrefix

	return []interface{}{m}
}

func makeXmlParams(in interface{}) *openapi.XmlParams {
	m := openapi.XmlParams{}

	for _, i := range in.(*schema.Set).List() {
		if val, ok := i.(map[string]interface{}); ok {
			for k, v := range val {
				switch k {
				case "root_tag":
					m.RootTag = toStringPtrNilIfEmpty(v.(string))
				case "encoding":
					m.Encoding = toStringPtrNilIfEmpty(v.(string))
				case "doc_tag":
					m.DocTag = toStringPtrNilIfEmpty(v.(string))
				case "value_tag":
					m.ValueTag = toStringPtrNilIfEmpty(v.(string))
				case "attribute_prefix":
					m.AttributePrefix = toStringPtrNilIfEmpty(v.(string))
				}
			}
		}
	}

	return &m
}

// shared between s3 and gcs collections
func formatSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeString,
		ForceNew: true,
		Required: true,
		ValidateFunc: validation.StringMatch(
			regexp.MustCompile("^(json|csv|xml)$"), "only 'json', 'xml', or 'csv' is supported"),
		Description: "Format of the data. One of: json, csv, xml. xml and csv blocks can only be set for their respective formats. ",
	}
}

// shared between s3 and gcs collections
func csvSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		ForceNew: true,
		Optional: true,
		MinItems: 0,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"first_line_as_column_names": {
					Type:        schema.TypeBool,
					ForceNew:    true,
					Optional:    true,
					Default:     true,
					Description: "If the first line in every object specifies the column names.",
				},
				"separator": {
					Type:        schema.TypeString,
					ForceNew:    true,
					Optional:    true,
					Default:     ",",
					Description: "A single character that is the column separator.",
				},
				"encoding": {
					Type:        schema.TypeString,
					ForceNew:    true,
					Optional:    true,
					Default:     "UTF-8",
					Description: "Can be one of: UTF-8, ISO_8859_1, UTF-16.",
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
					Description: "The names of the columns.",
				},
				"column_types": {
					Type:     schema.TypeList,
					ForceNew: true,
					Optional: true,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
					Description: "The types of the columns.",
				},
				"quote_char": {
					Type:        schema.TypeString,
					ForceNew:    true,
					Optional:    true,
					Default:     `"`,
					Description: "Character within which a cell value is enclosed. Defaults to double quote.",
				},
				"escape_char": {
					Type:        schema.TypeString,
					ForceNew:    true,
					Optional:    true,
					Default:     `\`,
					Description: "Escape character removes any special meaning from the character that follows it . Defaults to backslash.",
				},
			},
		},
	}
}

// shared between s3 and gcs collections
func xmlSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		ForceNew: true,
		Optional: true,
		MinItems: 0,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"root_tag": {
					Description: "Tag until which xml is ignored.",
					Type:        schema.TypeString,
					ForceNew:    true,
					Optional:    true,
				},
				"encoding": {
					Description: "Encoding in which data source is encoded.",
					Type:        schema.TypeString,
					ForceNew:    true,
					Optional:    true,
					Default:     "UTF-8",
					ValidateFunc: validation.StringMatch(
						regexp.MustCompile("^(UTF-8|UTF-16|ISO_8859_1)$"), "must be either 'UTF-8', 'UTF-16' or 'ISO_8859_1'"),
				},
				"doc_tag": {
					Description: "Tags with which documents are identified",
					Type:        schema.TypeString,
					ForceNew:    true,
					Optional:    true,
				},
				"value_tag": {
					Description: "Tag used for the value when there are attributes in the element having no child.",
					Type:        schema.TypeString,
					ForceNew:    true,
					Optional:    true,
					Default:     "value", // API sets this implicitly, if we don't match we get diffs
				},
				"attribute_prefix": {
					Description: "Tag to differentiate between attributes and elements.",
					Type:        schema.TypeString,
					ForceNew:    true,
					Optional:    true,
				},
			},
		},
	}
}

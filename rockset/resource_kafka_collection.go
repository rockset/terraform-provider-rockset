package rockset

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/rockset/rockset-go-client"
	"github.com/rockset/rockset-go-client/openapi"
	"github.com/rockset/rockset-go-client/option"
	"regexp"
)

func kafkaCollectionSchema() map[string]*schema.Schema {
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
						Description:  "The name of the Rockset Kafka integration.",
						Type:         schema.TypeString,
						ForceNew:     true,
						Required:     true,
						ValidateFunc: rocksetNameValidator,
					},
					"topic_name": {
						Description: "Name of Kafka topic to be tailed.",
						Type:        schema.TypeString,
						ForceNew:    true,
						Required:    true,
					},
					"consumer_group_id": {
						Description: "The Kafka consumer group Id being used.",
						Type:        schema.TypeString,
						Computed:    true,
					},
					"use_v3": {
						Type:        schema.TypeBool,
						ForceNew:    true,
						Optional:    true,
						Default:     false,
						Description: "Whether to use v3 integration. Required if the kafka integration uses v3.",
					},
					"offset_reset_policy": {
						Description: "The offset reset policy. Possible values: LATEST, EARLIEST. Only valid with v3 collections.",
						Type:        schema.TypeString,
						ForceNew:    true,
						Optional:    true,
						ValidateFunc: validation.StringMatch(
							regexp.MustCompile("^(LATEST|EARLIEST)$"), "only 'LATEST' or 'EARLIEST' is supported"),
					},
					"status": {
						Type:     schema.TypeList,
						Computed: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"state": {
									Description: "State of the Kafka source. Possible values: NO_DOCS_YET, ACTIVE, DORMANT.",
									Type:        schema.TypeString,
									Computed:    true,
								},
								"last_consumed_time": {
									Description: "The type of partitions on a field.",
									Type:        schema.TypeString,
									Computed:    true,
								},
								"documents_processed": {
									Description: "Number of documents processed by this Kafka topic.",
									Type:        schema.TypeInt,
									Computed:    true,
								},
								"partitions": {
									Description: "The status info per partition.",
									Type:        schema.TypeSet,
									Computed:    true,
									Elem: &schema.Resource{
										Schema: map[string]*schema.Schema{
											"partition_number": {
												Description: "The number of this partition.",
												Type:        schema.TypeInt,
												Computed:    true,
											},
											"partition_offset": {
												Description: "Latest offset of this partition.",
												Type:        schema.TypeInt,
												Computed:    true,
											},
											"offset_lag": {
												Description: "Per partition lag for offset.",
												Type:        schema.TypeInt,
												Computed:    true,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	} // End schema return
} // End func

func resourceKafkaCollection() *schema.Resource {
	return &schema.Resource{
		Description: "Manages a collection created from a Kafka source. " +
			"The `use_v3` field must match the integration which the collection is created from.",

		CreateContext: resourceKafkaCollectionCreate,
		ReadContext:   resourceKafkaCollectionRead,
		DeleteContext: resourceCollectionDelete, // No change from base collection delete

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: func(ctx context.Context, diff *schema.ResourceDiff, i interface{}) error {
			// TODO use_v3=true validation
			return nil
		},

		// This schema will use the base collection schema as a foundation
		// And layer on just the necessary fields for a Kafka collection
		Schema: mergeSchemas(baseCollectionSchema(), kafkaCollectionSchema()),
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(defaultCollectionTimeout),
		},
	}
}

func resourceKafkaCollectionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rc := meta.(*rockset.RockClient)
	var diags diag.Diagnostics
	var err error

	name := d.Get("name").(string)
	workspace := d.Get("workspace").(string)

	tflog.Debug(ctx, "create kafka collection", map[string]interface{}{
		"workspace": workspace,
		"name":      name,
	})

	// Add all base schema fields
	params := createBaseCollectionRequest(d)
	// Add fields for Kafka
	params.Sources = expandKafkaSourceParams(d.Get("source"))

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

func resourceKafkaCollectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	// Gets all the fields relevant to a Kafka collection
	err = parseKafkaCollection(&collection, d)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

// parseKafkaCollection takes in a collection returned from the api. Parses the fields relevant to a Kafka source and
// puts them into the schema object.
func parseKafkaCollection(collection *openapi.Collection, d *schema.ResourceData) error {
	var err error

	sourcesList := collection.Sources
	sourcesCount := len(sourcesList)
	if sourcesCount < 1 {
		return fmt.Errorf("expected %s to have at least 1 source", collection.GetName())
	}

	err = d.Set("source", flattenKafkaSourceParams(&sourcesList))
	if err != nil {
		return err
	}

	return nil // No errors
}

func flattenKafkaSourceParams(sources *[]openapi.Source) []interface{} {
	convertedList := make([]interface{}, 0, len(*sources))
	for _, source := range *sources {
		m := make(map[string]interface{})
		m["integration_name"] = source.IntegrationName
		m["topic_name"] = source.Kafka.KafkaTopicName
		m["consumer_group_id"] = source.Kafka.ConsumerGroupId
		m["offset_reset_policy"] = source.Kafka.OffsetResetPolicy
		m["use_v3"] = source.Kafka.UseV3

		m["status"] = flattenKafkaSourceStatus(source.Kafka.Status)
		convertedList = append(convertedList, m)
	}

	return convertedList
}

func flattenKafkaSourceStatus(status *openapi.StatusKafka) []interface{} {
	m := make(map[string]interface{})

	if v, ok := status.GetStateOk(); ok {
		m["state"] = v
	}

	if v, ok := status.GetLastConsumedTimeOk(); ok {
		m["last_consumed_time"] = v
	}

	if v, ok := status.GetNumDocumentsProcessedOk(); ok {
		m["documents_processed"] = v
	}

	if len(status.KafkaPartitions) > 0 {
		m["partitions"] = flattenKafkaSourcePartitions(status.KafkaPartitions)
	}

	return []interface{}{m}
}

func flattenKafkaSourcePartitions(statuses []openapi.StatusKafkaPartition) []interface{} {
	convertedList := make([]interface{}, 0, len(statuses))
	for _, status := range statuses {
		convertedList = append(convertedList, map[string]interface{}{
			"partition_number": status.PartitionNumber,
			"partition_offset": status.PartitionOffset,
			"offset_lag":       status.OffsetLag,
		})
	}

	return convertedList
}

func expandKafkaSourceParams(in interface{}) []openapi.Source {
	set := in.(*schema.Set)
	sources := make([]openapi.Source, 0, set.Len())

	for _, i := range set.List() {
		if val, ok := i.(map[string]interface{}); ok {
			source := openapi.Source{}
			source.Kafka = openapi.NewSourceKafkaWithDefaults()
			format := openapi.FormatParams{}
			source.FormatParams = &format

			for k, v := range val {
				switch k {
				case "integration_name":
					source.IntegrationName = toStringPtrNilIfEmpty(v.(string))
				case "topic_name":
					source.Kafka.KafkaTopicName = toStringPtrNilIfEmpty(v.(string))
				case "offset_reset_policy":
					source.Kafka.OffsetResetPolicy = toStringPtrNilIfEmpty(v.(string))
				case "use_v3":
					if v != nil {
						source.Kafka.UseV3 = openapi.PtrBool(v.(bool))
					}
				case "consumer_group_id":
					source.Kafka.ConsumerGroupId = toStringPtrNilIfEmpty(v.(string))
				}
			}
			sources = append(sources, source)
		}
	}

	return sources
}

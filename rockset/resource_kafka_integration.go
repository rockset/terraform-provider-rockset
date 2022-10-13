package rockset

import (
	"context"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/rockset/rockset-go-client"
	"github.com/rockset/rockset-go-client/option"
)

func resourceKafkaIntegration() *schema.Resource {
	return &schema.Resource{
		Description: "Manages a Rockset Kafka Integration.\n\n" +
			"If the integration is connected with Confluent Cloud, there is a " +
			"[Terraform provider](https://registry.terraform.io/providers/confluentinc/confluent/latest/docs) " +
			"which can be used to configure the Confluent Cloud side of the integration.",

		// No updatable fields at this time, all fields require recreation.
		CreateContext: resourceKafkaIntegrationCreate,
		ReadContext:   resourceKafkaIntegrationRead,
		DeleteContext: resourceIntegrationDelete, // common among <type>integrations

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Description:  "Unique identifier for the integration. Can contain alphanumeric or dash characters.",
				Type:         schema.TypeString,
				ForceNew:     true,
				Required:     true,
				ValidateFunc: rocksetNameValidator,
			},
			"description": {
				Description: "Text describing the integration.",
				Type:        schema.TypeString,
				Default:     "created by Rockset terraform provider",
				ForceNew:    true,
				Optional:    true,
			},
			// v2
			"kafka_topic_names": {
				Description:   "Kafka topics to tail.",
				Type:          schema.TypeSet,
				ForceNew:      true,
				Optional:      true,
				ConflictsWith: []string{"bootstrap_servers", "security_config", "schema_registry_config"},
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"kafka_data_format": {
				Description:   "The format of the Kafka topics being tailed. Possible values: JSON, AVRO.",
				Type:          schema.TypeString,
				ForceNew:      true,
				Optional:      true,
				ConflictsWith: []string{"bootstrap_servers", "security_config", "schema_registry_config"},
			},
			"connection_string": {
				Type:     schema.TypeString,
				Optional: true,
				//ForceNew:      true,
				Description: "Kafka connection string.",
				Computed:    true,
				//ConflictsWith: []string{"bootstrap_servers", "security_config", "schema_registry_config"},
				// Sensitive:     true,
				// TODO: can't be sensitive as it is needed to configure kafka-connect
			},
			"use_v3": {
				Description: "Use v3 for Confluent Cloud.",
				Type:        schema.TypeBool,
				ForceNew:    true,
				Optional:    true,
				Default:     false,
			},
			// v3 below
			"bootstrap_servers": {
				Description:   "The Kafka bootstrap server url(s). Required only for V3 integration.",
				Type:          schema.TypeString,
				ForceNew:      true,
				Optional:      true,
				RequiredWith:  []string{"use_v3", "security_config"},
				ConflictsWith: []string{"kafka_topic_names", "kafka_data_format", "connection_string"},
			},
			"security_config": {
				Description:      "Kafka security configurations. Required only for V3 integration.",
				Type:             schema.TypeMap,
				ForceNew:         true,
				Optional:         true,
				RequiredWith:     []string{"use_v3", "bootstrap_servers"},
				ConflictsWith:    []string{"kafka_topic_names", "kafka_data_format", "connection_string"},
				DiffSuppressFunc: secretDiffSuppress("security_config.secret"),
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"schema_registry_config": {
				Description:      "Kafka configuration for schema registry. Required only for V3 integration.",
				Type:             schema.TypeMap,
				ForceNew:         true,
				Optional:         true,
				RequiredWith:     []string{"use_v3", "bootstrap_servers", "security_config"},
				DiffSuppressFunc: secretDiffSuppress("schema_registry_config.secret"),
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"wait_for_integration": {
				Description: "Wait until the integration is active.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				ForceNew:    true,
			},
		},
	}
}

func secretDiffSuppress(key string) func(string, string, string, *schema.ResourceData) bool {
	return func(k, oldValue, newValue string, d *schema.ResourceData) bool {
		if k != key {
			return oldValue == newValue
		}

		// special case for security_config.secret to compare only the first 4 characters
		// as the Rockset API only return those for security reasons
		if oldValue != "" && newValue != "" && oldValue[:4] == newValue[:4] {
			return true
		}

		return oldValue == newValue
	}
}

func resourceKafkaIntegrationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rc := meta.(*rockset.RockClient)
	var diags diag.Diagnostics

	var opts []option.KafkaIntegrationOption

	name := d.Get("name").(string)
	tflog.Info(ctx, "creating kafka integration", map[string]interface{}{
		"name": name,
	})

	if v, ok := d.GetOk("description"); ok {
		opts = append(opts, option.WithKafkaIntegrationDescription(v.(string)))
	}

	//if v, ok := d.GetOk("connection_string"); ok {
	//	opts = append(opts, option.WithKafkaConnectionString(v.(string)))
	//}

	if v, ok := d.GetOk("bootstrap_servers"); ok {
		opts = append(opts, option.WithKafkaBootstrapServers(v.(string)))
	}

	if v, ok := d.GetOk("security_config"); ok {
		m := v.(map[string]interface{})
		opts = append(opts, option.WithKafkaSecurityConfig(m["api_key"].(string), m["secret"].(string)))
	}

	if v, ok := d.GetOk("schema_registry_config"); ok {
		m := v.(map[string]interface{})
		opts = append(opts, option.WithKafkaSchemaRegistryConfig(m["url"].(string), m["key"].(string), m["secret"].(string)))
	}

	if v, ok := d.GetOk("kafka_topic_names"); ok {
		topics := v.(*schema.Set)
		for _, t := range topics.List() {
			opts = append(opts, option.WithKafkaIntegrationTopic(t.(string)))
		}
	}

	if v, ok := d.GetOk("kafka_data_format"); ok {
		var format option.KafkaFormat
		switch v {
		case option.KafkaFormatJSON.String():
			format = option.KafkaFormatJSON
		case option.KafkaFormatAVRO.String():
			format = option.KafkaFormatAVRO
		default:
			return diag.Errorf("unknown format: %s", v)
		}
		opts = append(opts, option.WithKafkaDataFormat(format))
	}

	var v3 bool
	if v, ok := d.GetOk("use_v3"); ok {
		if v3 = v.(bool); v3 {
			opts = append(opts, option.WithKafkaV3())
		}
	}

	r, err := rc.CreateKafkaIntegration(ctx, name, opts...)
	if err != nil {
		return diag.FromErr(err)
	}

	tflog.Info(ctx, "integration created", map[string]interface{}{
		"name": name,
	})

	if v3 {
		if wait := d.Get("wait_for_integration").(bool); wait {
			tflog.Debug(ctx, "waiting for integration", map[string]interface{}{
				"name": name,
			})
			if err = rc.WaitUntilKafkaIntegrationActive(ctx, name); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	d.SetId(r.GetName())

	if cs, ok := r.Kafka.GetConnectionStringOk(); ok {
		if err = d.Set("connection_string", cs); err != nil {
			return diag.FromErr(err)
		}
		tflog.Trace(ctx, "set connection_string", map[string]interface{}{
			"value": cs,
		})
	}

	tflog.Debug(ctx, "create integration done", map[string]interface{}{})

	return diags
}

func resourceKafkaIntegrationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rc := meta.(*rockset.RockClient)
	var diags diag.Diagnostics

	name := d.Id()

	tflog.Info(ctx, "reading kafka integration", map[string]interface{}{
		"name": name,
	})

	response, err := rc.GetIntegration(ctx, name)
	if err != nil {
		return checkForNotFoundError(d, err)
	}

	if err = d.Set("name", response.Name); err != nil {
		return diag.FromErr(err)
	}

	if err = d.Set("description", response.Description); err != nil {
		return diag.FromErr(err)
	}

	if err = d.Set("kafka_data_format", response.Kafka.KafkaDataFormat); err != nil {
		return diag.FromErr(err)
	}

	if err = d.Set("use_v3", response.Kafka.UseV3); err != nil {
		return diag.FromErr(err)
	}

	if topics, ok := response.Kafka.GetKafkaTopicNamesOk(); ok {
		if err = d.Set("kafka_topic_names", topics); err != nil {
			return diag.FromErr(err)
		}
	}

	if err = d.Set("bootstrap_servers", response.Kafka.BootstrapServers); err != nil {
		return diag.FromErr(err)
	}

	if response.Kafka.SecurityConfig != nil {
		err = d.Set("security_config", map[string]any{
			"api_key": response.Kafka.SecurityConfig.ApiKey,
			"secret":  response.Kafka.SecurityConfig.Secret,
		})
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if response.Kafka.SchemaRegistryConfig != nil {
		err = d.Set("schema_registry_config", map[string]any{
			"url":    response.Kafka.SchemaRegistryConfig.Url,
			"key":    response.Kafka.SchemaRegistryConfig.Key,
			"secret": response.Kafka.SchemaRegistryConfig.Secret,
		})
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if cs, ok := response.Kafka.GetConnectionStringOk(); ok {
		if err = d.Set("connection_string", cs); err != nil {
			return diag.FromErr(err)
		}
	}

	return diags
}

package rockset

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/rockset/rockset-go-client"
	"github.com/rockset/rockset-go-client/option"
)

func resourceDynamoDBIntegration() *schema.Resource {
	return &schema.Resource{
		Description: "Manages a Rockset DynamoDB Integration.",

		// No updateable fields at this time, all fields require recreation.
		CreateContext: resourceDynamoDBIntegrationCreate,
		ReadContext:   resourceDynamoDBIntegrationRead,
		DeleteContext: resourceIntegrationDelete, // common among <type>integrations

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Description:  "Unique identifier for the integration. Can contain alphanumeric or dash characters.",
				Type:         schema.TypeString,
				ForceNew:     true,
				Required:     true,
				ValidateFunc: rocksetNameValidator,
			},
			"description": &schema.Schema{
				Description: "Text describing the integration.",
				Type:        schema.TypeString,
				Default:     "created by Rockset terraform provider",
				ForceNew:    true,
				Optional:    true,
			},
			"aws_role_arn": &schema.Schema{
				Description: "The AWS Role Arn to use for this integration.",
				Type:        schema.TypeString,
				ForceNew:    true,
				Required:    true,
			},
		},
	}
}

func resourceDynamoDBIntegrationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rc := meta.(*rockset.RockClient)
	var diags diag.Diagnostics

	r, err := rc.CreateDynamoDBIntegration(ctx, d.Get("name").(string),
		option.AWSRole(d.Get("aws_role_arn").(string)),
		option.WithDynamoDBIntegrationDescription(d.Get("description").(string)))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(r.GetName())

	return diags
}

func resourceDynamoDBIntegrationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rc := meta.(*rockset.RockClient)
	var diags diag.Diagnostics

	name := d.Id()

	integration, err := rc.GetIntegration(ctx, name)
	if err != nil {
		return diag.FromErr(err)
	}

	_ = d.Set("name", integration.Name)
	_ = d.Set("description", integration.Description)
	_ = d.Set("aws_role_arn", integration.Dynamodb.AwsRole.AwsRoleArn)

	return diags
}

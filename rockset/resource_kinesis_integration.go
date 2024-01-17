package rockset

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/rockset/rockset-go-client"
	"github.com/rockset/rockset-go-client/option"
)

func resourceKinesisIntegration() *schema.Resource {
	return &schema.Resource{
		Description: "Manages a Rockset Kinesis Integration.",

		// No updatable fields at this time, all fields require recreation.
		CreateContext: resourceKinesisIntegrationCreate,
		ReadContext:   resourceKinesisIntegrationRead,
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
			"aws_role_arn": {
				Description: "The AWS Role Arn to use for this integration.",
				Type:        schema.TypeString,
				ForceNew:    true,
				Required:    true,
			},
		},
	}
}

func resourceKinesisIntegrationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rc := meta.(*rockset.RockClient)
	var diags diag.Diagnostics

	r, err := rc.CreateKinesisIntegration(ctx, d.Get("name").(string),
		option.AWSRole(d.Get("aws_role_arn").(string)),
		option.WithKinesisIntegrationDescription(d.Get("description").(string)))
	// TODO: retry if we get an error from AWS
	//   Authentication failed for AWS cross-account role integration with Role ARN
	//   as it can be due to the role taking a few seconds to propagate
	if err != nil {
		return DiagFromErr(err)
	}

	d.SetId(r.GetName())

	return diags
}

func resourceKinesisIntegrationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rc := meta.(*rockset.RockClient)
	var diags diag.Diagnostics

	name := d.Id()

	response, err := rc.GetIntegration(ctx, name)
	if err != nil {
		return checkForNotFoundError(d, err)
	}

	_ = d.Set("name", response.Name)
	_ = d.Set("description", response.Description)
	_ = d.Set("aws_role_arn", response.Kinesis.AwsRole.AwsRoleArn)

	return diags
}

package rockset

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/rockset/rockset-go-client"
	"github.com/rockset/rockset-go-client/option"
)

func resourceS3Integration() *schema.Resource {
	return &schema.Resource{
		Description: "Manages a Rockset S3 Integration.",

		// No updatable fields at this time, all fields require recreation.
		CreateContext: resourceS3IntegrationCreate,
		ReadContext:   resourceS3IntegrationRead,
		DeleteContext: resourceIntegrationDelete,

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

func resourceS3IntegrationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rc := meta.(*rockset.RockClient)
	var diags diag.Diagnostics

	r, err := rc.CreateS3Integration(ctx, d.Get("name").(string),
		option.AWSRole(d.Get("aws_role_arn").(string)),
		option.WithS3IntegrationDescription(d.Get("description").(string)))
	// TODO: retry if we get an error from AWS
	//   Authentication failed for AWS cross-account role integration with Role ARN
	//   as it can be due to the role taking a few seconds to propagate
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(r.GetName())

	return diags
}

func resourceS3IntegrationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rc := meta.(*rockset.RockClient)
	var diags diag.Diagnostics

	name := d.Id()

	response, err := rc.GetIntegration(ctx, name)
	if err != nil {
		return checkForNotFoundError(d, err)
	}

	_ = d.Set("name", response.Name)
	_ = d.Set("description", response.Description)
	_ = d.Set("aws_role_arn", response.S3.AwsRole.AwsRoleArn)

	return diags
}

func resourceIntegrationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rc := meta.(*rockset.RockClient)
	var diags diag.Diagnostics

	name := d.Id()

	err := rc.DeleteIntegration(ctx, name)
	if err != nil {
		return diag.FromErr(err)
	}

	// d.SetId("") is automatically called assuming delete returns no errors,
	// but it is added here for explicitness.
	d.SetId("")

	return diags
}

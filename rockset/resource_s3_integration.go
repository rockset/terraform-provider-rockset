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

		// No updateable fields at this time, all fields require recreation.
		CreateContext: resourceS3IntegrationCreate,
		ReadContext:   resourceS3IntegrationRead,
		DeleteContext: resourceS3IntegrationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:         schema.TypeString,
				ForceNew:     true,
				Required:     true,
				ValidateFunc: rocksetNameValidator,
			},
			"description": &schema.Schema{
				Type:     schema.TypeString,
				Default:  "created by Rockset terraform provider",
				ForceNew: true,
				Optional: true,
			},
			"aws_role_arn": &schema.Schema{
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
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
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(r.Data.GetName())

	return diags
}

func resourceS3IntegrationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rc := meta.(*rockset.RockClient)
	var diags diag.Diagnostics

	name := d.Id()

	getReq := rc.IntegrationsApi.GetIntegration(ctx, name)
	response, _, err := getReq.Execute()
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("name", response.Data.Name)
	d.Set("description", response.Data.Description)
	d.Set("aws_role_arn", response.Data.S3.AwsRole.AwsRoleArn)

	return diags
}

func resourceS3IntegrationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

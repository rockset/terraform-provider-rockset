package rockset

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/rockset/rockset-go-client"
	"github.com/rockset/rockset-go-client/openapi"
	"github.com/rockset/rockset-go-client/option"
)

func resourceKinesisIntegration() *schema.Resource {
	return &schema.Resource{
		Description: "Manages a Rockset Kinesis Integration.",

		// No updateable fields at this time, all fields require recreation.
		CreateContext: resourceKinesisIntegrationCreate,
		ReadContext:   resourceKinesisIntegrationRead,
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
			"aws_role_arn": {
				Description: "The AWS Role Arn to use for this integration.",
				Type:        schema.TypeString,
				ForceNew:    true,
				Required:    true,
			},
		},
	}
}

// TODO: Replace once fixed in go client
func CreateKinesisIntegration(rc *rockset.RockClient, ctx context.Context, name string, creds option.AWSCredentialsFn,
	options ...option.KinesisIntegrationOption) (openapi.Integration, error) {
	var err error
	var resp openapi.CreateIntegrationResponse
	q := rc.IntegrationsApi.CreateIntegration(ctx)
	req := openapi.NewCreateIntegrationRequest(name)

	c := option.AWSCredentials{}
	creds(&c)

	opts := option.KinesisIntegration{}
	for _, o := range options {
		o(&opts)
	}

	req.Kinesis = &openapi.KinesisIntegration{}
	if opts.Description != nil {
		req.Description = opts.Description
	}
	if c.AwsRole != nil {
		req.Kinesis.AwsRole = c.AwsRole
	}

	err = rc.Retry(ctx, func() error {
		resp, _, err = q.Body(*req).Execute()
		return err
	})

	if err != nil {
		return openapi.Integration{}, err
	}

	return resp.GetData(), nil
}

func resourceKinesisIntegrationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rc := meta.(*rockset.RockClient)
	var diags diag.Diagnostics

	// TODO: replace once go client is fixed
	r, err := CreateKinesisIntegration(rc, ctx, d.Get("name").(string),
		option.AWSRole(d.Get("aws_role_arn").(string)),
		option.WithKinesisIntegrationDescription(d.Get("description").(string)))

	// r, err := rc.CreateKinesisIntegration(ctx, d.Get("name").(string),
	// 	option.AWSRole(d.Get("aws_role_arn").(string)),
	// 	option.WithKinesisIntegrationDescription(d.Get("description").(string)))
	if err != nil {
		return diag.FromErr(err)
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
		return diag.FromErr(err)
	}

	_ = d.Set("name", response.Name)
	_ = d.Set("description", response.Description)
	_ = d.Set("aws_role_arn", response.Kinesis.AwsRole.AwsRoleArn)

	return diags
}
package rockset

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/rockset/rockset-go-client"
	"github.com/rockset/rockset-go-client/option"
)

func resourceMongoDBIntegration() *schema.Resource {
	return &schema.Resource{
		Description: "Manages a Rockset MongoDB Integration.",

		// No updatable fields at this time, all fields require recreation.
		CreateContext: resourceMongoDBIntegrationCreate,
		ReadContext:   resourceMongoDBIntegrationRead,
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
			"connection_uri": {
				Description: "MongoDB connection URI string. The password is scrubbed from the URI when fetched " +
					"by the API so this field is NOT set on imports and reads.",
				Type:      schema.TypeString,
				ForceNew:  true,
				Required:  true,
				Sensitive: true,
			},
		},
	}
}

func resourceMongoDBIntegrationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rc := meta.(*rockset.RockClient)
	var diags diag.Diagnostics

	r, err := rc.CreateMongoDBIntegration(ctx, d.Get("name").(string), d.Get("connection_uri").(string),
		option.WithMongoDBIntegrationDescription(d.Get("description").(string)))
	if err != nil {
		return DiagFromErr(err)
	}

	d.SetId(r.GetName())

	return diags
}

func resourceMongoDBIntegrationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rc := meta.(*rockset.RockClient)
	var diags diag.Diagnostics

	name := d.Id()

	response, err := rc.GetIntegration(ctx, name)
	if err != nil {
		return checkForNotFoundError(d, err)
	}

	_ = d.Set("name", response.Name)
	_ = d.Set("description", response.Description)
	// We cannot read the connection URI here. The API sanitizes it and removes secrets.

	return diags
}

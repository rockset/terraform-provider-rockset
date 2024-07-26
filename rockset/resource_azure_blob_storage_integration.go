package rockset

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/rockset/rockset-go-client"
)

func resourceAzureBlobStorageIntegration() *schema.Resource {
	return &schema.Resource{
		Description: "Manages a Rockset Azure Blob Storage Integration.",

		// No updatable fields at this time, all fields require recreation.
		CreateContext: resourceAzureBlobStorageIntegrationCreate,
		ReadContext:   resourceAzureBlobStorageIntegrationRead,
		DeleteContext: resourceAzureBlobStorageIntegrationDelete,

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
			"connection_string": {
				Description: "The connection string to use for the storage account or the container in the Azure Blob Store",
				Type:        schema.TypeString,
				ForceNew:    true,
				Required:    true,
			},
		},
	}
}

func resourceAzureBlobStorageIntegrationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rc := meta.(*rockset.RockClient)
	var diags diag.Diagnostics

	r, err := rc.CreateAzureBlobStorageIntegration(ctx, d.Get("name").(string),
		d.Get("connection_string").(string))
	// TODO: retry if we get an error from Azure
	//   Authentication failed for Azure cross-account role integration with connection_string
	//   as it can be due to the role taking a few seconds to propagate
	if err != nil {
		return DiagFromErr(err)
	}

	d.SetId(r.GetName())

	return diags
}

func resourceAzureBlobStorageIntegrationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rc := meta.(*rockset.RockClient)
	var diags diag.Diagnostics

	name := d.Id()

	response, err := rc.GetIntegration(ctx, name)
	if err != nil {
		return checkForNotFoundError(d, err)
	}

	_ = d.Set("name", response.Name)
	_ = d.Set("connection_string", response.AzureBlobStorage.ConnectionString)

	return diags
}

func resourceAzureBlobStorageIntegrationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rc := meta.(*rockset.RockClient)
	var diags diag.Diagnostics

	name := d.Id()

	err := rc.DeleteIntegration(ctx, name)
	if err != nil {
		return DiagFromErr(err)
	}

	// d.SetId("") is automatically called assuming delete returns no errors,
	// but it is added here for explicitness.
	d.SetId("")

	return diags
}

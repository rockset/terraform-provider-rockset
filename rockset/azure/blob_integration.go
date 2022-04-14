package azure

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/rockset/terraform-provider-rockset/rockset/validators"
)

func BlobIntegration() *schema.Resource {
	return &schema.Resource{
		Description: "Manages a Rockset Azure Blob Storage Integration",

		// No updateable fields at this time, all fields require recreation.
		CreateContext: createBlobIntegration,
		ReadContext:   readBlobIntegration,
		DeleteContext: deleteBlobIntegration,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Description:  "Unique identifier for the integration. Can contain alphanumeric or dash characters.",
				Type:         schema.TypeString,
				ForceNew:     true,
				Required:     true,
				ValidateFunc: validators.NameValidator,
			},
			"description": {
				Description: "Text describing the integration.",
				Type:        schema.TypeString,
				Default:     "created by Rockset terraform provider",
				ForceNew:    true,
				Optional:    true,
			},
			"credential_string": {
				Description: "The Azure Blob Connection String to use for this integration.",
				Type:        schema.TypeString,
				ForceNew:    true,
				Required:    true,
			},
		},
	}
}

func createBlobIntegration(ctx context.Context, rd *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

func readBlobIntegration(ctx context.Context, rd *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

func deleteBlobIntegration(ctx context.Context, rd *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

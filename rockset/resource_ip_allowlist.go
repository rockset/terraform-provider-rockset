package rockset

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceIpAllowlist() *schema.Resource {
	return &schema.Resource{
		Description: "Sample resource in the Terraform provider IpAllowlist.",

		CreateContext: resourceIpAllowlistCreate,
		ReadContext:   resourceIpAllowlistRead,
		UpdateContext: resourceIpAllowlistUpdate,
		DeleteContext: resourceIpAllowlistDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"sample_attribute": {
				Description: "Sample attribute.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"last_updated": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func resourceIpAllowlistCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	//rc := meta.(*rockset.RockClient)
	//var diags diag.Diagnostics

	idFromAPI := "my-id"
	d.SetId(idFromAPI)

	return diag.Errorf("not implemented")
}

func resourceIpAllowlistRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	//rc := meta.(*rockset.RockClient)
	//var diags diag.Diagnostics

	return diag.Errorf("not implemented")
}

func resourceIpAllowlistUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	//rc := meta.(*rockset.RockClient)
	//var diags diag.Diagnostics

	return diag.Errorf("not implemented")
}

func resourceIpAllowlistDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	//rc := meta.(*rockset.RockClient)
	//var diags diag.Diagnostics

	return diag.Errorf("not implemented")
}

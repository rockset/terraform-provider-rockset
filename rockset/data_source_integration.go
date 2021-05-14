package rockset

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceRocksetIntegration() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceReadRocksetIntegration,

		Schema: map[string]*schema.Schema{
			"sample_attribute": {
				Description: "Sample attribute.",
				Type:        schema.TypeString,
				Optional:    true,
			},
		}}
}

func dataSourceReadRocksetIntegration(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	//rc := meta.(*rockset.RockClient)
	//var diags diag.Diagnostics

	return diag.Errorf("not implemented")
}

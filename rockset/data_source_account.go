package rockset

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/rockset/rockset-go-client"
)

func dataSourceRocksetAccount() *schema.Resource {
	return &schema.Resource{
		Description: "Gets information about the Rockset deployment for the specified api server.",
		ReadContext: dataSourceReadRocksetAccount,

		Schema: map[string]*schema.Schema{
			"external_id": &schema.Schema{
				Description: "The external ID to use in AWS trust policies.",
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
			},
			"account_id": &schema.Schema{
				Description: "The AWS account ID to reference in AWS policies.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		}}
}

const accountID = "318212636800"

func dataSourceReadRocksetAccount(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rc := meta.(*rockset.RockClient)
	var diags diag.Diagnostics

	err := d.Set("account_id", accountID)
	if err != nil {
		return diag.FromErr(err)
	}

	org, err := rc.GetOrganization(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("external_id", org.ExternalId)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(accountID)

	return diags
}

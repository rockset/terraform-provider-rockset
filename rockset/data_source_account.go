package rockset

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/rockset/rockset-go-client"
	"github.com/rockset/rockset-go-client/openapi"
)

func dataSourceRocksetAccount() *schema.Resource {
	return &schema.Resource{
		Description: "Gets information about the Rockset deployment for the specified api server.",
		ReadContext: dataSourceReadRocksetAccount,

		Schema: map[string]*schema.Schema{
			"external_id": {
				Description: "The external ID to use in AWS trust policies.",
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
			},
			"account_id": {
				Description: "The AWS account ID to reference in AWS policies.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"organization": {
				Description: "The name of the organization for the API key.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"company": {
				Description: "The name of the company for the API key.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"rockset_user": {
				Description: "The name of the Rockset user used for AWS trust policies.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"clusters": {
				Description: "The Rockset clusters available to this API key.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"aws_region": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"api_server": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

const accountID = "318212636800"

func dataSourceReadRocksetAccount(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rc := meta.(*rockset.RockClient)
	var diags diag.Diagnostics

	org, err := rc.GetOrganization(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	if err = d.Set("account_id", accountID); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("external_id", org.ExternalId); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("organization", org.DisplayName); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("rockset_user", org.RocksetUser); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("company", org.CompanyName); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("clusters", flattenClusterParams(*org.Clusters)); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(accountID)

	return diags
}

func flattenClusterParams(clusters []openapi.Cluster) []interface{} {
	out := make([]interface{}, len(clusters))

	for i, c := range clusters {
		m := make(map[string]interface{})

		m["type"] = *c.ClusterType
		m["aws_region"] = *c.AwsRegion
		m["api_server"] = *c.ApiserverUrl

		out[i] = m
	}

	return out
}

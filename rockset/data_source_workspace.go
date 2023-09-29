package rockset

import (
	"context"

	"github.com/rockset/rockset-go-client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceRocksetWorkspace() *schema.Resource {
	return &schema.Resource{
		Description: "Gets information about a workspace.",
		ReadContext: dataSourceReadRocksetWorkspace,

		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The workspace `name`.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"name": {
				Description: "Workspace name.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"description": {
				Description: "Workspace description.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"created_by": {
				Description: "Created by.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"created_at": {
				Description: "Created at in ISO-8601.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"collection_count": {
				Description: "Number of collections in the workspace.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
		}}
}

func dataSourceReadRocksetWorkspace(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rc := meta.(*rockset.RockClient)
	var diags diag.Diagnostics

	name := d.Get("name").(string)

	ws, err := rc.GetWorkspace(ctx, name)
	if err != nil {
		return diag.FromErr(err)
	}

	if err = d.Set("name", ws.GetName()); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("description", ws.GetDescription()); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("created_by", ws.GetCreatedBy()); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("created_at", ws.GetCreatedAt()); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("collection_count", ws.GetCollectionCount()); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(name)

	return diags
}

package rockset

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/rockset/rockset-go-client"
)

func dataSourceRocksetQueryLambda() *schema.Resource {
	return &schema.Resource{
		Description: fmt.Sprintf("Gets information about a query lambda. The `tag` defaults to `%s`.", rockset.LatestTag),
		ReadContext: dataSourceReadRocksetQueryLambda,

		Schema: map[string]*schema.Schema{
			"workspace": {
				Description: "Workspace the query lambda resides in.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"name": {
				Description: "Name of the query lambda.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"tag": {
				Description: "Tag name.",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     rockset.LatestTag,
			},
			"description": {
				Description: "Description of the query lambda.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"version": {
				Description: "Query lambda tag version.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"last_executed": {
				Description: "Last time the query lambda was executed.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"sql": {
				Description: "Query lambda SQL.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		}}
}

func dataSourceReadRocksetQueryLambda(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rc := meta.(*rockset.RockClient)
	var diags diag.Diagnostics

	ws := d.Get("workspace").(string)
	name := d.Get("name").(string)
	tag := d.Get("tag").(string)

	ql, err := rc.GetQueryLambdaVersionByTag(ctx, ws, name, tag)
	if err != nil {
		return DiagFromErr(err)
	}

	v := ql.GetVersion()
	if err = d.Set("description", v.GetDescription()); err != nil {
		return DiagFromErr(err)
	}
	if err = d.Set("version", v.GetVersion()); err != nil {
		return DiagFromErr(err)
	}
	if err = d.Set("sql", v.Sql.GetQuery()); err != nil {
		return DiagFromErr(err)
	}
	if err = d.Set("last_executed", v.Stats.GetLastExecuted()); err != nil {
		return DiagFromErr(err)
	}

	d.SetId(toID(ws, name))

	return diags
}

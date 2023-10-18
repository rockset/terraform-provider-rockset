package rockset

import (
	"context"
	"github.com/rockset/rockset-go-client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceRocksetVirtualInstance() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceReadRocksetVirtualInstance,

		Schema: map[string]*schema.Schema{
			"id": {
				Description: "Virtual Instance id.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"name": {
				Description: "Virtual Instance name.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"description": {
				Description: "Virtual Instance description.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			// auto_suspend_enabled
			"auto_suspend_seconds": {
				Description: "Number of seconds without queries after which the Virtual Instance is suspended.",
				Type:        schema.TypeInt,
				Optional:    true,
			},
			"current_size": {
				Description: "Virtual Instance current size.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"desired_size": {
				Description: "Virtual Instance desired size.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"default": {
				Description: "Virtual Instance name.",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"state": {
				Description: "Virtual Instance state.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"enable_remount_on_resume": {
				Description: "When a Virtual Instance is resumed, it will remount all collections that were mounted when the Virtual Instance was suspended.",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			// TODO: auto_scaling_policy
		}}
}

func dataSourceReadRocksetVirtualInstance(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rc := meta.(*rockset.RockClient)
	var diags diag.Diagnostics

	id := d.Get("id").(string)

	vi, err := rc.GetVirtualInstance(ctx, id)
	if err != nil {
		return diag.FromErr(err)
	}

	if err = d.Set("name", vi.GetName()); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("description", vi.GetDescription()); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("auto_suspend_seconds", vi.GetAutoSuspendSeconds()); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("current_size", vi.GetCurrentSize()); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("desired_size", vi.GetDesiredSize()); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("default", vi.GetDefaultVi()); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("state", vi.GetState()); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("enable_remount_on_resume", vi.GetEnableRemountOnResume()); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(id)

	return diags
}

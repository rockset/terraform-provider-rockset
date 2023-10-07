package rockset

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/rockset/rockset-go-client"
	"github.com/rockset/rockset-go-client/openapi"
	"github.com/rockset/rockset-go-client/option"
)

func resourceVirtualInstance() *schema.Resource {
	return &schema.Resource{
		Description: `Manages a Rockset Virtual Instance. To be able to create a new Virtual Instance,
The main virtual instance must use a dedicated instance to create a secondary virtual instance, 
which must be SMALL or larger. To enable live mount, the secondary virtual instance must be MEDIUM or larger.`,

		CreateContext: resourceVirtualInstanceCreate,
		ReadContext:   resourceVirtualInstanceRead,
		UpdateContext: resourceVirtualInstanceUpdate,
		DeleteContext: resourceVirtualInstanceDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"id": {
				Description: "Unique ID of this Virtual Instance.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"rrn": {
				Description: "RRN of this Virtual Instance.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"name": {
				Description:  "Name of the virtual instance.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: rocksetNameValidator,
			},
			"description": {
				Description: "Description of the virtual instance.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"size": {
				Description: "Requested virtual instance size. Note that this field is called type in the API documentation.",
				Type:        schema.TypeString,
				Required:    true,
				// TODO should we have a validator for the allowed types?
			},
			"current_size": {
				Description: "Current size of the virtual instance.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"desired_size": {
				Description: "Desired size of the virtual instance.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			// auto_scaling_policy can't be supported until it can be set on creation
			"remount_on_resume": {
				Description: "When a Virtual Instance is resumed, remount all collections that were mounted when the Virtual Instance was suspended.",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"auto_suspend_seconds": {
				Description: "Number of seconds without queries after which the Virtual Instance is suspended.",
				Type:        schema.TypeInt,
				Optional:    true,
			},
			"mount_refresh_interval_seconds": {
				Description: "Number of seconds between data refreshes for mounts on this Virtual Instance. A value of 0 means continuous refresh and a value of null means never refresh.",
				Type:        schema.TypeInt,
				Optional:    true,
			},
			"state": {
				Description: "Virtual Instance state.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"default": {
				Description: "Is this Virtual Instance the default.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"monitoring_enabled": {
				Description: "Is monitoring enabled for this Virtual Instance.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			// TODO scaled_pod_count has no documentation,
		},
	}
}

func resourceVirtualInstanceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rc := meta.(*rockset.RockClient)
	var diags diag.Diagnostics

	name := d.Get("name").(string)
	options := getVirtualInstanceOptions(d)

	vi, err := rc.CreateVirtualInstance(ctx, name, options...)
	if err != nil {
		return diag.FromErr(err)
	}

	id := vi.GetId()
	d.SetId(id)

	// TODO make it possible to skip waiting, and then parse the fields from the created vi
	err = rc.Wait.UntilVirtualInstanceActive(ctx, id)
	if err != nil {
		return diag.FromErr(err)
	}

	// get the vi info, so we have updated value for current_size
	vi, err = rc.GetVirtualInstance(ctx, id)
	if err != nil {
		return diag.FromErr(err)
	}
	if err = parseVirtualInstanceFields(vi, d); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceVirtualInstanceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rc := meta.(*rockset.RockClient)
	var diags diag.Diagnostics

	id := d.Id()

	vi, err := rc.GetVirtualInstance(ctx, id)
	if err != nil {
		return diag.FromErr(err)
	}

	if err = parseVirtualInstanceFields(vi, d); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceVirtualInstanceUpdate(ctx context.Context, d *schema.ResourceData,
	meta interface{}) diag.Diagnostics {
	rc := meta.(*rockset.RockClient)
	var diags diag.Diagnostics

	id := d.Id()
	options := getVirtualInstanceOptions(d)

	vi, err := rc.UpdateVirtualInstance(ctx, id, options...)
	if err != nil {
		return diag.FromErr(err)
	}

	// TODO make it possible to skip waiting
	err = rc.Wait.UntilVirtualInstanceActive(ctx, id)
	if err != nil {
		return diag.FromErr(err)
	}

	// get the vi info, so we have updated value for current_size
	vi, err = rc.GetVirtualInstance(ctx, id)
	if err != nil {
		return diag.FromErr(err)
	}
	if err = parseVirtualInstanceFields(vi, d); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceVirtualInstanceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rc := meta.(*rockset.RockClient)
	var diags diag.Diagnostics

	id := d.Id()

	_, err := rc.DeleteVirtualInstance(ctx, id)
	if err != nil {
		return diag.FromErr(err)
	}

	// TODO wait until deleted
	//err = rc.Wait.UntilVirtualInstanceGone(ctx, id)
	//if err != nil {
	//	return diag.FromErr(err)
	//}

	return diags
}

func getVirtualInstanceOptions(d *schema.ResourceData) []option.VirtualInstanceOption {
	var options []option.VirtualInstanceOption

	addOptionIfChanged(d, "name", &options, func(a any) option.VirtualInstanceOption {
		return option.WithVirtualInstanceName(a.(string))
	})
	addOptionIfChanged(d, "description", &options, func(a any) option.VirtualInstanceOption {
		return option.WithVirtualInstanceDescription(a.(string))
	})
	addOptionIfChanged(d, "size", &options, func(a any) option.VirtualInstanceOption {
		return option.WithVirtualInstanceSize(option.VirtualInstanceSize(a.(string)))
	})
	addOptionIfChanged(d, "auto_suspend_seconds", &options, func(a any) option.VirtualInstanceOption {
		seconds := time.Duration(a.(int)) * time.Second
		return option.WithAutoSuspend(seconds)
	})
	addOptionIfChanged(d, "remount_on_resume", &options, func(a any) option.VirtualInstanceOption {
		return option.WithRemountOnResume(a.(bool))
	})
	addOptionIfChanged(d, "mount_refresh_interval_seconds", &options, func(a any) option.VirtualInstanceOption {
		// option.WithNoMountRefresh()
		seconds := time.Duration(a.(int)) * time.Second
		return option.WithMountRefreshInterval(seconds)
	})

	return options
}

func parseVirtualInstanceFields(vi openapi.VirtualInstance, d *schema.ResourceData) error {
	if err := setValue(d, "name", vi.GetNameOk); err != nil {
		return err
	}
	if err := setValue(d, "description", vi.GetDescriptionOk); err != nil {
		return err
	}
	if err := setValue(d, "rrn", vi.GetRrnOk); err != nil {
		return err
	}
	if err := setValue(d, "current_size", vi.GetCurrentSizeOk); err != nil {
		return err
	}
	if err := setValue(d, "desired_size", vi.GetDesiredSizeOk); err != nil {
		return err
	}
	if err := setValue(d, "remount_on_resume", vi.GetEnableRemountOnResumeOk); err != nil {
		return err
	}
	if err := setValue(d, "auto_suspend_seconds", vi.GetAutoSuspendSecondsOk); err != nil {
		return err
	}
	if err := setValue(d, "mount_refresh_interval_seconds", vi.GetMountRefreshIntervalSecondsOk); err != nil {
		return err
	}
	if err := setValue(d, "state", vi.GetStateOk); err != nil {
		return err
	}
	if err := setValue(d, "default", vi.GetDefaultViOk); err != nil {
		return err
	}
	if err := setValue(d, "monitoring_enabled", vi.GetMonitoringEnabledOk); err != nil {
		return err
	}

	return nil
}

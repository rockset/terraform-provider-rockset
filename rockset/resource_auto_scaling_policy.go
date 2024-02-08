package rockset

import (
	"context"
	"strings"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/rockset/rockset-go-client"
	"github.com/rockset/rockset-go-client/openapi"
	"github.com/rockset/rockset-go-client/option"
)

func resourceAutoScalingPolicy() *schema.Resource {
	return &schema.Resource{
		Description: "Manages the auto-scaling policy for the **main** Virtual Instance.\n\n" +
			"The instance must be a dedicated (size `XSMALL` or larger) instance to set the policy.",

		CreateContext: resourceAutoScalingPolicyCreate,
		ReadContext:   resourceAutoScalingPolicyRead,
		UpdateContext: resourceAutoScalingPolicyUpdate,
		DeleteContext: resourceAutoScalingPolicyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"virtual_instance_id": {
				Description: "Unique ID of this Virtual Instance.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"enabled": {
				Description: "Is auto-scaling enabled for this Virtual Instance.",
				Type:        schema.TypeBool,
				Required:    true,
			},
			"min_size": {
				Description: "Minimum size of the Virtual Instance.",
				Type:        schema.TypeString,
				Required:    true,
				ValidateDiagFunc: func(i interface{}, path cty.Path) diag.Diagnostics {
					return validateVirtualInstanceSize("min_size", i.(string))
				},
			},
			"max_size": {
				Description: "Maximum size of the Virtual Instance.",
				Type:        schema.TypeString,
				Required:    true,
				ValidateDiagFunc: func(i interface{}, path cty.Path) diag.Diagnostics {
					return validateVirtualInstanceSize("max_size", i.(string))
				},
			},
		},
	}
}

// TODO move into the go client instead
var virtualInstanceSizes = []option.VirtualInstanceSize{
	option.SizeXSmall,
	option.SizeSmall,
	option.SizeMedium,
	option.SizeLarge,
	option.SizeXLarge,
	option.SizeXLarge2,
	option.SizeXLarge4,
	option.SizeXLarge8,
	option.SizeXLarge16,
}

func validateVirtualInstanceSize(field, size string) diag.Diagnostics {
	for _, s := range virtualInstanceSizes {
		if s.String() == size {
			return nil
		}
	}
	sizes := make([]string, len(virtualInstanceSizes))
	for i, s := range virtualInstanceSizes {
		sizes[i] = s.String()
	}

	return diag.Errorf("size must be one of: %s", strings.Join(sizes, ", "))
}

func resourceAutoScalingPolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rc := meta.(*rockset.RockClient)
	id := d.Get("virtual_instance_id").(string)
	d.SetId(id)

	tflog.Trace(ctx, "creating auto-scaling policy")
	return updateAutoScalingPolicy(ctx, rc, id, d)
}

func resourceAutoScalingPolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rc := meta.(*rockset.RockClient)
	id := d.Get("virtual_instance_id").(string)

	tflog.Trace(ctx, "updating auto-scaling policy")
	return updateAutoScalingPolicy(ctx, rc, id, d)
}

func updateAutoScalingPolicy(ctx context.Context, rc *rockset.RockClient, id string, d *schema.ResourceData) diag.Diagnostics {
	var policy openapi.AutoScalingPolicy
	enabled := d.Get("enabled").(bool)
	minSize := d.Get("min_size").(string)
	maxSize := d.Get("max_size").(string)

	policy.Enabled = &enabled
	policy.MinSize = &minSize
	policy.MaxSize = &maxSize

	tflog.Trace(ctx, "updating auto-scaling policy", map[string]interface{}{
		"virtual_instance_id": id,
		"enabled":             enabled,
		"min_size":            minSize,
		"max_size":            maxSize,
	})
	vi, err := rc.UpdateVirtualInstance(ctx, id, option.WithVirtualInstanceAutoScalingPolicy(policy))
	if err != nil {
		return DiagFromErr(err)
	}

	policy = vi.GetAutoScalingPolicy()
	tflog.Info(ctx, "updated auto-scaling policy", map[string]interface{}{
		"virtual_instance_id": id,
		"enabled":             policy.GetEnabled(),
		"min_size":            policy.GetMinSize(),
		"max_size":            policy.GetMaxSize(),
	})

	return nil
}

func resourceAutoScalingPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rc := meta.(*rockset.RockClient)
	var diags diag.Diagnostics

	id := d.Id()

	vi, err := rc.GetVirtualInstance(ctx, id)
	if err != nil {
		return DiagFromErr(err)
	}

	policy := vi.GetAutoScalingPolicy()
	if err = d.Set("enabled", policy.GetEnabled()); err != nil {
		return DiagFromErr(err)
	}
	if err = d.Set("min_size", policy.GetMinSize()); err != nil {
		return DiagFromErr(err)
	}
	if err = d.Set("max_size", policy.GetMaxSize()); err != nil {
		return DiagFromErr(err)
	}

	tflog.Info(ctx, "read auto-scaling policy", map[string]interface{}{
		"virtual_instance_id": id,
		"enabled":             policy.GetEnabled(),
		"min_size":            policy.GetMinSize(),
		"max_size":            policy.GetMaxSize(),
	})

	return diags
}

func resourceAutoScalingPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rc := meta.(*rockset.RockClient)
	var diags diag.Diagnostics

	id := d.Id()

	vi, err := rc.GetVirtualInstance(ctx, id)
	if err != nil {
		return DiagFromErr(err)
	}
	policy := vi.GetAutoScalingPolicy()
	policy.Enabled = openapi.PtrBool(false)

	_, err = rc.UpdateVirtualInstance(ctx, id, option.WithVirtualInstanceAutoScalingPolicy(policy))
	if err != nil {
		return DiagFromErr(err)
	}

	d.SetId("")
	tflog.Info(ctx, "deleted auto-scaling policy", map[string]interface{}{"virtual_instance_id": id})

	return diags
}

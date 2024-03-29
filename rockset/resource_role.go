package rockset

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/rockset/rockset-go-client"
	"github.com/rockset/rockset-go-client/openapi"
	"github.com/rockset/rockset-go-client/option"
)

func resourceRole() *schema.Resource {
	return &schema.Resource{
		Description: "Manages a Rockset [Role](https://rockset.com/docs/iam/).",

		CreateContext: resourceRoleCreate,
		ReadContext:   resourceRoleRead,
		UpdateContext: resourceRoleUpdate,
		DeleteContext: resourceRoleDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Role name.",
				Type:        schema.TypeString,
				ForceNew:    true,
				Required:    true,
			},
			"description": {
				Description: "Role description.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"owner_email": {
				Description: "The email of the user who currently owns the role.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"created_by": {
				Description: "Who created the role.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"created_at": {
				Description: "When the role was created.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"privilege": {
				Description: "Privileges associated with the role. " +
					"For a full list see [API documentation](https://rockset.com/docs/iam/#supported-privileges)",
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"action": {
							Description: "The action allowed by this privilege.",
							Required:    true,
							Type:        schema.TypeString,
							ValidateDiagFunc: func(i interface{}, path cty.Path) diag.Diagnostics {
								action := i.(string)
								if option.IsGlobalAction(action) || option.IsIntegrationAction(action) ||
									option.IsWorkspaceAction(action) || option.IsVirtualInstanceAction(action) {
									return diag.Diagnostics{}
								}
								return diag.Errorf("%s is not a valid action", action) // TODO list all actions?
							},
						},
						"resource_name": {
							Description: "The resource on which this action is allowed. Defaults to 'All' if not specified.",
							Optional:    true,
							Type:        schema.TypeString,
						},
						"cluster": {
							Type:     schema.TypeString,
							Optional: true,
							Description: "Rockset cluster ID for which this action is allowed. " +
								"Only valid for Workspace actions. Use '*ALL*' for actions which apply to all clusters.",
						},
					},
				},
			},
		},
	}
}

func resourceRoleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rc := meta.(*rockset.RockClient)
	var diags diag.Diagnostics
	var options []option.RoleOption

	name := d.Get("name").(string)
	desc := d.Get("description").(string)
	if desc != "" {
		options = append(options, option.WithRoleDescription(desc))
	}

	in := d.Get("privilege")
	if in != nil {
		privs, err := expandRolePrivileges(in)
		if err != nil {
			return DiagFromErr(err)
		}

		for _, p := range privs {
			tflog.Info(ctx, "privs", map[string]interface{}{
				"action":   p.GetAction(),
				"resource": p.GetResourceName(),
				"cluster":  p.GetCluster(),
			})
		}

		opts, err := rolePrivsToOptions(privs)
		if err != nil {
			return DiagFromErr(err)
		}
		options = append(options, opts...)
	}

	_, err := rc.CreateRole(ctx, name, options...)
	if err != nil {
		return DiagFromErr(err)
	}

	d.SetId(name)

	return diags
}

func resourceRoleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rc := meta.(*rockset.RockClient)
	var diags diag.Diagnostics

	var options []option.RoleOption

	desc := d.Get("description").(string)
	if desc != "" {
		options = append(options, option.WithRoleDescription(desc))
	}

	in := d.Get("privilege")
	if in != nil {
		privs, err := expandRolePrivileges(in)
		if err != nil {
			return DiagFromErr(err)
		}

		opts, err := rolePrivsToOptions(privs)
		if err != nil {
			return DiagFromErr(err)
		}
		options = append(options, opts...)
	}

	_, err := rc.UpdateRole(ctx, d.Id(), options...)
	if err != nil {
		return DiagFromErr(err)
	}

	return diags
}

func resourceRoleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rc := meta.(*rockset.RockClient)
	var diags diag.Diagnostics

	name := d.Id()
	role, err := rc.GetRole(ctx, name)
	if err != nil {
		return checkForNotFoundError(d, err)
	}

	err = d.Set("name", role.GetRoleName())
	if err != nil {
		return DiagFromErr(err)
	}

	err = d.Set("owner_email", role.GetOwnerEmail())
	if err != nil {
		return DiagFromErr(err)
	}

	err = d.Set("created_by", role.GetCreatedBy())
	if err != nil {
		return DiagFromErr(err)
	}

	err = d.Set("created_at", role.GetCreatedAt())
	if err != nil {
		return DiagFromErr(err)
	}

	err = d.Set("privilege", flattenRolePrivileges(role.Privileges))
	if err != nil {
		return DiagFromErr(err)
	}

	return diags
}

func resourceRoleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rc := meta.(*rockset.RockClient)
	var diags diag.Diagnostics

	name := d.Id()
	err := rc.DeleteRole(ctx, name)

	if err != nil {
		return DiagFromErr(err)
	}

	return diags
}

// Flatten:
// API JSON > Unmarshal > native Go struct > Flatten > Internal Representation > State JSON

func flattenRolePrivileges(privs []openapi.Privilege) []interface{} {
	out := make([]interface{}, len(privs))

	for i, p := range privs {
		m := make(map[string]interface{})

		m["action"] = p.GetAction()

		if rn, ok := p.GetResourceNameOk(); ok {
			m["resource_name"] = rn
		}

		if c, ok := p.GetClusterOk(); ok {
			m["cluster"] = c
		}

		// if
		if m["cluster"] == "" && option.GetWorkspaceAction(p.GetAction()) != option.UnknownWorkspaceAction {
			m["cluster"] = option.AllClusters
		}

		out[i] = m
	}

	return out
}

// Expand:
// HCL > Parsed > Internal Representation > Expand > native Go struct > Marshal > API JSON

func expandRolePrivileges(in interface{}) ([]openapi.Privilege, error) {
	privs := make([]openapi.Privilege, 0, in.(*schema.Set).Len())

	for _, i := range in.(*schema.Set).List() {
		if val, ok := i.(map[string]interface{}); ok {
			var err error
			var priv openapi.Privilege
			priv.Action, err = lookupString(val, "action", true)
			if err != nil {
				return nil, err
			}

			priv.ResourceName, err = lookupString(val, "resource_name", false)
			if err != nil {
				return nil, err
			}

			if priv.GetResourceName() != "" && option.IsGlobalAction(priv.GetAction()) {
				return nil, fmt.Errorf("can't specify resource_name for %s action", priv.GetAction())
			}

			priv.Cluster, err = lookupString(val, "cluster", false)
			if err != nil {
				return nil, err
			}
			if priv.GetCluster() != "" &&
				(option.IsGlobalAction(priv.GetAction()) || option.IsIntegrationAction(priv.GetAction())) {
				return nil, fmt.Errorf("can't specify cluster for %s action", priv.GetAction())
			}

			if option.IsWorkspaceAction(priv.GetAction()) && priv.Cluster == nil {
				all := option.AllClusters
				priv.Cluster = &all
			}

			privs = append(privs, priv)
		}
	}

	return privs, nil
}

func lookupString(val map[string]interface{}, key string, required bool) (*string, error) {
	a, ok := val[key]
	if !ok {
		if required {
			return nil, nil
		}
		return nil, fmt.Errorf("can't define privilege block without %s", key)
	}
	str, ok := a.(string)
	if !ok {
		return nil, fmt.Errorf("failed to cast %s %v to string", key, a)
	}

	return &str, nil
}

func rolePrivsToOptions(privs []openapi.Privilege) ([]option.RoleOption, error) {
	var opts []option.RoleOption

	for _, p := range privs {
		if a := option.GetGlobalAction(p.GetAction()); a != option.UnknownGlobalAction {
			opts = append(opts, option.WithGlobalPrivilege(a))
			continue
		}

		if a := option.GetIntegrationAction(p.GetAction()); a != option.UnknownIntegrationAction {
			opts = append(opts, option.WithIntegrationPrivilege(a, p.GetResourceName()))
			continue
		}

		if a := option.GetVirtualInstanceAction(p.GetAction()); a != option.UnknownVirtualInstanceAction {
			var c []option.ClusterPrivileges
			if p.GetCluster() != "" {
				c = append(c, option.WithCluster(p.GetCluster()))
			}
			opts = append(opts, option.WithVirtualInstancePrivilege(a, p.GetResourceName(), c...))
			continue
		}

		if a := option.GetWorkspaceAction(p.GetAction()); a != option.UnknownWorkspaceAction {
			var c []option.ClusterPrivileges
			if p.GetCluster() != "" {
				c = append(c, option.WithCluster(p.GetCluster()))
			}
			opts = append(opts, option.WithWorkspacePrivilege(a, p.GetResourceName(), c...))
			continue
		}

		return nil, fmt.Errorf("unknown privilege action %s", p.GetAction())
	}
	return opts, nil
}

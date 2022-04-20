package rockset

import (
	"context"
	"fmt"

	"github.com/rockset/rockset-go-client/option"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/rockset/rockset-go-client"
	"github.com/rockset/rockset-go-client/openapi"
)

func resourceQueryLambda() *schema.Resource {
	return &schema.Resource{
		Description: "Manages a Rockset Query Lambda.",

		CreateContext: resourceQueryLambdaCreate,
		ReadContext:   resourceQueryLambdaRead,
		UpdateContext: resourceQueryLambdaUpdate,
		DeleteContext: resourceQueryLambdaDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"workspace": {
				Description: "The name of the workspace.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"name": {
				Description:  "Unique identifier for the query lambda. Can contain alphanumeric or dash characters.",
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: rocksetNameValidator,
			},
			"description": {
				Description: "Text describing the query lambda.",
				Type:        schema.TypeString,
				Default:     "created by Rockset terraform provider",
				Optional:    true,
			},
			"version": {
				Description: "The latest version string of this query lambda.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"state": {
				Description: "The latest state of this query lambda.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"sql": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"query": {
							Type:     schema.TypeString,
							Required: true,
						},
						"default_parameter": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:     schema.TypeString,
										Required: true,
									},
									"type": {
										Type:     schema.TypeString,
										Required: true,
									},
									"value": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func resourceQueryLambdaCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rc := meta.(*rockset.RockClient)
	var diags diag.Diagnostics

	workspace := d.Get("workspace").(string)
	name := d.Get("name").(string)
	sql := makeQueryLambdaSQL(d.Get("sql"))
	options := makeQueryLambdaOptions(d.Get("description").(string), sql.DefaultParameters)

	ql, err := rc.CreateQueryLambda(ctx, workspace, name, sql.Query, options...)
	if err != nil {
		return diag.FromErr(err)
	}

	if ql.Version != nil {
		err = d.Set("version", ql.Version)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if ql.State != nil {
		err = d.Set("state", ql.State)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	d.SetId(toID(workspace, name))

	return diags
}

func makeQueryLambdaOptions(desc string, params []openapi.QueryParameter) []option.CreateQueryLambdaOption {
	var options []option.CreateQueryLambdaOption

	if desc != "" {
		options = append(options, option.WithQueryLambdaDescription(desc))
	}

	for _, p := range params {
		options = append(options, option.WithDefaultParameter(p.Name, p.Type, p.Value))
	}

	return options
}

func resourceQueryLambdaRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rc := meta.(*rockset.RockClient)
	var diags diag.Diagnostics

	workspace, name := workspaceAndNameFromID(d.Id())
	ql, err := getQueryLambda(ctx, rc, workspace, name)
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("workspace", ql.Workspace)
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("name", ql.Name)
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("description", ql.LatestVersion.Description)
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("version", ql.LatestVersion.Version)
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("state", ql.LatestVersion.State)
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("sql", flattenQueryLambdaSQL(ql.LatestVersion.Sql))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(toID(workspace, name))

	return diags
}

func resourceQueryLambdaUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rc := meta.(*rockset.RockClient)
	var diags diag.Diagnostics

	workspace := d.Get("workspace").(string)
	name := d.Get("name").(string)
	sql := makeQueryLambdaSQL(d.Get("sql"))
	options := makeQueryLambdaOptions(d.Get("description").(string), sql.DefaultParameters)

	ql, err := rc.UpdateQueryLambda(ctx, workspace, name, sql.Query, options...)
	if err != nil {
		return diag.FromErr(err)
	}

	if ql.Version != nil {
		err = d.Set("version", ql.Version)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if ql.State != nil {
		err = d.Set("state", ql.State)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	d.SetId(toID(workspace, name))

	return diags
}

func resourceQueryLambdaDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rc := meta.(*rockset.RockClient)
	var diags diag.Diagnostics

	workspace, name := workspaceAndNameFromID(d.Id())
	err := rc.DeleteQueryLambda(ctx, workspace, name)
	if err != nil {
		return diag.Errorf("failed to delete query lambda %s.%s: %v", workspace, name, err)
	}

	return diags
}

func makeQueryLambdaSQL(in interface{}) openapi.QueryLambdaSql {
	sql := openapi.QueryLambdaSql{}
	var empty []openapi.QueryParameter
	sql.DefaultParameters = empty

	if set, ok := in.(*schema.Set); ok {
		for _, s := range set.List() {
			if i, ok := s.(map[string]interface{}); ok {
				for k, v := range i {
					switch k {
					case "query":
						sql.Query = v.(string)
					case "default_parameter":
						sql.DefaultParameters = makeDefaultParameters(v)
					}
				}
			}
		}
	}

	return sql
}

func getQueryLambda(ctx context.Context, rc *rockset.RockClient, workspace string,
	name string) (*openapi.QueryLambda, error) {
	lambdas, err := rc.ListQueryLambdas(ctx, option.WithQueryLambdaWorkspace(workspace))
	if err != nil {
		return nil, err
	}

	for _, ql := range lambdas {
		if *ql.Name == name {
			return &ql, nil
		}
	}

	return nil, fmt.Errorf("query lambda %s not found in workspace %s", name, workspace)
}

func makeDefaultParameters(input interface{}) []openapi.QueryParameter {
	dps := make([]openapi.QueryParameter, 0, input.(*schema.Set).Len())

	for _, i := range input.(*schema.Set).List() {
		if cfg, ok := i.(map[string]interface{}); ok {
			dp := openapi.QueryParameter{}
			for k, v := range cfg {
				switch k {
				case "name":
					dp.Name = v.(string)
				case "type":
					dp.Type = v.(string)
				case "value":
					dp.Value = v.(string)
				}
			}
			dps = append(dps, dp)
		}
	}

	return dps
}

func flattenQueryLambdaSQL(sql *openapi.QueryLambdaSql) []interface{} {
	var m = make(map[string]interface{})
	m["query"] = sql.Query

	var r []interface{}
	for _, qp := range sql.DefaultParameters {
		m := make(map[string]interface{})
		m["name"] = qp.Name
		m["type"] = qp.Type
		m["value"] = qp.Value
		r = append(r, m)
	}
	m["default_parameter"] = r

	return []interface{}{m}
}

package rockset

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/rockset/rockset-go-client"
	"github.com/rockset/rockset-go-client/option"
)

func resourceView() *schema.Resource {
	return &schema.Resource{
		Description: "Manages a Rockset view.",

		CreateContext: resourceViewCreate,
		ReadContext:   resourceViewRead,
		DeleteContext: resourceViewDelete,
		UpdateContext: resourceViewUpdate,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"workspace": {
				Description:  "Workspace name.",
				Type:         schema.TypeString,
				ForceNew:     true,
				Required:     true,
				ValidateFunc: rocksetNameValidator,
			},
			"name": {
				Description:  "Unique name for the view in the workspace. Can contain alphanumeric or dash characters.",
				Type:         schema.TypeString,
				ForceNew:     true,
				Required:     true,
				ValidateFunc: rocksetNameValidator,
			},
			"query": {
				Description: "SQL query used for thw view.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"description": {
				Description: "Text describing the collection.",
				Type:        schema.TypeString,
				Default:     "created by Rockset terraform provider",
				ForceNew:    true,
				Optional:    true,
			},
			"created_by": {
				Description: "The user who created the view.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func resourceViewCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rc := meta.(*rockset.RockClient)
	var diags diag.Diagnostics

	workspace := d.Get("workspace").(string)
	name := d.Get("name").(string)
	description := d.Get("description").(string)
	query := d.Get("query").(string)

	view, err := rc.CreateView(ctx, workspace, name, query, option.WithViewDescription(description))
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("created_by", view.GetCreatorEmail())
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(toID(workspace, name))

	return diags
}

func resourceViewUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rc := meta.(*rockset.RockClient)
	var diags diag.Diagnostics

	workspace, name := workspaceAndNameFromID(d.Id())

	desc := d.Get("description").(string)
	query := d.Get("query").(string)
	var opts []option.ViewOption
	if desc != "" {
		opts = append(opts, option.WithViewDescription(desc))
	}

	_, err := rc.UpdateView(ctx, workspace, name, query, opts...)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceViewRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rc := meta.(*rockset.RockClient)
	var diags diag.Diagnostics

	workspace, name := workspaceAndNameFromID(d.Id())

	view, err := rc.GetView(ctx, workspace, name)
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("name", view.GetName())
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("query", view.GetQuerySql())
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("description", view.GetDescription())
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("created_by", view.GetCreatorEmail())
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceViewDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rc := meta.(*rockset.RockClient)
	var diags diag.Diagnostics

	workspace, name := workspaceAndNameFromID(d.Id())

	err := rc.DeleteView(ctx, workspace, name)
	if err != nil {
		return diag.FromErr(err)
	}

	err = rc.WaitUntilViewGone(ctx, workspace, name)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

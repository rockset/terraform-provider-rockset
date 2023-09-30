package rockset

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/rockset/rockset-go-client"
	"github.com/rockset/rockset-go-client/option"
)

func resourceWorkspace() *schema.Resource {
	return &schema.Resource{
		Description: "Manages a Rockset workspace, which can hold collections, query lambdas and views.",

		CreateContext: resourceWorkspaceCreate,
		ReadContext:   resourceWorkspaceRead,
		DeleteContext: resourceWorkspaceDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The workspace ID, in the form of the workspace `name`.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"name": {
				Description:  "Unique identifier for workspace. Can contain alphanumeric or dash characters.",
				Type:         schema.TypeString,
				ForceNew:     true,
				Required:     true,
				ValidateFunc: rocksetNameValidator,
			},
			"description": {
				Description: "Text describing the collection.",
				Type:        schema.TypeString,
				Default:     "created by Rockset terraform provider",
				ForceNew:    true,
				Optional:    true,
			},
			"created_by": {
				Description: "The user who created the workspace.",
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
		},
	}
}

func resourceWorkspaceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rc := meta.(*rockset.RockClient)
	var diags diag.Diagnostics

	name := d.Get("name").(string)
	description := d.Get("description").(string)

	workspace, err := rc.CreateWorkspace(ctx, name, option.WithWorkspaceDescription(description))
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("created_at", workspace.GetCreatedAt())
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("created_by", workspace.GetCreatedBy())
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("collection_count", workspace.GetCollectionCount())
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(name)

	return diags
}

func resourceWorkspaceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rc := meta.(*rockset.RockClient)
	var diags diag.Diagnostics

	name := d.Id()

	workspace, err := rc.GetWorkspace(ctx, name)
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("name", workspace.GetName())
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("description", workspace.GetDescription())
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("created_by", workspace.GetCreatedBy())
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("created_at", workspace.GetCreatedAt())
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("collection_count", workspace.GetCollectionCount())
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceWorkspaceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rc := meta.(*rockset.RockClient)
	var diags diag.Diagnostics

	name := d.Id()

	err := rc.DeleteWorkspace(ctx, name)
	if err != nil {
		return diag.FromErr(err)
	}

	err = rc.WaitUntilWorkspaceGone(ctx, name)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

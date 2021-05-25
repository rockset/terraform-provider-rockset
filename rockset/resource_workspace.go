package rockset

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/rockset/rockset-go-client"
	"github.com/rockset/rockset-go-client/option"
)

func resourceWorkspace() *schema.Resource {
	return &schema.Resource{
		Description: "Sample resource in the Terraform provider Workspace.",

		CreateContext: resourceWorkspaceCreate,
		ReadContext:   resourceWorkspaceRead,
		DeleteContext: resourceWorkspaceDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				ForceNew:     true,
				Required:     true,
				ValidateFunc: rocksetNameValidator,
			},
			"description": {
				Type:     schema.TypeString,
				Default:  "created by Rockset terraform provider",
				ForceNew: true,
				Optional: true,
			},
			"created_by": {
				Type:     schema.TypeString,
				Computed: true,
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

	fmt.Printf("Got description: %s", workspace.GetDescription())

	err = d.Set("created_by", workspace.GetCreatedBy())
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

	return diags
}

package rockset

import (
	"context"
	"reflect"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/rockset/rockset-go-client"
	"github.com/rockset/rockset-go-client/openapi"
)

func resourceAlias() *schema.Resource {
	return &schema.Resource{
		Description: "Manages an alias for a set of collections.",

		CreateContext: resourceAliasCreate,
		ReadContext:   resourceAliasRead,
		UpdateContext: resourceAliasUpdate,
		DeleteContext: resourceAliasDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:         schema.TypeString,
				ForceNew:     true,
				Required:     true,
				ValidateFunc: rocksetNameValidator,
			},
			"workspace": &schema.Schema{
				Type:         schema.TypeString,
				ForceNew:     true,
				Required:     true,
				ValidateFunc: rocksetNameValidator,
			},
			"description": &schema.Schema{
				Type:     schema.TypeString,
				Default:  "created by Rockset terraform provider",
				ForceNew: false,
				Optional: true,
			},
			"collections": {
				/*
					NOTE: This is a list for forward compatibility
					but it will fail for now if the list isn't exactly 1 item
					Check in and update this when aliases that point to multiple
					collections becomes a feature.
				*/
				Description: "List of collections for this alis to refer to.",
				Type:        schema.TypeList,
				MinItems:    1,
				MaxItems:    1,
				ForceNew:    false,
				Required:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func aliasCollectionsSet(ctx context.Context, rc *rockset.RockClient, workspace string, name string, collections []string) rockset.RetryFunc {
	/*
		Implements a Retry func to wait for the create or update
		to finalize and show the specified collections.
		If we don't do this two applies in a row will show pending changes.
	*/
	return func() (bool, error) {
		q := rc.AliasesApi.GetAlias(ctx, workspace, name)

		resp, _, err := q.Execute()
		if err != nil {
			return false, err
		}

		collectionsCorrect := reflect.DeepEqual(resp.Data.GetCollections(), collections)

		// If true, return false so we stop looping
		return !collectionsCorrect, nil
	}
}

func resourceAliasCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rc := meta.(*rockset.RockClient)
	var diags diag.Diagnostics

	workspace := d.Get("workspace").(string)
	name := d.Get("name").(string)

	collections := toStringArray(d.Get("collections").([]interface{}))

	q := rc.AliasesApi.CreateAlias(ctx, workspace)
	req := openapi.NewCreateAliasRequest(name, collections)
	req.SetDescription(d.Get("description").(string))

	resp, _, err := q.Body(*req).Execute()
	if err != nil {
		return diag.FromErr(err)
	}

	// There's a lag between create and update and the alias
	// showing those collections in the response.
	err = rc.Retry(ctx, aliasCollectionsSet(ctx, rc, workspace, name, collections))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(toID(resp.Data.GetWorkspace(), resp.Data.GetName()))

	return diags
}

func resourceAliasRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rc := meta.(*rockset.RockClient)
	var diags diag.Diagnostics

	workspace, name := workspaceAndNameFromID(d.Id())

	q := rc.AliasesApi.GetAlias(ctx, workspace, name)

	resp, _, err := q.Execute()
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("name", resp.Data.GetName())
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("workspace", resp.Data.GetWorkspace())
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("description", resp.Data.GetDescription())
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("collections", resp.Data.GetCollections())
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceAliasUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rc := meta.(*rockset.RockClient)
	var diags diag.Diagnostics

	workspace, name := workspaceAndNameFromID(d.Id())

	collections := toStringArray(d.Get("collections").([]interface{}))

	q := rc.AliasesApi.UpdateAlias(ctx, workspace, name)
	req := openapi.NewUpdateAliasRequest(collections)
	req.SetDescription(d.Get("description").(string))

	_, _, err := q.Body(*req).Execute()
	if err != nil {
		return diag.FromErr(err)
	}

	// There's a lag between create and update and the alias
	// showing those collections in the response.
	err = rc.Retry(ctx, aliasCollectionsSet(ctx, rc, workspace, name, collections))
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceAliasDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rc := meta.(*rockset.RockClient)
	var diags diag.Diagnostics

	workspace, name := workspaceAndNameFromID(d.Id())

	q := rc.AliasesApi.DeleteAlias(ctx, workspace, name)

	_, _, err := q.Execute()
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

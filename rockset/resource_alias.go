package rockset

import (
	"context"
	"errors"
	"reflect"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/rockset/rockset-go-client"
	"github.com/rockset/rockset-go-client/option"
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
				Description:  "Unique identifier for the alias. Can contain alphanumeric or dash characters.",
				Type:         schema.TypeString,
				ForceNew:     true,
				Required:     true,
				ValidateFunc: rocksetNameValidator,
			},
			"workspace": &schema.Schema{
				Description:  "Name of the workspace the alias will be in.",
				Type:         schema.TypeString,
				ForceNew:     true,
				Required:     true,
				ValidateFunc: rocksetNameValidator,
			},
			"description": &schema.Schema{
				Description: "Text describing the alias.",
				Type:        schema.TypeString,
				Default:     "created by Rockset terraform provider",
				ForceNew:    false,
				Optional:    true,
			},
			"collections": {
				/*
					NOTE: This is a list for forward compatibility
					but it will fail for now if the list isn't exactly 1 item
					Check in and update this when aliases that point to multiple
					collections becomes a feature.
				*/
				Description: "List of collections for this alias to refer to.",
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

func aliasCollectionsSet(ctx context.Context, rc *rockset.RockClient, workspace string, name string, collections []string) rockset.RetryCheck {
	/*
		Implements a Retry func to wait for the create or update
		to finalize and show the specified collections.
		If we don't do this two applies in a row will show pending changes.
	*/
	return func() (bool, error) {
		alias, err := rc.GetAlias(ctx, workspace, name)
		if err != nil {
			return false, err
		}

		collectionsCorrect := reflect.DeepEqual(alias.GetCollections(), collections)

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

	_, err := rc.CreateAlias(ctx, workspace, name, collections, option.WithAliasDescription(d.Get("description").(string)))
	if err != nil {
		return diag.FromErr(err)
	}

	// There's a lag between create and update and the alias
	// showing those collections in the response.
	err = rc.RetryWithCheck(ctx, aliasCollectionsSet(ctx, rc, workspace, name, collections))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(toID(workspace, name))

	return diags
}

// checkForNotFoundError check is the error is a Rockset NotFoundError, and then clears the id which makes
// terraform create the resource, but if it isn't a NotFoundError it will return the error wrapped in diag.Diagnostics
func checkForNotFoundError(d *schema.ResourceData, err error) diag.Diagnostics {
	var re rockset.Error
	if !errors.As(err, &re) {
		return diag.FromErr(err)
	}

	if !re.IsNotFoundError() {
		return diag.FromErr(err)
	}

	if err = d.Set("id", ""); err != nil {
		return diag.FromErr(err)
	}

	return diag.Diagnostics{}
}

func resourceAliasRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rc := meta.(*rockset.RockClient)
	var diags diag.Diagnostics

	workspace, name := workspaceAndNameFromID(d.Id())

	alias, err := rc.GetAlias(ctx, workspace, name)
	if err != nil {
		return checkForNotFoundError(d, err)
	}

	err = d.Set("name", alias.GetName())
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("workspace", alias.GetWorkspace())
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("description", alias.GetDescription())
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("collections", alias.GetCollections())
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

	err := rc.UpdateAlias(ctx, workspace, name, collections, option.WithAliasDescription(d.Get("description").(string)))
	if err != nil {
		return diag.FromErr(err)
	}

	// There's a lag between create and update and the alias
	// showing those collections in the response.
	err = rc.RetryWithCheck(ctx, aliasCollectionsSet(ctx, rc, workspace, name, collections))
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceAliasDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rc := meta.(*rockset.RockClient)
	var diags diag.Diagnostics

	workspace, name := workspaceAndNameFromID(d.Id())

	err := rc.DeleteAlias(ctx, workspace, name)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

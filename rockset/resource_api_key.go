package rockset

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/rockset/rockset-go-client"
	"github.com/rockset/rockset-go-client/option"
)

func resourceApiKey() *schema.Resource {
	return &schema.Resource{
		Description: "Manage a Rockset Api Key.\n\n" +
			"Can be used together with roles to scope the actions the api key can take.",

		CreateContext: resourceApiKeyCreate,
		ReadContext:   resourceApiKeyRead,
		DeleteContext: resourceApiKeyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Name of the api key.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"role": {
				Description: `The role the api key will use. If not specified, "All User Assigned Roles" will be used.`,
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			"user": {
				Description: "The user the key is created for.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"key": {
				Description: "The resulting Rockset api key.",
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
			},
		},
	}
}

func nameAndUserToId(name string, user string) string {
	/*
		User may be empty if the key wasn't created for another user.
		ID format will be name or name:user if a user is specified.
	*/
	if user == "" {
		return name
	}

	return fmt.Sprintf("%s:%s", name, user)
}

func idToUserAndName(id string) (string, string, error) {
	tokens := strings.Split(id, ":")

	switch len(tokens) {
	case 1:
		return tokens[0], "", nil // Name, No user
	case 2:
		return tokens[0], tokens[1], nil // Name, User
	default:
		return "", "", fmt.Errorf("id %s is not of the expected format", id) // Bad ID format
	}
}

func resourceApiKeyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rc := meta.(*rockset.RockClient)
	var diags diag.Diagnostics

	name := d.Get("name").(string)
	user := d.Get("user").(string)
	role := d.Get("role").(string)
	id := nameAndUserToId(name, user)

	var opts []option.APIKeyRoleOption
	if role != "" {
		opts = append(opts, option.WithRole(role))
	}

	var err error
	// Use CreateApiKey to create as current authenticated user
	key, err := rc.CreateAPIKey(ctx, name, opts...)

	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("key", key.GetKey())
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(id)

	return diags
}

func resourceApiKeyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rc := meta.(*rockset.RockClient)
	var diags diag.Diagnostics

	name, user, err := idToUserAndName(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	var options []option.APIKeyOption
	if user != "" {
		options = append(options, option.ForUser(user))
	}
	foundApiKey, err := rc.GetAPIKey(ctx, name, options...)
	if err != nil {
		return checkForNotFoundError(d, err)
	}

	if foundApiKey.GetName() == "" { // Failed to find
		return diag.FromErr(fmt.Errorf("API key not found in list"))
	}

	err = d.Set("name", foundApiKey.GetName())
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("role", foundApiKey.GetRole())
	if err != nil {
		return diag.FromErr(err)
	}

	// We intentionally omit here the actual key, as it comes obfuscated from the API.

	// The user field is exposed in no way by the api
	// We can only use the field set from the id
	err = d.Set("user", user)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceApiKeyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rc := meta.(*rockset.RockClient)
	var diags diag.Diagnostics

	name, user, err := idToUserAndName(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	var options []option.APIKeyOption
	if user != "" {
		// Delete as the specified user
		options = append(options, option.ForUser(user))
	}
	err = rc.DeleteAPIKey(ctx, name, options...)

	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

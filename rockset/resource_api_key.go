package rockset

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/rockset/rockset-go-client"
	"github.com/rockset/rockset-go-client/openapi"
)

func resourceApiKey() *schema.Resource {
	return &schema.Resource{
		Description: "Manage an ApiKey.",

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
			"user": {
				Description: "User to create the key for. If not set, defaults to authenticated user.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			"key": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
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
		return "", "", fmt.Errorf("Id %s is not of the expected format.", id) // Bad ID format
	}
}

func resourceApiKeyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rc := meta.(*rockset.RockClient)
	var diags diag.Diagnostics

	name := d.Get("name").(string)
	user := d.Get("user").(string)
	id := nameAndUserToId(name, user)
	req := openapi.NewCreateApiKeyRequest(name)

	var resp openapi.CreateApiKeyResponse
	var err error
	if user != "" {
		// Then we need to use CreateApiKeyAdmin create as the specified user
		resp, _, err = rc.APIKeysApi.CreateApiKeyAdmin(ctx, d.Get("user").(string)).Body(*req).Execute()
	} else {
		// Use CreateApiKey to create as current authenticated user
		resp, _, err = rc.APIKeysApi.CreateApiKey(ctx).Body(*req).Execute()
	}

	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("key", resp.Data.GetKey())
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(id)

	return diags
}

func getApiKey(ctx context.Context, rc *rockset.RockClient, name string, user string) (*openapi.ApiKey, error) {
	// There's no get endpoint in this api
	// We must list and find the api key
	var resp openapi.ListApiKeysResponse
	var err error
	if user != "" {
		// Then we need to use ListApiKeysAdmin to list the specified user's keys
		resp, _, err = rc.APIKeysApi.ListApiKeysAdmin(ctx, user).Execute()
	} else {
		// Use ListApiKeys to list the current authentciated user's keys.
		resp, _, err = rc.APIKeysApi.ListApiKeys(ctx).Execute()
	}

	if err != nil {
		return nil, err
	}

	var foundApiKey openapi.ApiKey
	for _, apiKey := range *resp.Data {
		if apiKey.Name == name {
			foundApiKey = apiKey
			break
		}
	}

	if foundApiKey.GetName() == "" { // Failed to find
		return nil, fmt.Errorf("API key not found in list.")
	}

	return &foundApiKey, nil
}

func resourceApiKeyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rc := meta.(*rockset.RockClient)
	var diags diag.Diagnostics

	name, user, err := idToUserAndName(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	foundApiKey, err := getApiKey(ctx, rc, name, user)
	if err != nil {
		return diag.FromErr(err)
	}

	if foundApiKey.GetName() == "" { // Failed to find
		return diag.FromErr(fmt.Errorf("API key not found in list."))
	}

	err = d.Set("name", foundApiKey.GetName())
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("key", foundApiKey.GetKey())
	if err != nil {
		return diag.FromErr(err)
	}

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

	if user != "" {
		// Then we need to use DeleteApiKeyAdmin delete as the specified user
		_, _, err = rc.APIKeysApi.DeleteApiKeyAdmin(ctx, name, user).Execute()
	} else {
		// Use DeleteApiKey to delete as current authenticated user
		_, _, err = rc.APIKeysApi.DeleteApiKey(ctx, name).Execute()
	}

	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

package rockset

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/rockset/rockset-go-client"
	"github.com/rockset/rockset-go-client/openapi"
)

func resourceUser() *schema.Resource {
	return &schema.Resource{
		Description: "Manages a Rockset User.",

		CreateContext: resourceUserCreate,
		ReadContext:   resourceUserRead,
		DeleteContext: resourceUserDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"email": {
				Description: "Email address of the user. Also used to identify the user.",
				Type:        schema.TypeString,
				ForceNew:    true,
				Required:    true,
			},
			"roles": {
				Description: "List of roles for the user. E.g. 'admin', 'member', 'read-only'.",
				Type:        schema.TypeList,
				MinItems:    1, // Api returns 500 error currently if no role is set
				ForceNew:    true,
				Required:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceUserCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rc := meta.(*rockset.RockClient)
	var diags diag.Diagnostics

	email := d.Get("email").(string)
	roles := toStringArray(d.Get("roles").([]interface{}))

	resp, err := rc.CreateUser(ctx, email, roles)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(resp.GetEmail())

	return diags
}

func resourceUserRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rc := meta.(*rockset.RockClient)
	var diags diag.Diagnostics

	email := d.Id()

	user, err := getUserByEmail(ctx, rc, email)
	if err != nil {
		return checkForNotFoundError(d, err)
	}

	err = d.Set("email", user.GetEmail())
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("roles", user.GetRoles())
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceUserDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rc := meta.(*rockset.RockClient)
	var diags diag.Diagnostics

	email := d.Id()
	err := rc.DeleteUser(ctx, email)

	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func getUserByEmail(ctx context.Context, rc *rockset.RockClient, email string) (*openapi.User, error) {
	// The api currently has no get user method
	users, err := rc.ListUsers(ctx)
	if err != nil {
		return nil, err
	}

	var foundUser openapi.User
	for _, currentUser := range users {
		if currentUser.Email == email {
			foundUser = currentUser
			break
		}
	}

	if foundUser.GetEmail() == "" { // Failed to find
		return nil, fmt.Errorf("user not found in user list")
	}

	return &foundUser, nil
}

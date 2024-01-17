package rockset

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/rockset/rockset-go-client"
	"github.com/rockset/rockset-go-client/option"
)

func resourceUser() *schema.Resource {
	return &schema.Resource{
		Description: `Manages a Rockset User.

First and last name can only be managed for users who have accepted the invite,
i.e. when the state is ACCEPTED.`,

		CreateContext: resourceUserCreate,
		ReadContext:   resourceUserRead,
		DeleteContext: resourceUserDelete,
		UpdateContext: resourceUserUpdate,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"created_at": {
				Description: "The ISO-8601 time of when the user was created.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"email": {
				Description: "Email address of the user. Also used to identify the user.",
				Type:        schema.TypeString,
				ForceNew:    true,
				Required:    true,
			},
			"first_name": {
				Description: "User's first name. This can only be set once the state is ACTIVE, " +
					"i.e after the user has accepted the invite.",
				Type:     schema.TypeString,
				Optional: true,
			},
			"last_name": {
				Description: "User's last name. This can only be set once the state is ACTIVE, " +
					"i.e after the user has accepted the invite.",
				Type:     schema.TypeString,
				Optional: true,
			},
			"roles": {
				Description: "List of roles for the user. E.g. 'admin', 'member', 'read-only'.",
				Type:        schema.TypeList,
				MinItems:    1, // Api returns 500 error currently if no role is set
				Required:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"state": {
				Description: "State of the user, either NEW or ACTIVE.",
				Type:        schema.TypeString,
				Computed:    true,
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
		return DiagFromErr(err)
	}

	d.SetId(resp.GetEmail())

	return diags
}

func resourceUserRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rc := meta.(*rockset.RockClient)
	var diags diag.Diagnostics

	email := d.Id()

	user, err := rc.GetUser(ctx, email)
	if err != nil {
		return checkForNotFoundError(d, err)
	}

	err = d.Set("email", user.GetEmail())
	if err != nil {
		return DiagFromErr(err)
	}

	if name, ok := user.GetFirstNameOk(); ok {
		if err = d.Set("first_name", name); err != nil {
			return DiagFromErr(err)
		}
	}

	if name, ok := user.GetLastNameOk(); ok {
		if err = d.Set("last_name", name); err != nil {
			return DiagFromErr(err)
		}
	}

	if createdAt, ok := user.GetCreatedAtOk(); ok {
		if err = d.Set("created_at", createdAt); err != nil {
			return DiagFromErr(err)
		}
	}

	if state, ok := user.GetStateOk(); ok {
		if err = d.Set("state", state); err != nil {
			return DiagFromErr(err)
		}
	}

	if err = d.Set("roles", user.GetRoles()); err != nil {
		return DiagFromErr(err)
	}

	return diags
}

func resourceUserDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rc := meta.(*rockset.RockClient)
	var diags diag.Diagnostics

	email := d.Id()
	err := rc.DeleteUser(ctx, email)
	if err != nil {
		return DiagFromErr(err)
	}

	return diags
}

func resourceUserUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rc := meta.(*rockset.RockClient)
	var diags diag.Diagnostics

	email := d.Id()
	roles := toStringArray(d.Get("roles").([]interface{}))
	names := []option.UserOption{
		option.WithUserFirstName(d.Get("first_name").(string)),
		option.WithUserLastName(d.Get("last_name").(string)),
	}

	_, err := rc.UpdateUser(ctx, email, roles, names...)
	if err != nil {
		return DiagFromErr(err)
	}

	return diags
}

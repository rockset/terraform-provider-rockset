package rockset

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/rockset/rockset-go-client"
	"github.com/rockset/rockset-go-client/openapi"
)

func dataSourceRocksetUser() *schema.Resource {
	return &schema.Resource{
		Description: `This data source can be used to fetch information about a specific user.`,
		ReadContext: dataSourceReadRocksetUser,

		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The user ID, in the form of the `email`.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"email": {
				Description: "User email. If absent or blank, it gets the current user.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"first_name": {
				Description: "User's first name.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"last_name": {
				Description: "User's last name.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"roles": {
				Description: "List of roles for the user. E.g. 'admin', 'member', 'read-only'.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"state": {
				Description: "State of the user, either NEW or ACTIVE.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		}}
}

func dataSourceReadRocksetUser(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rc := meta.(*rockset.RockClient)
	var diags diag.Diagnostics
	var err error
	var user openapi.User

	email := d.Get("email").(string)

	if email == "" {
		user, err = rc.GetCurrentUser(ctx)
	} else {
		user, err = rc.GetUser(ctx, email)
	}

	if err != nil {
		return diag.FromErr(err)
	}

	if err = d.Set("email", user.GetEmail()); err != nil {
		return diag.FromErr(err)
	}

	if err = d.Set("first_name", user.GetFirstName()); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("last_name", user.GetLastName()); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("roles", user.GetRoles()); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("state", user.GetState()); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(user.GetEmail())

	return diags
}

package rockset

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/rockset/rockset-go-client"
	models "github.com/rockset/rockset-go-client/lib/go"
)

func resourceWorkspace() *schema.Resource {
	return &schema.Resource{
		Create: resourceWorkspaceCreate,
		Read:   resourceWorkspaceRead,
		Delete: resourceWorkspaceDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
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

func resourceWorkspaceCreate(d *schema.ResourceData, m interface{}) error {
	rc := m.(*rockset.RockClient)

	name := d.Get("name").(string)

	request := models.CreateWorkspaceRequest{
		Name: name,
		Description: d.Get("description").(string),
	}

	_, _, err := rc.Workspaces.Create(request)
	if err != nil {
		return asSwaggerMessage(err)
	}

	d.SetId(name)

	return resourceWorkspaceRead(d, m)
}

func resourceWorkspaceRead(d *schema.ResourceData, m interface{}) error {
	rc := m.(*rockset.RockClient)

	res, _, err := rc.Workspaces.Get(d.Id())
	if err != nil {
		return asSwaggerMessage(err)
	}


	d.Set("name", res.Data.Name)
	d.Set("description", res.Data.Description)
	d.Set("created_by", res.Data.CreatedBy)

	return nil
}

func resourceWorkspaceDelete(d *schema.ResourceData, m interface{}) error {
	rc := m.(*rockset.RockClient)

	name := d.Get("name").(string)

	_, _, err := rc.Workspaces.Delete(name)
	if err != nil {
		return asSwaggerMessage(err)
	}

	d.SetId("")
	return nil
}

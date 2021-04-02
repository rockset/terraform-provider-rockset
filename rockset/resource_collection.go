package rockset

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/rockset/rockset-go-client"
	models "github.com/rockset/rockset-go-client/lib/go"
	"log"
	"net/http"
)

func resourceCollection() *schema.Resource {
	return &schema.Resource{
		Create: resourceCollectionCreate,
		Read:   resourceCollectionRead,
		Delete: resourceCollectionDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
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
			"workspace": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
		},
	}
}

func resourceCollectionCreate(d *schema.ResourceData, m interface{}) error {
	// https://docs.rockset.com/rest-api/#createcollection
	rc := m.(*rockset.RockClient)

	workspace := d.Get("workspace").(string)
	name := d.Get("name").(string)

	request := models.CreateCollectionRequest{
		Name:        name,
		Description: d.Get("description").(string),
	}

	log.Println(spew.Sdump(request))
	_, _, err := rc.Collection.Create(workspace, request)
	if err != nil {
		return asSwaggerMessage(err)
	}

	d.SetId(toID(workspace, name))

	return resourceCollectionRead(d, m)
}

func resourceCollectionRead(d *schema.ResourceData, m interface{}) error {
	// https://docs.rockset.com/rest-api/#getcollection
	rc := m.(*rockset.RockClient)

	workspace, name := workspaceAndNameFromID(d.Id())

	res, _, err := rc.Collection.Get(workspace, name)
	if err != nil {
		return asSwaggerMessage(err)
	}

	if len(res.Data.Sources) != 0 {
		return fmt.Errorf("expected %s to have no source, got %d", name, len(res.Data.Sources))
	}

	d.Set("name", res.Data.Name)
	d.Set("description", res.Data.Description)

	d.SetId(toID(workspace, name))

	return nil
}

func resourceCollectionDelete(d *schema.ResourceData, m interface{}) error {
	// https://docs.rockset.com/rest-api/#deletecollection
	rc := m.(*rockset.RockClient)

	workspace := d.Get("workspace").(string)
	name := d.Get("name").(string)

	_, _, err := rc.Collection.Delete(workspace, name)
	if err != nil {
		return asSwaggerMessage(err)
	}

	d.SetId("")

	// loop until the collection is gone as the deletion is asynchronous
	err = resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		log.Printf("checking if %s in workspace %s still exist", name, workspace)
		_, httpResp, err := rc.Collection.Get(workspace, name)
		if err == nil {
			return resource.RetryableError(fmt.Errorf("collection %s still exist", name))
		}
		if httpResp.StatusCode == http.StatusNotFound {
			return nil
		}
		return resource.NonRetryableError(err)
	})
	if err != nil {
		log.Println(err)
	}

	return err
}

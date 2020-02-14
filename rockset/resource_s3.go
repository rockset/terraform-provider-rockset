package rockset

import (
	"encoding/json"
	"errors"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/rockset/rockset-go-client"
	models "github.com/rockset/rockset-go-client/lib/go"
	"log"
	"net/http"
)

// https://www.terraform.io/docs/extend/writing-custom-providers.html

func resourceS3() *schema.Resource {
	return &schema.Resource{
		Create: resourceS3IntegrationCreate,
		Read:   resourceS3IntegrationRead,
		Delete: resourceS3IntegrationDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"description": &schema.Schema{
				Type:     schema.TypeString,
				Default:  "created by Rockset terraform provider",
				ForceNew: true,
				Optional: true,
			},
			"aws_role_arn": &schema.Schema{
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
		},
	}
}

func resourceS3IntegrationCreate(d *schema.ResourceData, m interface{}) error {
	// https://docs.rockset.com/rest-api#createintegration
	rc := m.(*rockset.RockClient)

	req := models.CreateIntegrationRequest{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		S3: &models.S3Integration{
			AwsRole: &models.AwsRole{AwsRoleArn: d.Get("aws_role_arn").(string)},
		},
	}

	res, _, err := rc.Integration.Create(req)
	if err != nil {
		return asSwaggerMessage(err)
	}

	d.SetId(res.Data.Name)

	return resourceS3IntegrationRead(d, m)
}

// TODO: this type must be defined in models, find it
type swaggerErrorMessage struct {
	Message string `json:"message"`
	Type    string `json:"type"`
}

func asSwaggerMessage(err error) error {
	var e models.GenericSwaggerError
	if errors.As(err, &e) {
		var msg swaggerErrorMessage
		if err = json.Unmarshal(e.Body(), &msg); err != nil {
			log.Printf("failed to unmarshal error response: %v", err)
		}
		return errors.New(msg.Message)
	}

	return err
}

func resourceS3IntegrationRead(d *schema.ResourceData, m interface{}) error {
	// https://docs.rockset.com/rest-api#getintegration
	rc := m.(*rockset.RockClient)

	id := d.Id()

	log.Printf("getting integration with ID %s", id)
	res, httpResp, err := rc.Integration.Get(id)
	if err != nil {
		return asSwaggerMessage(err)
	}

	if httpResp.StatusCode == http.StatusNotFound {
		d.SetId("")
		return nil
	}
	log.Printf("%+v", *res.Data)

	if res.Data.S3 == nil && res.Data.S3.AwsRole == nil {
		d.SetId("")
		return nil
	}

	err = d.Set("aws_role_arn", res.Data.S3.AwsRole.AwsRoleArn)
	if err != nil {
		return nil
	}

	err = d.Set("name", res.Data.Name)
	if err != nil {
		return nil
	}

	err = d.Set("description", res.Data.Description)
	if err != nil {
		return nil
	}

	return nil
}

func resourceS3IntegrationDelete(d *schema.ResourceData, m interface{}) error {
	// https://docs.rockset.com/rest-api#deleteintegration
	name := d.Get("name").(string)

	rc := m.(*rockset.RockClient)

	log.Printf("deleting integration %s", name)
	_, _, err := rc.Integration.Delete(name)
	if err != nil {
		return asSwaggerMessage(err)
	}

	return nil
}

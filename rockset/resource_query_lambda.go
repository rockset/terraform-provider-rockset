package rockset

import (
	"errors"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/rockset/rockset-go-client"
	models "github.com/rockset/rockset-go-client/lib/go"
	"log"
)

func resourceQueryLambda() *schema.Resource {
	return &schema.Resource{
		Create: resourceQueryLambdaCreate,
		Read:   resourceQueryLambdaRead,
		Delete: resourceQueryLambdaDelete,
		Update: resourceQueryLambdaUpdate,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"workspace": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: rocksetNameValidator,
			},
			"description": {
				Type:     schema.TypeString,
				Default:  "created by Rockset terraform provider",
				Optional: true,
			},
			"version": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"sql": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"query": {
							Type:     schema.TypeString,
							Required: true,
						},
						"default_parameter": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:     schema.TypeString,
										Required: true,
									},
									"type": {
										Type:     schema.TypeString,
										Required: true,
									},
									"value": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func resourceQueryLambdaCreate(d *schema.ResourceData, m interface{}) error {
	// https://docs.rockset.com/rest-api/#createquerylambda
	rc := m.(*rockset.RockClient)

	workspace := d.Get("workspace").(string)
	name := d.Get("name").(string)

	req := models.CreateQueryLambdaRequest{
		Name:        name,
		Description: d.Get("description").(string),
		Sql:         makeQueryLambdaSQL(d.Get("sql")),
	}
	resp, _, err := rc.QueryLambdas.Create(workspace, req)
	if err != nil {
		return asSwaggerMessage(err)
	}

	d.SetId(toID(workspace, name))
	log.Println(spew.Sdump(resp))
	return nil
}

func resourceQueryLambdaUpdate(d *schema.ResourceData, m interface{}) error {
	// https://docs.rockset.com/rest-api/#updatequerylambda
	rc := m.(*rockset.RockClient)

	workspace := d.Get("workspace").(string)
	name := d.Get("name").(string)

	req := models.UpdateQueryLambdaRequest{
		Description: d.Get("description").(string),
		Sql:         makeQueryLambdaSQL(d.Get("sql")),
	}
	log.Println(spew.Sdump(req))

	resp, _, err := rc.QueryLambdas.Update(workspace, name, req)
	if err != nil {
		return asSwaggerMessage(err)
	}

	d.SetId(toID(workspace, name))
	log.Println(spew.Sdump(resp))
	return nil
}

func resourceQueryLambdaRead(d *schema.ResourceData, m interface{}) error {
	// https://docs.rockset.com/rest-api/#getquerylambdaversion
	rc := m.(*rockset.RockClient)

	workspace, name := workspaceAndNameFromID(d.Id())

	res, _, err := rc.QueryLambdas.List_1(workspace)
	if err != nil {
		if re, ok := rockset.AsRocksetError(err); ok {
			return errors.New(re.Message)
		}
		return err
	}

	for _, ql := range res.Data {
		if ql.Name == name {
			d.Set("workspace", ql.Workspace)
			d.Set("name", ql.Name)
			d.Set("description", ql.Description)
			d.Set("version", ql.Version)
			d.Set("state", ql.State)
			d.Set("sql", flattenQueryLambdaSQL(ql.Sql))

			d.SetId(toID(workspace, name))
			return nil
		}
	}

	return fmt.Errorf("query lambda %s not found in workspace %s", name, workspace)
}

func resourceQueryLambdaDelete(d *schema.ResourceData, m interface{}) error {
	// https://docs.rockset.com/rest-api/#deletequerylambda
	rc := m.(*rockset.RockClient)

	workspace := d.Get("workspace").(string)
	name := d.Get("name").(string)

	_, _, err := rc.QueryLambdas.Delete(workspace, name)
	if err != nil {
		return asSwaggerMessage(err)
	}

	d.SetId("")

	return nil
}

func makeQueryLambdaSQL(in interface{}) *models.QueryLambdaSql {
	sql := models.QueryLambdaSql{}

	if set, ok := in.(*schema.Set); ok {
		for _, s := range set.List() {
			if i, ok := s.(map[string]interface{}); ok {
				for k, v := range i {
					log.Printf("%s: %s", k, v)
					log.Printf("v: %T", v)
					switch k {
					case "query":
						sql.Query = v.(string)
					case "default_parameter":
						sql.DefaultParameters = append(sql.DefaultParameters, makeDefaultParameters(v))
					}
				}
			}
		}
	}

	return &sql
}

func makeDefaultParameters(input interface{}) models.QueryParameter {
	dp := models.QueryParameter{}

	if in, ok := input.([]interface{}); ok {
		for _, i := range in {
			if cfg, ok := i.(map[string]interface{}); ok {
				for k, v := range cfg {
					switch k {
					case "name":
						dp.Name = v.(string)
					case "type":
						dp.Type_ = v.(string)
					case "value":
						dp.Value = v.(string)
					}
				}
			}
		}
	}

	return dp
}

func flattenQueryLambdaSQL(sql *models.QueryLambdaSql) []interface{} {
	var m = make(map[string]interface{})
	m["query"] = sql.Query

	var r []interface{}
	for _, qp := range sql.DefaultParameters {
		m := make(map[string]interface{})
		m["name"] = qp.Name
		m["type"] = qp.Type_
		m["value"] = qp.Value
		r = append(r, m)
	}
	m["default_parameter"] = r

	return []interface{}{m}
}

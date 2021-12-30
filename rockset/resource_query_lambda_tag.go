package rockset

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/rockset/rockset-go-client"
	"github.com/rockset/rockset-go-client/openapi"
)

func resourceQueryLambdaTag() *schema.Resource {
	return &schema.Resource{
		Description: "Manages a Rockset Query Lambda Tag.",

		CreateContext: resourceQueryLambdaTagCreate,
		ReadContext:   resourceQueryLambdaTagRead,
		DeleteContext: resourceQueryLambdaTagDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Description:  "Unique identifier for the tag. Can contain alphanumeric or dash characters.",
				Type:         schema.TypeString,
				ForceNew:     true,
				Required:     true,
				ValidateFunc: rocksetNameValidator,
			},
			"workspace": {
				Description: "The name of the workspace the query lambda is in.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"query_lambda": {
				Description:  "Unique identifier for the query lambda. Can contain alphanumeric or dash characters.",
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: rocksetNameValidator,
			},
			"version": {
				Description:  "Version of the query lambda this tag should point to.",
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: rocksetNameValidator,
			},
		},
	}
}

func toQueryLambdaTagID(workspace string, queryLambdaName string, tagName string) string {
	return fmt.Sprintf("%s.%s.%s", workspace, queryLambdaName, tagName)
}

func fromQueryLambdaTagID(id string) (string, string, string) {
	tokens := strings.SplitN(id, ".", 3)
	if len(tokens) != 3 {
		log.Printf("unparsable id: %s", id)
		return "", "", ""
	}
	return tokens[0], tokens[1], tokens[2]
}

func resourceQueryLambdaTagCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rc := meta.(*rockset.RockClient)
	var diags diag.Diagnostics

	workspace := d.Get("workspace").(string)
	tagName := d.Get("name").(string)
	version := d.Get("version").(string)
	queryLambdaName := d.Get("query_lambda").(string)

	d.SetId(toQueryLambdaTagID(workspace, queryLambdaName, tagName))

	_, err := rc.CreateQueryLambdaTag(ctx, workspace, queryLambdaName, version, tagName)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceQueryLambdaTagRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rc := meta.(*rockset.RockClient)
	var diags diag.Diagnostics

	workspace, queryLambdaName, tagName := fromQueryLambdaTagID(d.Id())

	queryLambdaTag, err := rc.GetQueryLambdaVersionByTag(ctx, workspace, queryLambdaName, tagName)
	if err != nil {
		return diag.FromErr(err)
	}

	err = parseQueryLambdaTag(&queryLambdaTag, d)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceQueryLambdaTagDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rc := meta.(*rockset.RockClient)
	var diags diag.Diagnostics

	workspace, queryLambdaName, tagName := fromQueryLambdaTagID(d.Id())

	ql, err := rc.GetQueryLambdaVersionByTag(ctx, workspace, queryLambdaName, tagName)
	if err != nil {
		return diag.FromErr(err)
	}

	v := ql.GetVersion()
	v.GetVersion()

	err = rc.DeleteQueryLambdaVersion(ctx, workspace, queryLambdaName, v.GetVersion())
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

/*
	Takes in a query lambda tag returned from the api.
	Parses the query lambda tag fields and
	puts them into the schema object.
*/
func parseQueryLambdaTag(queryLambdaTag *openapi.QueryLambdaTag, d *schema.ResourceData) error {

	var err error

	err = d.Set("name", queryLambdaTag.GetTagName())
	if err != nil {
		return err
	}

	err = d.Set("workspace", queryLambdaTag.Version.GetWorkspace())
	if err != nil {
		return err
	}

	err = d.Set("query_lambda", queryLambdaTag.Version.GetName())
	if err != nil {
		return err
	}

	err = d.Set("version", queryLambdaTag.Version.GetVersion())
	if err != nil {
		return err
	}

	return nil // No errors
}

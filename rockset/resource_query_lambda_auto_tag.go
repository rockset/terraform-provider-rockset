package rockset

import (
	"context"
	"errors"
	"regexp"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/rockset/rockset-go-client"
	rockerr "github.com/rockset/rockset-go-client/errors"
)

func resourceQueryLambdaAutoTag() *schema.Resource {
	percentS := regexp.MustCompile("%s")

	return &schema.Resource{
		Description: `Automatically tags a Query Lambda using a name template, which must contain %s, that is replaced with an ever increasing number.

~> On resource delete it will remove all old tags that it created.`,

		CreateContext: resourceQueryLambdaAutoTagCreate,
		ReadContext:   resourceQueryLambdaAutoTagRead,
		DeleteContext: resourceQueryLambdaAutoTagDelete,
		UpdateContext: resourceQueryLambdaTagUpdate,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Unique identifier for the tag, generated from the `template` and `tag_version`.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"template": {
				Description:  "Template for the tag `name`. Can contain alphanumeric or dash characters. Use `%s` to insert the version.",
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(percentS, "must contain %s"),
			},
			"tag_version": {
				Description: "Auto-incremented version number for the tag `name`. Starts at 1 when the tag is created and is incremented every time the tag is updated.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"max_tags": {
				Description:  "Maximum number of previous auto-generated tags to keep.",
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      5,
				ValidateFunc: validation.IntBetween(1, 20),
			},
			"workspace": {
				Description: "The name of the workspace the query lambda is in.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"query_lambda": {
				Description:  "Unique identifier for the query lambda.",
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: rocksetNameValidator,
			},
			"version": {
				Description:  "Version of the query lambda this tag should point to.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: rocksetNameValidator,
			},
			"tags": {
				Description: "List of previous auto-generated tags that are kept in accordance with `max_tags`.",
				Type:        schema.TypeList,
				Computed:    true,
				// this is used to keep track of the tags that were created by this resource, so they can be deleted on destroy
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceQueryLambdaAutoTagCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rc := meta.(*rockset.RockClient)
	var diags diag.Diagnostics

	workspace := d.Get("workspace").(string)
	template := d.Get("template").(string)
	tagName := strings.ReplaceAll(template, "%s", "1")
	version := d.Get("version").(string)
	queryLambdaName := d.Get("query_lambda").(string)

	d.SetId(toQueryLambdaTagID(workspace, queryLambdaName, template))
	_ = d.Set("tag_version", 1)
	_ = d.Set("name", tagName)
	_ = d.Set("tags", []string{}) // the current tag is not included in the list of old tags

	_, err := rc.CreateQueryLambdaTag(ctx, workspace, queryLambdaName, version, tagName)
	if err != nil {
		return DiagFromErr(err)
	}
	tflog.Info(ctx, "Created query lambda auto tag", map[string]interface{}{
		"workspace": workspace,
		"name":      queryLambdaName,
		"tag":       tagName,
		"version":   version,
	})

	return diags
}

func resourceQueryLambdaTagUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rc := meta.(*rockset.RockClient)
	var diags diag.Diagnostics

	workspace, queryLambdaName, template := fromQueryLambdaTagID(d.Id())
	tagVersion := d.Get("tag_version").(int)
	oldTagName := strings.ReplaceAll(template, "%s", strconv.Itoa(tagVersion))
	tagVersion++
	tagName := strings.ReplaceAll(template, "%s", strconv.Itoa(tagVersion))
	version := d.Get("version").(string)

	maxTags := d.Get("max_tags").(int)
	tags := toStringArray(d.Get("tags").([]interface{}))

	var err error
	if len(tags) >= maxTags {
		// remove the oldest tag so we make room for the new one
		oldest := tags[0]
		if err = rc.DeleteQueryLambdaTag(ctx, workspace, queryLambdaName, oldest); err != nil {
			return diag.Errorf("failed to delete oldest tag %s.%s %s: %v", workspace, queryLambdaName, oldest, err)
		}
		tflog.Debug(ctx, "Deleted oldest query lambda auto tag", map[string]interface{}{
			"workspace": workspace,
			"name":      queryLambdaName,
			"tag":       oldest,
		})
		tags = tags[1:]
	}

	if _, err = rc.CreateQueryLambdaTag(ctx, workspace, queryLambdaName, version, tagName); err != nil {
		return DiagFromErr(err)
	}
	_ = d.Set("tag_version", tagVersion)
	_ = d.Set("name", tagName)
	tags = append(tags, oldTagName)
	_ = d.Set("tags", tags)

	tflog.Info(ctx, "Updated query lambda auto tag", map[string]interface{}{
		"workspace": workspace,
		"name":      queryLambdaName,
		"tag":       tagName,
		"version":   version,
		"tags":      tags,
	})

	return diags
}

func resourceQueryLambdaAutoTagRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rc := meta.(*rockset.RockClient)
	var diags diag.Diagnostics

	workspace, queryLambdaName, template := fromQueryLambdaTagID(d.Id())
	tagVersion := d.Get("tag_version").(int)
	tagName := strings.ReplaceAll(template, "%s", strconv.Itoa(tagVersion))

	queryLambdaTag, err := rc.GetQueryLambdaVersionByTag(ctx, workspace, queryLambdaName, tagName)
	if err != nil {
		return checkForNotFoundError(d, err)
	}

	err = parseQueryLambdaTag(&queryLambdaTag, d)
	if err != nil {
		return DiagFromErr(err)
	}

	return diags
}

func resourceQueryLambdaAutoTagDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rc := meta.(*rockset.RockClient)
	var diags diag.Diagnostics

	workspace, queryLambdaName, template := fromQueryLambdaTagID(d.Id())
	tagVersion := d.Get("tag_version").(int)
	tagName := strings.ReplaceAll(template, "%s", strconv.Itoa(tagVersion))

	err := rc.DeleteQueryLambdaTag(ctx, workspace, queryLambdaName, tagName)
	if err != nil {
		return diag.Errorf("failed to delete query lambda auto tag %s.%s:%s: %v", workspace, queryLambdaName, tagName, err)
	}
	tflog.Info(ctx, "Created query lambda auto tag", map[string]interface{}{
		"workspace": workspace,
		"name":      queryLambdaName,
		"tag":       tagName,
		"version":   tagVersion,
	})

	tags := toStringArray(d.Get("tags").([]interface{}))
	for _, tag := range tags {
		if err = rc.DeleteQueryLambdaTag(ctx, workspace, queryLambdaName, tag); err != nil {
			var re rockerr.Error
			if errors.As(err, &re) && re.IsNotFoundError() {
				// we ignore not found errors, as the tag is already gone and we want it gone
				tflog.Error(ctx, "Failed to delete old tag", map[string]interface{}{
					"workspace": workspace,
					"name":      queryLambdaName,
					"tag":       tag,
					"error":     err,
				})
				continue
			}

			return diag.Errorf("failed to delete old tag %s: %v", tag, err)
		}
	}

	return diags
}

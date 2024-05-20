package rockset

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccQueryLambdaAutoTag_Basic(t *testing.T) {
	tagName := randomName("tag_%s")
	v1Tag := strings.ReplaceAll(tagName, "%s", "1")
	v2Tag := strings.ReplaceAll(tagName, "%s", "2")
	v3Tag := strings.ReplaceAll(tagName, "%s", "3")
	v4Tag := strings.ReplaceAll(tagName, "%s", "4")

	v1 := Values{
		Name:        randomName("ql"),
		Tag:         tagName, // used for template
		Alias:       "commons._events",
		Workspace:   randomName("ws"),
		Description: description(),
		SQL:         "SELECT 1",
	}
	v2 := v1
	v2.SQL = "SELECT 2"
	v3 := v1
	v3.SQL = "SELECT 3"
	v4 := v1
	v4.SQL = "SELECT 4"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRocksetQueryLambdaTagDestroy,
		Steps: []resource.TestStep{
			{
				Config: getHCLTemplate("query_lambda_auto_tag.tf", v1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("rockset_query_lambda.test", "name", v1.Name),
					resource.TestCheckResourceAttr("rockset_query_lambda.test", "description", v1.Description),
					resource.TestCheckResourceAttrSet("rockset_query_lambda.test", "version"),
					resource.TestCheckResourceAttrSet("rockset_query_lambda.test", "state"),
					resource.TestCheckResourceAttr("rockset_query_lambda_auto_tag.test", "name", v1Tag),
					resource.TestCheckResourceAttr("rockset_query_lambda_auto_tag.test", "tag_version", "1"),
					resource.TestCheckResourceAttr("rockset_query_lambda_auto_tag.test", "workspace", "acc"),
					resource.TestCheckResourceAttrSet("rockset_query_lambda_auto_tag.test", "version"),
					resource.TestCheckResourceAttr("rockset_query_lambda_auto_tag.test", "tags.#", "0"),
				),
			},
			{
				Config: getHCLTemplate("query_lambda_auto_tag.tf", v2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("rockset_query_lambda.test", "name", v2.Name),
					resource.TestCheckResourceAttr("rockset_query_lambda.test", "description", v2.Description),
					resource.TestCheckResourceAttrSet("rockset_query_lambda.test", "version"),
					resource.TestCheckResourceAttrSet("rockset_query_lambda.test", "state"),
					resource.TestCheckResourceAttr("rockset_query_lambda_auto_tag.test", "name", v2Tag),
					resource.TestCheckResourceAttr("rockset_query_lambda_auto_tag.test", "tag_version", "2"),
					resource.TestCheckResourceAttr("rockset_query_lambda_auto_tag.test", "workspace", "acc"),
					resource.TestCheckResourceAttrSet("rockset_query_lambda_auto_tag.test", "version"),
					resource.TestCheckResourceAttr("rockset_query_lambda_auto_tag.test", "tags.#", "1"),
				),
			},
			{
				Config: getHCLTemplate("query_lambda_auto_tag.tf", v3),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("rockset_query_lambda.test", "name", v3.Name),
					resource.TestCheckResourceAttr("rockset_query_lambda.test", "description", v3.Description),
					resource.TestCheckResourceAttr("rockset_query_lambda_auto_tag.test", "name", v3Tag),
					resource.TestCheckResourceAttr("rockset_query_lambda_auto_tag.test", "tag_version", "3"),
					resource.TestCheckResourceAttr("rockset_query_lambda_auto_tag.test", "tags.#", "2"),
				),
			},
			{
				Config: getHCLTemplate("query_lambda_auto_tag.tf", v4),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("rockset_query_lambda.test", "name", v4.Name),
					resource.TestCheckResourceAttr("rockset_query_lambda.test", "description", v4.Description),
					resource.TestCheckResourceAttr("rockset_query_lambda_auto_tag.test", "name", v4Tag),
					resource.TestCheckResourceAttr("rockset_query_lambda_auto_tag.test", "tag_version", "4"),
					resource.TestCheckResourceAttr("rockset_query_lambda_auto_tag.test", "tags.#", "2"),
				),
			},
		},
	})
}

package rockset

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/rockset/rockset-go-client"
	"github.com/rockset/rockset-go-client/openapi"
)

const testQueryLambdaNameTagTest = "tpat-ql-diff"
const testQueryLambdaTagName = "test"

func TestAccQueryLambdaTag_Basic(t *testing.T) {
	var queryLambdaTag openapi.QueryLambdaTag

	type values struct {
		Query string
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRocksetQueryLambdaTagDestroy,
		Steps: []resource.TestStep{
			{
				Config: getHCLTemplate("query_lambda_no_defaults.tf", values{"SELECT 1"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRocksetQueryLambdaTagExists("rockset_query_lambda_tag.test", &queryLambdaTag),
					resource.TestCheckResourceAttr("rockset_query_lambda.test", "name", testQueryLambdaNameTagTest),
					resource.TestCheckResourceAttr("rockset_query_lambda.test", "description", "basic lambda"),
					resource.TestCheckResourceAttrSet("rockset_query_lambda.test", "version"),
					resource.TestCheckResourceAttrSet("rockset_query_lambda.test", "state"),
					resource.TestCheckResourceAttr("rockset_query_lambda_tag.test", "name", testQueryLambdaTagName),
					resource.TestCheckResourceAttr("rockset_query_lambda_tag.test", "workspace", "commons"),
					resource.TestCheckResourceAttrSet("rockset_query_lambda_tag.test", "version"),
				),
				ExpectNonEmptyPlan: false,
				Destroy:            false,
			},
			{
				Config: getHCLTemplate("query_lambda_no_defaults.tf", values{"SELECT 2"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRocksetQueryLambdaTagExists("rockset_query_lambda_tag.test", &queryLambdaTag),
					resource.TestCheckResourceAttr("rockset_query_lambda.test", "name", testQueryLambdaNameTagTest),
					resource.TestCheckResourceAttr("rockset_query_lambda.test", "description", "basic lambda"),
					resource.TestCheckResourceAttrSet("rockset_query_lambda.test", "version"),
					resource.TestCheckResourceAttrSet("rockset_query_lambda.test", "state"),
					resource.TestCheckResourceAttr("rockset_query_lambda_tag.test", "name", testQueryLambdaTagName),
					resource.TestCheckResourceAttr("rockset_query_lambda_tag.test", "workspace", "commons"),
					resource.TestCheckResourceAttrSet("rockset_query_lambda_tag.test", "version"),
				),
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func testAccCheckRocksetQueryLambdaTagDestroy(s *terraform.State) error {
	rc := testAccProvider.Meta().(*rockset.RockClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "rockset_query_lambda_tag" {
			continue
		}

		workspace, queryLambdaName, tagName := fromQueryLambdaTagID(rs.Primary.ID)
		_, err := rc.GetQueryLambdaVersionByTag(testCtx, workspace, queryLambdaName, tagName)

		// An error would mean we didn't find the key, we expect an error
		if err == nil {
			// We did not get an error, so we failed to delete the key.
			return fmt.Errorf("query Lambda %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckRocksetQueryLambdaTagExists(resource string,
	queryLambdaTag *openapi.QueryLambdaTag) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rc := testAccProvider.Meta().(*rockset.RockClient)

		rs, err := getResourceFromState(state, resource)
		if err != nil {
			return err
		}

		workspace, queryLambdaName, tagName := fromQueryLambdaTagID(rs.Primary.ID)

		resp, err := rc.GetQueryLambdaVersionByTag(testCtx, workspace, queryLambdaName, tagName)
		if err != nil {
			return err
		}

		*queryLambdaTag = resp

		return nil
	}
}

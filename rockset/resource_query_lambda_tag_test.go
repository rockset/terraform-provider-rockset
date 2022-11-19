package rockset

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/rockset/rockset-go-client"
	"github.com/rockset/rockset-go-client/openapi"
)

func TestAccQueryLambdaTag_Basic(t *testing.T) {
	var queryLambdaTag1, queryLambdaTag2 openapi.QueryLambdaTag

	v1 := Values{
		Name:        randomName("ql"),
		Tag:         randomName("tag"),
		Alias:       "commons._events",
		Workspace:   randomName("ws"),
		Description: description(),
		SQL:         "SELECT 1",
	}
	v2 := v1
	v2.SQL = "SELECT 2"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRocksetQueryLambdaTagDestroy,
		Steps: []resource.TestStep{
			{
				Config: getHCLTemplate("query_lambda_no_defaults.tf", v1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRocksetQueryLambdaTagExists("rockset_query_lambda_tag.test", &queryLambdaTag1),
					resource.TestCheckResourceAttr("rockset_query_lambda.test", "name", v1.Name),
					resource.TestCheckResourceAttr("rockset_query_lambda.test", "description", v1.Description),
					resource.TestCheckResourceAttrSet("rockset_query_lambda.test", "version"),
					resource.TestCheckResourceAttrSet("rockset_query_lambda.test", "state"),
					resource.TestCheckResourceAttr("rockset_query_lambda_tag.test", "name", v1.Tag),
					resource.TestCheckResourceAttr("rockset_query_lambda_tag.test", "workspace", "acc"),
					resource.TestCheckResourceAttrSet("rockset_query_lambda_tag.test", "version"),
				),
				ExpectNonEmptyPlan: false,
				Destroy:            false,
			},
			{
				Config: getHCLTemplate("query_lambda_no_defaults.tf", v2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRocksetQueryLambdaTagExists("rockset_query_lambda_tag.test", &queryLambdaTag2),
					resource.TestCheckResourceAttr("rockset_query_lambda.test", "name", v2.Name),
					resource.TestCheckResourceAttr("rockset_query_lambda.test", "description", v2.Description),
					resource.TestCheckResourceAttrSet("rockset_query_lambda.test", "version"),
					resource.TestCheckResourceAttrSet("rockset_query_lambda.test", "state"),
					resource.TestCheckResourceAttr("rockset_query_lambda_tag.test", "name", v2.Tag),
					resource.TestCheckResourceAttr("rockset_query_lambda_tag.test", "workspace", "acc"),
					resource.TestCheckResourceAttrSet("rockset_query_lambda_tag.test", "version"),
					testAccCheckRocksetQueryLambdaTagDifferent(&queryLambdaTag1, &queryLambdaTag2),
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

func testAccCheckRocksetQueryLambdaTagDifferent(qlt1, qlt2 *openapi.QueryLambdaTag) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		v1 := qlt1.GetVersion()
		v2 := qlt2.GetVersion()
		if v1.GetVersion() == v2.GetVersion() {
			return fmt.Errorf("expected query lambda version (%s, %s) to have changed ", v1.GetVersion(), v2.GetVersion())
		}

		return nil
	}
}

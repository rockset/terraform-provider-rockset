package rockset

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/rockset/rockset-go-client"
	"github.com/rockset/rockset-go-client/openapi"
	"github.com/stretchr/testify/assert"
)

const testQueryLambdaName = "terraform-provider-acceptance-tests-query-lambda-basic"

func TestAccQueryLambda_Basic(t *testing.T) {
	var queryLambda openapi.QueryLambda

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRocksetQueryLambdaDestroy,
		Steps: []resource.TestStep{
			{
				Config: getHCL("query_lambda_basic.tf"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRocksetQueryLambdaExists("rockset_query_lambda.test", &queryLambda),
					resource.TestCheckResourceAttr("rockset_query_lambda.test", "name", testQueryLambdaName),
					resource.TestCheckResourceAttr("rockset_query_lambda.test", "description", "basic lambda"),
					testAccCheckSql(t, &queryLambda, "SELECT * FROM commons._events WHERE "+
						"_events._event_time > :start AND _events._event_time < :end LIMIT 1"),
					resource.TestCheckResourceAttrSet("rockset_query_lambda.test", "version"),
					resource.TestCheckResourceAttrSet("rockset_query_lambda.test", "state"),
				),
				ExpectNonEmptyPlan: false,
			},
			{
				Config: getHCL("query_lambda_basic_updated.tf"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRocksetQueryLambdaExists("rockset_query_lambda.test", &queryLambda),
					resource.TestCheckResourceAttr("rockset_query_lambda.test", "name", testQueryLambdaName),
					resource.TestCheckResourceAttr("rockset_query_lambda.test", "description", "updated description"),
					testAccCheckSql(t, &queryLambda, "SELECT * FROM commons._events WHERE "+
						"_events._event_time > :start AND _events._event_time < :end LIMIT 2"),
					resource.TestCheckResourceAttrSet("rockset_query_lambda.test", "version"),
					resource.TestCheckResourceAttrSet("rockset_query_lambda.test", "state"),
				),
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccQueryLambda_NoDefaults(t *testing.T) {
	var queryLambda openapi.QueryLambda
	type values struct {
		Query string
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRocksetQueryLambdaDestroy,
		Steps: []resource.TestStep{
			{
				Config: getHCLTemplate("query_lambda_no_defaults.tf", values{"SELECT 1"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRocksetQueryLambdaExists("rockset_query_lambda.test", &queryLambda),
					resource.TestCheckResourceAttr("rockset_query_lambda.test", "name", "tpat-ql-diff"),
					resource.TestCheckResourceAttr("rockset_query_lambda.test", "description", "basic lambda"),
					testAccCheckSql(t, &queryLambda, "SELECT 1"),
					resource.TestCheckResourceAttrSet("rockset_query_lambda.test", "version"),
					resource.TestCheckResourceAttrSet("rockset_query_lambda.test", "state"),
				),
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccQueryLambda_Recreate(t *testing.T) {
	var queryLambda openapi.QueryLambda
	type values struct {
		Query string
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRocksetQueryLambdaDestroy,
		Steps: []resource.TestStep{
			{
				Config: getHCLTemplate("query_lambda_no_defaults.tf", values{"SELECT 1"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRocksetQueryLambdaExists("rockset_query_lambda.test", &queryLambda),
					resource.TestCheckResourceAttr("rockset_query_lambda.test", "name", "tpat-ql-diff"),
					resource.TestCheckResourceAttr("rockset_query_lambda.test", "description", "basic lambda"),
					testAccCheckSql(t, &queryLambda, "SELECT 1"),
					resource.TestCheckResourceAttrSet("rockset_query_lambda.test", "version"),
					resource.TestCheckResourceAttrSet("rockset_query_lambda.test", "state"),
				),
				ExpectNonEmptyPlan: false,
				Destroy:            false,
			},
		},
	})
}

func testAccCheckRocksetQueryLambdaDestroy(s *terraform.State) error {
	rc := testAccProvider.Meta().(*rockset.RockClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "rockset_query_lambda" {
			continue
		}

		workspace, name := workspaceAndNameFromID(rs.Primary.ID)
		_, err := getQueryLambda(testCtx, rc, workspace, name)

		// An error would mean we didn't find the key, we expect an error
		if err == nil {
			// We did not get an error, so we failed to delete the key.
			return fmt.Errorf("query Lambda %s still exists", name)
		}
	}

	return nil
}

func testAccCheckRocksetQueryLambdaExists(resource string, queryLambda *openapi.QueryLambda) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rc := testAccProvider.Meta().(*rockset.RockClient)

		rs, err := getResourceFromState(state, resource)
		if err != nil {
			return err
		}

		workspace, name := workspaceAndNameFromID(rs.Primary.ID)

		resp, err := getQueryLambda(testCtx, rc, workspace, name)
		if err != nil {
			return err
		}

		*queryLambda = *resp

		return nil
	}
}

func testAccCheckSql(t *testing.T, queryLambda *openapi.QueryLambda, expectedSql string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		sql := queryLambda.LatestVersion.Sql.Query

		assert.Equal(t, expectedSql, sql, "SQL string didn't match.")

		return nil
	}
}

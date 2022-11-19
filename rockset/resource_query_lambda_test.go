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

func TestAccQueryLambda_Basic(t *testing.T) {
	var queryLambda openapi.QueryLambda

	sql := "SELECT * FROM commons._events WHERE _events._event_time > :start AND _events._event_time < :end "
	v1 := Values{
		Name:        randomName("ql"),
		Tag:         randomName("tag"),
		Alias:       "commons._events",
		Workspace:   randomName("ws"),
		Description: description(),
		SQL:         sql + "LIMIT 1",
	}
	v2 := v1
	v2.SQL = sql + "LIMIT 2"
	v2.Description = description()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRocksetQueryLambdaDestroy,
		Steps: []resource.TestStep{
			{
				Config: getHCLTemplate("query_lambda_basic.tf", v1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRocksetQueryLambdaExists("rockset_query_lambda.test", &queryLambda),
					resource.TestCheckResourceAttr("rockset_query_lambda.test", "name", v1.Name),
					resource.TestCheckResourceAttr("rockset_query_lambda.test", "description", v1.Description),
					testAccCheckSql(t, &queryLambda, v1.SQL),
					resource.TestCheckResourceAttrSet("rockset_query_lambda.test", "version"),
					resource.TestCheckResourceAttrSet("rockset_query_lambda.test", "state"),
				),
				ExpectNonEmptyPlan: false,
			},
			{
				Config: getHCLTemplate("query_lambda_basic.tf", v2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRocksetQueryLambdaExists("rockset_query_lambda.test", &queryLambda),
					resource.TestCheckResourceAttr("rockset_query_lambda.test", "name", v2.Name),
					resource.TestCheckResourceAttr("rockset_query_lambda.test", "description", v2.Description),
					testAccCheckSql(t, &queryLambda, v2.SQL),
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

	v1 := Values{
		Name:        randomName("ql"),
		Description: description(),
		SQL:         "SELECT 1",
		Tag:         randomName("tag"),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRocksetQueryLambdaDestroy,
		Steps: []resource.TestStep{
			{
				Config: getHCLTemplate("query_lambda_no_defaults.tf", v1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRocksetQueryLambdaExists("rockset_query_lambda.test", &queryLambda),
					resource.TestCheckResourceAttr("rockset_query_lambda.test", "name", v1.Name),
					resource.TestCheckResourceAttr("rockset_query_lambda.test", "description", v1.Description),
					testAccCheckSql(t, &queryLambda, v1.SQL),
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

	v1 := Values{
		Name:        randomName("ql"),
		Description: description(),
		SQL:         "SELECT 1",
		Tag:         randomName("tag"),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRocksetQueryLambdaDestroy,
		Steps: []resource.TestStep{
			{
				Config: getHCLTemplate("query_lambda_no_defaults.tf", v1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRocksetQueryLambdaExists("rockset_query_lambda.test", &queryLambda),
					resource.TestCheckResourceAttr("rockset_query_lambda.test", "name", v1.Name),
					resource.TestCheckResourceAttr("rockset_query_lambda.test", "description", v1.Description),
					testAccCheckSql(t, &queryLambda, v1.SQL),
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

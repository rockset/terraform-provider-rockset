package rockset

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/rockset/rockset-go-client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccScheduledLambda_Basic(t *testing.T) {
	scheduledLambda := "rockset_scheduled_lambda.test_scheduled_lambda"

	type cfg struct {
		CronString          string
		TotalTimesToExecute int64
	}
	s1 := cfg{"0 0 0 ? * * *", 1}
	s2 := cfg{"0 0 0 ? * * *", 2}
	s3 := cfg{"0 0 * ? * * *", 3}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRocksetScheduledLambdaDestroy,
		Steps: []resource.TestStep{
			{
				Config: getHCLTemplate("scheduled_lambda_basic.tf", s1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(scheduledLambda, "rrn"),
					resource.TestCheckResourceAttrSet(scheduledLambda, "apikey"),
					resource.TestCheckResourceAttr(scheduledLambda, "workspace", "acc"),
					resource.TestCheckResourceAttr(scheduledLambda, "cron_string", s1.CronString),
					resource.TestCheckResourceAttr(scheduledLambda, "ql_name", "test_query_lambda_name"),
					resource.TestCheckResourceAttr(scheduledLambda, "tag", "latest"),
					resource.TestCheckResourceAttr(scheduledLambda, "total_times_to_execute", strconv.FormatInt(s1.TotalTimesToExecute, 10)),
				),
			},
			{
				Config: getHCLTemplate("scheduled_lambda_basic.tf", s2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(scheduledLambda, "rrn"),
					resource.TestCheckResourceAttrSet(scheduledLambda, "apikey"),
					resource.TestCheckResourceAttr(scheduledLambda, "workspace", "acc"),
					resource.TestCheckResourceAttr(scheduledLambda, "cron_string", s2.CronString),
					resource.TestCheckResourceAttr(scheduledLambda, "ql_name", "test_query_lambda_name"),
					resource.TestCheckResourceAttr(scheduledLambda, "tag", "latest"),
					resource.TestCheckResourceAttr(scheduledLambda, "total_times_to_execute", strconv.FormatInt(s2.TotalTimesToExecute, 10)),
				),
			},
			{
				Config: getHCLTemplate("scheduled_lambda_basic.tf", s3),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(scheduledLambda, "rrn"),
					resource.TestCheckResourceAttrSet(scheduledLambda, "apikey"),
					resource.TestCheckResourceAttr(scheduledLambda, "workspace", "acc"),
					resource.TestCheckResourceAttr(scheduledLambda, "cron_string", s3.CronString),
					resource.TestCheckResourceAttr(scheduledLambda, "ql_name", "test_query_lambda_name"),
					resource.TestCheckResourceAttr(scheduledLambda, "tag", "latest"),
					resource.TestCheckResourceAttr(scheduledLambda, "total_times_to_execute", strconv.FormatInt(s3.TotalTimesToExecute, 10)),
				),
			},
		},
	})
}

func testAccCheckRocksetScheduledLambdaDestroy(s *terraform.State) error {
	rc := testAccProvider.Meta().(*rockset.RockClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "rockset_scheduled_lambda" {
			continue
		}

		workspace, scheduledLambdaRRN := workspaceAndNameFromID(rs.Primary.ID)
		_, err := rc.GetScheduledLambda(testCtx, workspace, scheduledLambdaRRN)

		// An error would mean we didn't find the key, we expect an error
		if err == nil {
			// We did not get an error, so we failed to delete the key.
			return fmt.Errorf("scheduled lambda %s still exists", scheduledLambdaRRN)
		}
	}

	return nil
}

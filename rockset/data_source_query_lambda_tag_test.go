package rockset

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccQueryLambdaTag_Data(t *testing.T) {
	resourceName := "data.rockset_query_lambda_tag.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: getHCL("data_rockset_query_lambda_tag.tf"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "workspace", "persistent"),
					resource.TestCheckResourceAttr(resourceName, "name", "events"),
					resource.TestCheckResourceAttr(resourceName, "tag", "test"),
					resource.TestCheckResourceAttr(resourceName, "version", "0eb7783c81ef339e"),
					resource.TestCheckResourceAttr(resourceName, "description", "used for testing"),
					resource.TestCheckResourceAttrSet(resourceName, "sql"),
					resource.TestCheckResourceAttrSet(resourceName, "last_executed"),
				),
			},
		},
	})
}

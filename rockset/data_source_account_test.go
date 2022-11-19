package rockset

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAccount_Basic(t *testing.T) {
	resourceName := "data.rockset_account.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: getHCL("rockset_account.tf"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "account_id", "318212636800"),
					resource.TestCheckResourceAttr(resourceName, "organization", "Rockset Circleci"),
					resource.TestCheckResourceAttr(resourceName, "rockset_user", "arn:aws:iam::318212636800:user/rockset"),
				),
			},
		},
	})
}

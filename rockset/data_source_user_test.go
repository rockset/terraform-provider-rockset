package rockset

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataUser_Basic(t *testing.T) {
	user := "data.rockset_user.pme"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: getHCL("data_rockset_user.tf"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(user, "email", "pme+readonly@rockset.com"),
					resource.TestCheckResourceAttr(user, "first_name", "Martin"),
					resource.TestCheckResourceAttr(user, "last_name", "Englund"),
					resource.TestCheckResourceAttr(user, "state", "NEW"),
					resource.TestCheckResourceAttr(user, "roles.0", "read-only"),
					resource.TestCheckResourceAttr(user, "roles.1", "query-only"),
				),
			},
		},
	})
}

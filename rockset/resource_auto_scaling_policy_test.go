package rockset

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAutoScalingPolicy_Basic(t *testing.T) {
	vi := "rockset_autoscaling_policy.main"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: getHCL("autoscaling_policy_enable.tf"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(vi, "id"),
					resource.TestCheckResourceAttr(vi, "enabled", "true"),
					resource.TestCheckResourceAttr(vi, "min_size", "XSMALL"),
					resource.TestCheckResourceAttr(vi, "max_size", "MEDIUM"),
				),
			},
			{
				Config: getHCL("autoscaling_policy_disable.tf"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(vi, "id"),
					resource.TestCheckResourceAttr(vi, "enabled", "false"),
				),
			},
		},
	})
}

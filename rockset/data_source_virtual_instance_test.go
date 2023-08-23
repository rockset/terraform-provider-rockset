package rockset

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVirtualInstance_Data(t *testing.T) {
	resourceName := "data.rockset_virtual_instance.main"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: getHCL("data_rockset_virtual_instance.tf"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "main"),
					resource.TestCheckResourceAttr(resourceName, "description", "The default VI used for streaming ingest and queries"),
					resource.TestCheckResourceAttr(resourceName, "auto_suspend_seconds", "0"),
					resource.TestCheckResourceAttr(resourceName, "current_size", "SMALL"),
					resource.TestCheckResourceAttr(resourceName, "desired_size", "SMALL"),
					resource.TestCheckResourceAttr(resourceName, "default", "true"),
					resource.TestCheckResourceAttr(resourceName, "state", "ACTIVE"),
					resource.TestCheckResourceAttr(resourceName, "enable_remount_on_resume", "false"),
					resource.TestCheckResourceAttr(resourceName, "mount_refresh_interval_seconds", "0"),
				),
			},
		},
	})
}

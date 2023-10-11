package rockset

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/rockset/rockset-go-client"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVirtualInstance_Basic(t *testing.T) {
	vi := "rockset_virtual_instance.test"
	mount := "rockset_collection_mount.patch"

	type cfg struct {
		Name        string
		Description string
		Size        string
		Remount     bool
	}
	v1 := cfg{"small", "v1 desc", "SMALL", true}
	v2 := cfg{"medium", "v2 desc", "MEDIUM", false}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRocksetVirtualInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: getHCLTemplate("virtual_instance_basic.tf", v1),
				Check: resource.ComposeTestCheckFunc(
					// virtual instance
					resource.TestCheckResourceAttrSet(vi, "id"),
					resource.TestCheckResourceAttrSet(vi, "rrn"),
					resource.TestCheckResourceAttr(vi, "name", v1.Name),
					resource.TestCheckResourceAttr(vi, "description", v1.Description),
					resource.TestCheckResourceAttr(vi, "current_size", v1.Size),
					resource.TestCheckResourceAttr(vi, "desired_size", v1.Size),
					resource.TestCheckResourceAttr(vi, "remount_on_resume", strconv.FormatBool(v1.Remount)),
					resource.TestCheckResourceAttr(vi, "default", "false"),
					resource.TestCheckResourceAttr(vi, "state", "ACTIVE"),
					resource.TestCheckResourceAttr(vi, "auto_suspend_seconds", "900"),
					// mount
					resource.TestCheckResourceAttrSet(mount, "id"),
					resource.TestCheckResourceAttrSet(mount, "rrn"),
					resource.TestCheckResourceAttrSet(mount, "virtual_instance_id"),
					resource.TestCheckResourceAttrSet(mount, "virtual_instance_rrn"),
					resource.TestCheckResourceAttrSet(mount, "created_at"),
					//resource.TestCheckResourceAttrSet(mount, "last_refresh_time"),
					resource.TestCheckResourceAttrSet(mount, "snapshot_expiration_time"),
					resource.TestCheckResourceAttr(mount, "state", "ACTIVE"),
					resource.TestCheckResourceAttr(mount, "path", "persistent.patch"),
				),
			},
			{
				Config: getHCLTemplate("virtual_instance_basic.tf", v2),
				Check: resource.ComposeTestCheckFunc(
					// virtual instance
					resource.TestCheckResourceAttr(vi, "name", v2.Name),
					resource.TestCheckResourceAttr(vi, "description", v2.Description),
					resource.TestCheckResourceAttr(vi, "current_size", v2.Size),
					resource.TestCheckResourceAttr(vi, "desired_size", v2.Size),
					resource.TestCheckResourceAttr(vi, "remount_on_resume", strconv.FormatBool(v2.Remount)),
					resource.TestCheckResourceAttr(vi, "default", "false"),
					resource.TestCheckResourceAttr(vi, "state", "ACTIVE"),
					// mount
					resource.TestCheckResourceAttr(mount, "state", "ACTIVE"),
				),
			},
		},
	})
}

func testAccCheckRocksetVirtualInstanceDestroy(s *terraform.State) error {
	rc := testAccProvider.Meta().(*rockset.RockClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "rockset_virtual_instance" {
			continue
		}

		id := rs.Primary.ID
		vi, err := rc.GetVirtualInstance(testCtx, id)

		// An error would mean we didn't find the key, we expect an error
		if err == nil {
			// We did not get an error, so we failed to delete the key.
			return fmt.Errorf("virtual instance %s (%s) still exists", vi.GetName(), id)
		}
	}

	return nil
}

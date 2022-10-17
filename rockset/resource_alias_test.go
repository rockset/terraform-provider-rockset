package rockset

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/rockset/rockset-go-client"
	"github.com/rockset/rockset-go-client/openapi"
)

func TestAccAlias_Basic(t *testing.T) {
	var alias openapi.Alias

	type values struct {
		Name        string
		Description string
		Alias       string
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccRemoveAlias(t, "acc", "name") },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRocksetAliasDestroy,
		Steps: []resource.TestStep{
			{
				Config: getHCLTemplate("alias_basic.tf", values{"name", "description", "commons._events"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRocksetAliasExists("rockset_alias.test", &alias),
					resource.TestCheckResourceAttr("rockset_alias.test", "name", "name"),
					resource.TestCheckResourceAttr("rockset_alias.test", "description", "description"),
					resource.TestCheckResourceAttr("rockset_alias.test", "workspace", "acc"),
					testAccAliasCollectionListMatches(&alias, []string{"commons._events"}),
				),
				ExpectNonEmptyPlan: false,
			},
			{
				Config: getHCLTemplate("alias_basic.tf", values{"name", "updated description", "commons._events"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRocksetAliasExists("rockset_alias.test", &alias),
					resource.TestCheckResourceAttr("rockset_alias.test", "name", "name"),
					resource.TestCheckResourceAttr("rockset_alias.test", "description", "updated description"),
					resource.TestCheckResourceAttr("rockset_alias.test", "workspace", "acc"),
					testAccAliasCollectionListMatches(&alias, []string{"commons._events"}),
				),
				ExpectNonEmptyPlan: false,
			},
			{
				Config: getHCLTemplate("alias_basic.tf", values{"name", "updated description", "commons.test-alias"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRocksetAliasExists("rockset_alias.test", &alias),
					resource.TestCheckResourceAttr("rockset_alias.test", "name", "name"),
					resource.TestCheckResourceAttr("rockset_alias.test", "description", "updated description"),
					resource.TestCheckResourceAttr("rockset_alias.test", "workspace", "acc"),
					testAccAliasCollectionListMatches(&alias, []string{"commons.test-alias"}),
				),
				ExpectNonEmptyPlan: false,
			},
			{
				Config: getHCLTemplate("alias_basic.tf", values{"name", "description", "commons._events"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRocksetAliasExists("rockset_alias.test", &alias),
					resource.TestCheckResourceAttr("rockset_alias.test", "name", "name"),
					resource.TestCheckResourceAttr("rockset_alias.test", "description", "description"),
					resource.TestCheckResourceAttr("rockset_alias.test", "workspace", "acc"),
					testAccAliasCollectionListMatches(&alias, []string{"commons._events"}),
				),
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// clean up any lingering test alias from a previous run
func testAccRemoveAlias(t *testing.T, workspace, alias string) {
	rc := testAccProvider.Meta().(*rockset.RockClient)

	err := rc.DeleteAlias(context.TODO(), workspace, alias)
	if err != nil {
		t.Logf("could not delete alias %s.%s: %v", workspace, alias, err)
	}
}

func testAccCheckRocksetAliasDestroy(s *terraform.State) error {
	rc := testAccProvider.Meta().(*rockset.RockClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "rockset_alias" {
			continue
		}

		workspace, name := workspaceAndNameFromID(rs.Primary.ID)

		_, err := rc.GetAlias(testCtx, name, workspace)
		// A 404 would return an error. We expect a 404 here.
		// Getting a 200 means we failed to delete, so terraform destroy failed.
		if err == nil {
			// We did not get a 404, delete must have failed.
			return fmt.Errorf("alias %s:%s still exists", name, workspace)
		}

		var re rockset.Error
		if errors.As(err, &re) {
			if re.IsNotFoundError() {
				// this is what we expect
				continue
			}
		}
		return err
	}

	return nil
}

func testAccCheckRocksetAliasExists(resource string, alias *openapi.Alias) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, err := getResourceFromState(state, resource)
		if err != nil {
			return err
		}
		workspace, name := workspaceAndNameFromID(rs.Primary.ID)

		rc := testAccProvider.Meta().(*rockset.RockClient)

		resp, err := rc.GetAlias(testCtx, workspace, name)
		if err != nil {
			return fmt.Errorf("failed to get alias %s:%s", workspace, name)
		}

		*alias = resp

		return nil
	}
}

func testAccAliasCollectionListMatches(alias *openapi.Alias, expectedCollections []string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		if !reflect.DeepEqual(alias.GetCollections(), expectedCollections) {
			return fmt.Errorf("expected %s collections, got %s", expectedCollections, alias.GetCollections())
		}

		return nil
	}
}

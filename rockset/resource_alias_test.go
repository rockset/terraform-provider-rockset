package rockset

import (
	"context"
	"errors"
	"fmt"
	rockerr "github.com/rockset/rockset-go-client/errors"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/rockset/rockset-go-client"
	"github.com/rockset/rockset-go-client/openapi"
)

func TestAccAlias_Basic(t *testing.T) {
	var alias openapi.Alias

	name := randomName("alias")
	a1 := Values{
		Name:        name,
		Alias:       "commons._events",
		Workspace:   randomName("ws"),
		Description: description(),
	}
	a2 := a1
	a2.Description = a1.Description + " update"
	a3 := a2
	a3.Alias = "persistent.snp"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) }, //; testAccRemoveAlias(t, "acc", "name") },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRocksetAliasDestroy,
		Steps: []resource.TestStep{
			{
				Config: getHCLTemplate("alias_basic.tf", a1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRocksetAliasExists("rockset_alias.test", &alias),
					resource.TestCheckResourceAttr("rockset_alias.test", "name", a1.Name),
					resource.TestCheckResourceAttr("rockset_alias.test", "description", a1.Description),
					resource.TestCheckResourceAttr("rockset_alias.test", "workspace", a1.Workspace),
					testAccAliasCollectionListMatches(&alias, []string{a1.Alias}),
				),
				ExpectNonEmptyPlan: false,
			},
			{ // change description
				Config: getHCLTemplate("alias_basic.tf", a2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRocksetAliasExists("rockset_alias.test", &alias),
					resource.TestCheckResourceAttr("rockset_alias.test", "name", a2.Name),
					resource.TestCheckResourceAttr("rockset_alias.test", "description", a2.Description),
					resource.TestCheckResourceAttr("rockset_alias.test", "workspace", a2.Workspace),
					testAccAliasCollectionListMatches(&alias, []string{a2.Alias}),
				),
				ExpectNonEmptyPlan: false,
			},
			{ // change collection
				Config: getHCLTemplate("alias_basic.tf", a3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRocksetAliasExists("rockset_alias.test", &alias),
					resource.TestCheckResourceAttr("rockset_alias.test", "name", a3.Name),
					resource.TestCheckResourceAttr("rockset_alias.test", "description", a3.Description),
					resource.TestCheckResourceAttr("rockset_alias.test", "workspace", a3.Workspace),
					testAccAliasCollectionListMatches(&alias, []string{a3.Alias}),
				),
				ExpectNonEmptyPlan: false,
			},
			{ // back to the beginning
				Config: getHCLTemplate("alias_basic.tf", a1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRocksetAliasExists("rockset_alias.test", &alias),
					resource.TestCheckResourceAttr("rockset_alias.test", "name", a1.Name),
					resource.TestCheckResourceAttr("rockset_alias.test", "description", a1.Description),
					resource.TestCheckResourceAttr("rockset_alias.test", "workspace", a1.Workspace),
					testAccAliasCollectionListMatches(&alias, []string{a1.Alias}),
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

		var re rockerr.Error
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

package rockset

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/rockset/rockset-go-client"
)

const testAliasName = "terraform-provider-acceptance-tests"
const testAliasDescription = "terraform provider acceptance tests"
const testAliasWorkspace = "commons"
const testCollection1 = "commons._events"
const testCollection2 = "commons.test-alias"

func TestAccAlias_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRocksetAliasDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAliasBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRocksetAliasExists("rockset_alias.test"),
					resource.TestCheckResourceAttr("rockset_alias.test", "name", testAliasName),
					resource.TestCheckResourceAttr("rockset_alias.test", "description", testAliasDescription),
					resource.TestCheckResourceAttr("rockset_alias.test", "workspace", testAliasWorkspace),
					testAccCheckResourceListMatches("rockset_alias.test", "collections", []string{testCollection1}),
				),
				ExpectNonEmptyPlan: false,
			},
			{
				Config: testAccCheckAliasUpdateDescription(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRocksetAliasExists("rockset_alias.test"),
					resource.TestCheckResourceAttr("rockset_alias.test", "name", testAliasName),
					resource.TestCheckResourceAttr("rockset_alias.test", "description", fmt.Sprintf("%s-updated", testAliasDescription)),
					resource.TestCheckResourceAttr("rockset_alias.test", "workspace", testAliasWorkspace),
					testAccCheckResourceListMatches("rockset_alias.test", "collections", []string{testCollection1}),
				),
				ExpectNonEmptyPlan: false,
			},
			{
				Config: testAccCheckAliasUpdateCollections(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRocksetAliasExists("rockset_alias.test"),
					resource.TestCheckResourceAttr("rockset_alias.test", "name", testAliasName),
					resource.TestCheckResourceAttr("rockset_alias.test", "description", fmt.Sprintf("%s-updated", testAliasDescription)),
					resource.TestCheckResourceAttr("rockset_alias.test", "workspace", testAliasWorkspace),
					testAccCheckResourceListMatches("rockset_alias.test", "collections", []string{testCollection2}),
				),
				ExpectNonEmptyPlan: false,
			},
			{
				Config: testAccCheckAliasUpdateMultipleFields(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRocksetAliasExists("rockset_alias.test"),
					resource.TestCheckResourceAttr("rockset_alias.test", "name", testAliasName),
					resource.TestCheckResourceAttr("rockset_alias.test", "description", testAliasDescription),
					resource.TestCheckResourceAttr("rockset_alias.test", "workspace", testAliasWorkspace),
					testAccCheckResourceListMatches("rockset_alias.test", "collections", []string{testCollection1}),
				),
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func testAccCheckAliasBasic() string {
	return fmt.Sprintf(`
resource rockset_alias test {
	name        = "%s"
	description	= "%s"
	workspace		= "%s"
	collections = ["%s"]
}
`, testAliasName, testAliasDescription, testAliasWorkspace, testCollection1)
}

func testAccCheckAliasUpdateDescription() string {
	return fmt.Sprintf(`
resource rockset_alias test {
	name        = "%s"
	description	= "%s-updated"
	workspace		= "%s"
	collections = ["%s"]
}
`, testAliasName, testAliasDescription, testAliasWorkspace, testCollection1)
}

func testAccCheckAliasUpdateCollections() string {
	return fmt.Sprintf(`
resource rockset_collection test {
	name = "test-alias"
	workspace = "commons"
}
resource rockset_alias test {
	name        = "%s"
	description	= "%s-updated"
	workspace		= "%s"
	collections = ["%s"] 
}
`, testAliasName, testAliasDescription, testAliasWorkspace, testCollection2)
}

func testAccCheckAliasUpdateMultipleFields() string {
	return fmt.Sprintf(`
resource rockset_alias test {
	name        = "%s"
	description	= "%s"
	workspace		= "%s"
	collections = ["%s"] 
}
`, testAliasName, testAliasDescription, testAliasWorkspace, testCollection1)
}

func testAccCheckAliasUpdateNameForceRecreate() string {
	return fmt.Sprintf(`
resource rockset_alias test {
	name        = "%s-updated"
	description	= "%s"
	workspace		= "%s"
	collections = ["%s.%s"] 
}
`, testAliasName, testAliasDescription, testAliasWorkspace, testAliasWorkspace, testCollection1)
}

func testAccCheckRocksetAliasDestroy(s *terraform.State) error {
	rc := testAccProvider.Meta().(*rockset.RockClient)
	ctx := context.TODO()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "rockset_alias" {
			continue
		}

		workspace, name := workspaceAndNameFromID(rs.Primary.ID)

		q := rc.AliasesApi.GetAlias(ctx, name, workspace)
		_, _, err := q.Execute()

		// A 404 would return an error. We expect a 404 here.
		// Getting a 200 means we failed to delete, so terraform destroy failed.
		if err == nil {
			// We did not get a 404, delete must have failed.
			return fmt.Errorf("Alias %s:%s still exists.", name, workspace)
		}
	}

	return nil
}

func testAccCheckRocksetAliasExists(resource string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, err := getResourceFromState(state, resource)
		if err != nil {
			return err
		}
		workspace, name := workspaceAndNameFromID(rs.Primary.ID)

		rc := testAccProvider.Meta().(*rockset.RockClient)
		ctx := context.TODO()

		q := rc.AliasesApi.GetAlias(ctx, workspace, name)
		_, _, err = q.Execute()
		if err != nil {
			return fmt.Errorf("Failed to get alias %s:%s", workspace, name)
		}

		return nil
	}
}

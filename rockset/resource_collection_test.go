package rockset

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/rockset/rockset-go-client"
	"github.com/rockset/rockset-go-client/openapi"
)

func TestAccCollection_Basic(t *testing.T) {
	var collection openapi.Collection

	values := Values{
		Name:        randomName("collection"),
		Description: description(),
		Workspace:   "acc",
		Retention:   60,
	}
	updated := values
	updated.Retention = 61

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRocksetCollectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: getHCLTemplate("collection_basic.tf", values),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRocksetCollectionExists("rockset_collection.test", &collection),
					resource.TestCheckResourceAttr("rockset_collection.test", "name", values.Name),
					resource.TestCheckResourceAttr("rockset_collection.test", "workspace", values.Workspace),
					resource.TestCheckResourceAttr("rockset_collection.test", "description", values.Description),
					testAccCheckRetentionSecsMatches(&collection, values.Retention),
				),
				ExpectNonEmptyPlan: false,
			},
			{
				Config: getHCLTemplate("collection_basic.tf", updated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRocksetCollectionExists("rockset_collection.test", &collection),
					resource.TestCheckResourceAttr("rockset_collection.test", "name", values.Name),
					resource.TestCheckResourceAttr("rockset_collection.test", "workspace", values.Workspace),
					resource.TestCheckResourceAttr("rockset_collection.test", "description", values.Description),
					testAccCheckRetentionSecsMatches(&collection, updated.Retention),
				),
				ExpectNonEmptyPlan: false,
			},
		},
	})
}
func TestAccCollection_Timeout(t *testing.T) {
	values := Values{
		Name:          randomName("collection"),
		Description:   description(),
		Workspace:     "acc",
		CreateTimeout: "5s",
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRocksetCollectionDestroy,
		Steps: []resource.TestStep{
			{
				Config:      getHCLTemplate("collection_basic.tf", values),
				ExpectError: regexp.MustCompile("Error: context deadline exceeded"),
			},
		},
	})
}

func TestAccCollection_IngestTransformation(t *testing.T) {
	var collection openapi.Collection

	values := Values{
		Name:                 randomName("collection"),
		Description:          description(),
		Workspace:            "acc",
		Retention:            60,
		IngestTransformation: "SELECT LOWER(_input.name) AS lower, * FROM _input",
	}
	updatedValues := values
	updatedValues.Description = "updated description"
	updatedValues.IngestTransformation = "SELECT UPPER(_input.name) AS upper, * FROM _input"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRocksetCollectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: getHCLTemplate("collection_basic.tf", values),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRocksetCollectionExists("rockset_collection.test", &collection),
					resource.TestCheckResourceAttr("rockset_collection.test", "name", values.Name),
					resource.TestCheckResourceAttr("rockset_collection.test", "workspace", values.Workspace),
					resource.TestCheckResourceAttr("rockset_collection.test", "description", values.Description),
					resource.TestCheckResourceAttr("rockset_collection.test", "ingest_transformation", values.IngestTransformation),
					testAccCheckRetentionSecsMatches(&collection, values.Retention),
				),
				ExpectNonEmptyPlan: false,
			},
			{
				Config: getHCLTemplate("collection_basic.tf", updatedValues),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRocksetCollectionExists("rockset_collection.test", &collection),
					resource.TestCheckResourceAttr("rockset_collection.test", "name", updatedValues.Name),
					resource.TestCheckResourceAttr("rockset_collection.test", "workspace", updatedValues.Workspace),
					resource.TestCheckResourceAttr("rockset_collection.test", "description", updatedValues.Description),
					resource.TestCheckResourceAttr("rockset_collection.test", "ingest_transformation", updatedValues.IngestTransformation),
				),
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// TestAccCollection_Deprecated_FieldMappingQuery is used to make sure the field_mapping_query attribute can be used
// until we finally remove it.
func TestAccCollection_Deprecated_FieldMappingQuery(t *testing.T) {
	var collection openapi.Collection

	values := Values{
		Name:                 randomName("collection"),
		Description:          description(),
		Workspace:            "acc",
		Retention:            60,
		IngestTransformation: "SELECT COUNT(*) AS cnt FROM _input",
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRocksetCollectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: getHCLTemplate("collection_basic_fmq.tf", values),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRocksetCollectionExists("rockset_collection.test", &collection),
					resource.TestCheckResourceAttr("rockset_collection.test", "name", values.Name),
					resource.TestCheckResourceAttr("rockset_collection.test", "workspace", values.Workspace),
					resource.TestCheckResourceAttr("rockset_collection.test", "description", values.Description),
					resource.TestCheckResourceAttr("rockset_collection.test", "field_mapping_query", values.IngestTransformation),
					testAccCheckRetentionSecsMatches(&collection, values.Retention),
				),
			},
		},
	})
}

/*
Check if any type of collection was successfully destroyed
*/
func testAccCheckRocksetCollectionDestroy(s *terraform.State) error {
	rc := testAccProvider.Meta().(*rockset.RockClient)

	for _, rs := range s.RootModule().Resources {
		if !strings.Contains(rs.Type, "_collection") {
			continue
		}

		workspace, name := workspaceAndNameFromID(rs.Primary.ID)

		_, err := rc.GetCollection(testCtx, workspace, name)

		// An error would mean we didn't find the key, we expect an error
		if err == nil {
			// We did not get an error, so we failed to delete the key.
			return fmt.Errorf("collection %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckRocksetCollectionExists(resource string, collection *openapi.Collection) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rc := testAccProvider.Meta().(*rockset.RockClient)

		rs, err := getResourceFromState(state, resource)
		if err != nil {
			return err
		}

		workspace, name := workspaceAndNameFromID(rs.Primary.ID)
		resp, err := rc.GetCollection(testCtx, workspace, name)
		if err != nil {
			return err
		}

		*collection = resp

		return nil
	}
}

func testAccCheckRetentionSecsMatches(collection *openapi.Collection, expectedValue int) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		retentionSecs := collection.GetRetentionSecs()
		if retentionSecs != int64(expectedValue) {
			return fmt.Errorf("RetentionSeconds was expected to be %d got %d", expectedValue, retentionSecs)
		}

		return nil
	}
}

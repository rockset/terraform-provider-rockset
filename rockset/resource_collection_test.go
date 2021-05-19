package rockset

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/rockset/rockset-go-client"
	"github.com/rockset/rockset-go-client/openapi"
)

const testCollectionName = "terraform-provider-acceptance-tests-basic"
const testCollectionWorkspace = "commons"
const testCollectionDescription = "Terraform provider acceptance tests."
const testCollectionNameFieldMappings = "terraform-provider-acceptance-tests-fieldmapping"

func TestAccCollection_Basic(t *testing.T) {
	var collection openapi.Collection

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRocksetCollectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCollectionBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRocksetCollectionExists("rockset_collection.test", &collection),
					resource.TestCheckResourceAttr("rockset_collection.test", "name", testCollectionName),
					resource.TestCheckResourceAttr("rockset_collection.test", "workspace", testCollectionWorkspace),
					resource.TestCheckResourceAttr("rockset_collection.test", "description", testCollectionDescription),
				),
				ExpectNonEmptyPlan: false,
			},
			{
				Config: testAccCheckCollectionUpdateForceRecreate(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRocksetCollectionExists("rockset_collection.test", &collection),
					resource.TestCheckResourceAttr("rockset_collection.test", "name", fmt.Sprintf("%s-updated", testCollectionName)),
					resource.TestCheckResourceAttr("rockset_collection.test", "workspace", testCollectionWorkspace),
					resource.TestCheckResourceAttr("rockset_collection.test", "description", testCollectionDescription),
				),
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccCollection_FieldMapping(t *testing.T) {
	var collection openapi.Collection

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRocksetCollectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCollectionFieldMapping(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRocksetCollectionExists("rockset_collection.test", &collection),
					resource.TestCheckResourceAttr("rockset_collection.test", "name", testCollectionNameFieldMappings),
					resource.TestCheckResourceAttr("rockset_collection.test", "workspace", testCollectionWorkspace),
					resource.TestCheckResourceAttr("rockset_collection.test", "description", testCollectionDescription),
					testAccCheckFieldMappingMatches(&collection),
				),
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func testAccCheckCollectionFieldMapping() string {
	return fmt.Sprintf(`
resource rockset_collection test {
	name        = "%s"
	workspace   = "%s"
	description = "%s"
	field_mapping {
		name = "string to float"
		input_fields {
			field_name = "population"
			if_missing = "SKIP"
			is_drop    = false
			param      = "pop"
		}

		output_field {
			field_name = "pop"
			on_error   = "FAIL"
			sql        = "CAST(:pop as int)"
		}
	}
}
`, testCollectionNameFieldMappings, testCollectionWorkspace, testCollectionDescription)
}

func testAccCheckCollectionBasic() string {
	return fmt.Sprintf(`
resource rockset_collection test {
	name        = "%s"
	workspace   = "%s"
	description = "%s"
}
`, testCollectionName, testCollectionWorkspace, testCollectionDescription)
}

func testAccCheckCollectionUpdateForceRecreate() string {
	return fmt.Sprintf(`
resource rockset_collection test {
	name        = "%s-updated"
	workspace   = "%s"
	description = "%s"
}`, testCollectionName, testCollectionWorkspace, testCollectionDescription)
}

func testAccCheckRocksetCollectionDestroy(s *terraform.State) error {
	rc := testAccProvider.Meta().(*rockset.RockClient)
	ctx := context.TODO()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "rockset_collection" {
			continue
		}

		workspace, name := workspaceAndNameFromID(rs.Primary.ID)

		_, err := rc.GetCollection(ctx, workspace, name)

		// An error would mean we didn't find the key, we expect an error
		if err == nil {
			// We did not get an error, so we failed to delete the key.
			return fmt.Errorf("Collection %s still exists.", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckRocksetCollectionExists(resource string, collection *openapi.Collection) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rc := testAccProvider.Meta().(*rockset.RockClient)
		ctx := context.TODO()

		rs, err := getResourceFromState(state, resource)
		if err != nil {
			return err
		}

		workspace, name := workspaceAndNameFromID(rs.Primary.ID)
		resp, err := rc.GetCollection(ctx, workspace, name)
		if err != nil {
			return err
		}

		*collection = resp

		return nil
	}
}

func testAccCheckFieldMappingMatches(collection *openapi.Collection) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		foundFieldMappings := collection.GetFieldMappings()
		numMappings := len(foundFieldMappings)
		if numMappings != 1 {
			return fmt.Errorf("Expected one field mapping, found %d.", numMappings)
		}

		fieldMapping := foundFieldMappings[0]
		if fieldMapping.GetName() != "string to float" {
			return fmt.Errorf("Field mapping name expected 'string to float' got %s", fieldMapping.GetName())
		}

		inputFields := fieldMapping.GetInputFields()
		numFields := len(inputFields)
		if numFields != 1 {
			return fmt.Errorf("Expected one input field, found %d.", numFields)
		}

		inputField := inputFields[0]
		if inputField.GetFieldName() != "population" {
			return fmt.Errorf("Expected first input FieldName to be 'population', found %s.", inputField.GetFieldName())
		}

		if inputField.GetIfMissing() != "SKIP" {
			return fmt.Errorf("Expected first input IfMissing to be 'SKIP', found %s.", inputField.GetIfMissing())
		}

		if inputField.GetIsDrop() != false {
			return fmt.Errorf("Expected first input IsDrop to be false, found %t.", inputField.GetIsDrop())
		}

		if inputField.GetParam() != "pop" {
			return fmt.Errorf("Expected first input Param to be 'pop', found %s.", inputField.GetParam())
		}

		outputField := fieldMapping.GetOutputField()
		if outputField.GetFieldName() != "pop" {
			return fmt.Errorf("Expected output FieldName to be 'pop', found %s.", outputField.GetFieldName())
		}

		outputValue := outputField.GetValue()
		outputSql := outputValue.GetSql()
		if outputSql != "CAST(:pop as int)" {
			return fmt.Errorf("Expected output Value.Sql to be 'CAST(:pop as int)', found %s.", outputSql)
		}

		if outputField.GetOnError() != "FAIL" {
			return fmt.Errorf("Expected output FieldName to be 'FAIL', found %s.", outputField.GetOnError())
		}

		return nil
	}
}

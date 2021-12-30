package rockset

import (
	"fmt"
	"reflect"
	"strings"
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
const testCollectionNameFieldMappingQuery = "terraform-provider-acceptance-tests-fieldmappingquery"
const testCollectionNameClustering = "terraform-provider-acceptance-tests-clustering"

/*
	NOTES:
		clustering_key requires field partioning to be enabled for the org.
			otherwise, a 400 bad request is returned.
*/

func TestAccCollection_Basic(t *testing.T) {
	var collection openapi.Collection

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRocksetCollectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCollectionBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRocksetCollectionExists("rockset_collection.test", &collection),
					resource.TestCheckResourceAttr("rockset_collection.test", "name", testCollectionName),
					resource.TestCheckResourceAttr("rockset_collection.test", "workspace", testCollectionWorkspace),
					resource.TestCheckResourceAttr("rockset_collection.test", "description", testCollectionDescription),
					testAccCheckRetentionSecsMatches(&collection, 60),
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
					testAccCheckRetentionSecsMatches(&collection, 61),
				),
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccCollection_FieldMapping(t *testing.T) {
	var collection openapi.Collection

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRocksetCollectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCollectionFieldMapping(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRocksetCollectionExists("rockset_collection.test", &collection),
					resource.TestCheckResourceAttr("rockset_collection.test", "name", testCollectionNameFieldMappings),
					resource.TestCheckResourceAttr("rockset_collection.test", "workspace", testCollectionWorkspace),
					resource.TestCheckResourceAttr("rockset_collection.test", "description", testCollectionDescription),
					testAccCheckFieldMappingMatches(&collection),
					testAccCheckRetentionSecsMatches(&collection, 65),
				),
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccCollection_FieldMappingQuery(t *testing.T) {
	var collection openapi.Collection

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRocksetCollectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCollectionFieldMappingQuery(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRocksetCollectionExists("rockset_collection.test", &collection),
					resource.TestCheckResourceAttr("rockset_collection.test", "name", testCollectionNameFieldMappingQuery),
					resource.TestCheckResourceAttr("rockset_collection.test", "workspace", testCollectionWorkspace),
					resource.TestCheckResourceAttr("rockset_collection.test", "description", testCollectionDescription),
					resource.TestCheckResourceAttr("rockset_collection.test", "field_mapping_query", "SELECT * FROM _input"),
					testAccCheckRetentionSecsMatches(&collection, 65),
				),
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccCollection_ClusteringKey(t *testing.T) {
	var collection openapi.Collection

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRocksetCollectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCollectionClusteringKeyAuto(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRocksetCollectionExists("rockset_collection.test", &collection),
					resource.TestCheckResourceAttr("rockset_collection.test", "name", testCollectionNameClustering),
					resource.TestCheckResourceAttr("rockset_collection.test", "workspace", testCollectionWorkspace),
					resource.TestCheckResourceAttr("rockset_collection.test", "description", testCollectionDescription),
					testAccCheckClusteringKeyMatches(&collection, "population", "AUTO", []string{}),
					testAccCheckRetentionSecsMatches(&collection, 60),
				),
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func testAccCheckCollectionClusteringKeyAuto() string {
	return fmt.Sprintf(`
resource rockset_collection test {
	name						= "%s"
	workspace				= "%s"
	description			= "%s"
	retention_secs	= 60
	clustering_key {
		field_name 	= "population"
		type 				= "AUTO"
	}
}
`, testCollectionNameClustering, testCollectionWorkspace, testCollectionDescription)
}

func testAccCheckCollectionFieldMapping() string {
	return fmt.Sprintf(`
resource rockset_collection test {
	name        = "%s"
	workspace   = "%s"
	description = "%s"
	retention_secs 	= 65

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

func testAccCheckCollectionFieldMappingQuery() string {
	return fmt.Sprintf(`
resource rockset_collection test {
	name        = "%s"
	workspace   = "%s"
	description = "%s"
	retention_secs 	= 65

	field_mapping_query = "SELECT * FROM _input"
}
`, testCollectionNameFieldMappings, testCollectionWorkspace, testCollectionDescription)
}

func testAccCheckCollectionBasic() string {
	return fmt.Sprintf(`
resource rockset_collection test {
	name        		= "%s"
	workspace   		= "%s"
	description 		= "%s"
	retention_secs 	= 60
}
`, testCollectionName, testCollectionWorkspace, testCollectionDescription)
}

func testAccCheckCollectionUpdateForceRecreate() string {
	return fmt.Sprintf(`
resource rockset_collection test {
	name        = "%s-updated"
	workspace   = "%s"
	description = "%s"
	retention_secs 	= 61
}`, testCollectionName, testCollectionWorkspace, testCollectionDescription)
}

/*
	Check if any type of collection was successfuly destroyed
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
			return fmt.Errorf("Collection %s still exists.", rs.Primary.ID)
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

func testAccCheckRetentionSecsMatches(collection *openapi.Collection, expectedValue int) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		retentionSecs := collection.GetRetentionSecs()
		if retentionSecs != int64(expectedValue) {
			return fmt.Errorf("RetentionSeconds was expected to be %d got %d", expectedValue, retentionSecs)
		}

		return nil
	}
}

func testAccCheckClusteringKeyMatches(collection *openapi.Collection,
	fieldName string, partitionType string, partitionKeys []string) resource.TestCheckFunc {

	return func(state *terraform.State) error {
		// Check just the first partition key, assume one is set
		numKeys := len(collection.GetClusteringKey())
		if numKeys != 1 {
			return fmt.Errorf("Expected 1 clustering key, got %d", numKeys)
		}
		clusteringKey := collection.GetClusteringKey()[0]

		if *clusteringKey.FieldName != fieldName {
			return fmt.Errorf("Expected field name %s got %s", fieldName, *clusteringKey.FieldName)
		}

		if *clusteringKey.Type != partitionType {
			return fmt.Errorf("Expected type %s got %s", partitionType, *clusteringKey.Type)
		}

		if !reflect.DeepEqual(*clusteringKey.Keys, partitionKeys) {
			return fmt.Errorf("Expected keys %s got %s", partitionKeys, *clusteringKey.Keys)
		}

		return nil
	}
}

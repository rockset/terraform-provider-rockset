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

func TestAccCollection_IngestTransformation(t *testing.T) {
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
				Config: getHCLTemplate("collection_basic.tf", values),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRocksetCollectionExists("rockset_collection.test", &collection),
					resource.TestCheckResourceAttr("rockset_collection.test", "name", values.Name),
					resource.TestCheckResourceAttr("rockset_collection.test", "workspace", values.Workspace),
					resource.TestCheckResourceAttr("rockset_collection.test", "description", values.Description),
					resource.TestCheckResourceAttr("rockset_collection.test", "field_mapping_query", values.IngestTransformation),
					resource.TestCheckResourceAttr("rockset_collection.test", "clustering_key.#", "0"),
					testAccCheckRetentionSecsMatches(&collection, values.Retention),
				),
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccCollection_ClusteringKey(t *testing.T) {
	var collection openapi.Collection

	values := Values{
		Name:        randomName("collection"),
		Description: description(),
		Workspace:   "acc",
		Retention:   50,
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRocksetCollectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: getHCLTemplate("collection_clustering_key.tf", values),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRocksetCollectionExists("rockset_collection.test", &collection),
					resource.TestCheckResourceAttr("rockset_collection.test", "name", values.Name),
					resource.TestCheckResourceAttr("rockset_collection.test", "workspace", values.Workspace),
					resource.TestCheckResourceAttr("rockset_collection.test", "description", values.Description),
					testAccCheckClusteringKeyMatches(&collection, "population", "AUTO", []string{}),
					testAccCheckRetentionSecsMatches(&collection, values.Retention),
				),
				ExpectNonEmptyPlan: false,
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

func testAccCheckFieldMappingMatches(collection *openapi.Collection) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		foundFieldMappings := collection.GetFieldMappings()
		numMappings := len(foundFieldMappings)
		if numMappings != 1 {
			return fmt.Errorf("expected one field mapping, found %d", numMappings)
		}

		fieldMapping := foundFieldMappings[0]
		if fieldMapping.GetName() != "string to float" {
			return fmt.Errorf("field mapping name expected 'string to float' got %s", fieldMapping.GetName())
		}

		inputFields := fieldMapping.GetInputFields()
		numFields := len(inputFields)
		if numFields != 1 {
			return fmt.Errorf("expected one input field, found %d", numFields)
		}

		inputField := inputFields[0]
		if inputField.GetFieldName() != "population" {
			return fmt.Errorf("expected first input FieldName to be 'population', found %s", inputField.GetFieldName())
		}

		if inputField.GetIfMissing() != "SKIP" {
			return fmt.Errorf("expected first input IfMissing to be 'SKIP', found %s", inputField.GetIfMissing())
		}

		if inputField.GetIsDrop() != false {
			return fmt.Errorf("expected first input IsDrop to be false, found %t", inputField.GetIsDrop())
		}

		if inputField.GetParam() != "pop" {
			return fmt.Errorf("expected first input Param to be 'pop', found %s", inputField.GetParam())
		}

		outputField := fieldMapping.GetOutputField()
		if outputField.GetFieldName() != "pop" {
			return fmt.Errorf("expected output FieldName to be 'pop', found %s", outputField.GetFieldName())
		}

		outputValue := outputField.GetValue()
		outputSql := outputValue.GetSql()
		if outputSql != "CAST(:pop as int)" {
			return fmt.Errorf("expected output Value.Sql to be 'CAST(:pop as int)', found %s", outputSql)
		}

		if outputField.GetOnError() != "FAIL" {
			return fmt.Errorf("expected output FieldName to be 'FAIL', found %s", outputField.GetOnError())
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
			return fmt.Errorf("expected 1 clustering key, got %d", numKeys)
		}
		clusteringKey := collection.GetClusteringKey()[0]

		if *clusteringKey.FieldName != fieldName {
			return fmt.Errorf("expected field name %s got %s", fieldName, *clusteringKey.FieldName)
		}

		if *clusteringKey.Type != partitionType {
			return fmt.Errorf("expected type %s got %s", partitionType, *clusteringKey.Type)
		}

		if !reflect.DeepEqual(clusteringKey.Keys, partitionKeys) {
			return fmt.Errorf("expected keys %s got %s", partitionKeys, clusteringKey.Keys)
		}

		return nil
	}
}

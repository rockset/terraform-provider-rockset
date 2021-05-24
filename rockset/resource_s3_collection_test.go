package rockset

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/rockset/rockset-go-client/openapi"
)

const testS3CollectionName = "terraform-provider-acceptance-tests-basic-s3"
const testS3CollectionWorkspace = "commons"
const testS3CollectionDescription = "Terraform provider acceptance tests."
const testS3CollectionIntegrationRoleArn = "arn:aws:iam::469279130686:role/terraform-provider-rockset-tests"
const testS3ColletionBucket = "terraform-provider-rockset-tests"
const testS3ColletionPatternCsv = "cities.csv"

func TestAccS3Collection_Basic(t *testing.T) {
	var collection openapi.Collection

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRocksetCollectionDestroy, // Reused from base collection
		Steps: []resource.TestStep{
			{
				Config: testAccCheckS3CollectionBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRocksetCollectionExists("rockset_s3_collection.test", &collection), // Reused from base collection
					resource.TestCheckResourceAttr("rockset_s3_collection.test", "name", testS3CollectionName),
					resource.TestCheckResourceAttr("rockset_s3_collection.test", "workspace", testS3CollectionWorkspace),
					resource.TestCheckResourceAttr("rockset_s3_collection.test", "description", testS3CollectionDescription),
					testAccCheckRetentionSecsMatches(&collection, 3600),
				),
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func testAccCheckS3CollectionBasic() string {
	return fmt.Sprintf(`
resource rockset_s3_integration test {
	name = "%s"
	description = "%s"
	aws_role_arn = "%s"
}
resource rockset_s3_collection test {
	name        			= "%s"
	workspace   			= "%s"
	description 			= "%s"
	retention_secs 		= 3600
	integration_name 	= rockset_s3_integration.test.name
	bucket 						= "%s"
	pattern 					= "%s"
	format						= "csv"
	csv {
    first_line_as_column_names = false
    column_names               = [
      "country",
      "city",
      "population",
      "visited"
		]
		column_types 							 = [
			"STRING",
			"STRING",
			"STRING",
			"STRING"
		]
  }
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
	field_mapping {
    name = "string to bool"
    input_fields {
      field_name = "visited"
      if_missing = "SKIP"
      is_drop    = false
      param      = "visited"
    }

    output_field {
      field_name = "been there"
      on_error   = "SKIP"
      sql        = "CAST(:visited as bool)"
    }
  }
}
`, testS3CollectionName, testS3CollectionDescription, testS3CollectionIntegrationRoleArn,
		testS3CollectionName, testS3CollectionWorkspace, testS3CollectionDescription,
		testS3ColletionBucket, testS3ColletionPatternCsv)
}

func testAccCheckS3SourceExpected(collection *openapi.Collection) resource.TestCheckFunc {

	return func(state *terraform.State) error {
		foundFieldMappings := collection.GetFieldMappings()
		numMappings := len(foundFieldMappings)
		if numMappings != 2 {
			return fmt.Errorf("Expected two field mapping, found %d.", numMappings)
		}

		fieldMapping1 := foundFieldMappings[0]
		fieldMapping2 := foundFieldMappings[1]
		if fieldMapping1.GetName() != "string to float" {
			return fmt.Errorf("Field mapping name expected 'string to float' got %s", fieldMapping1.GetName())
		}
		if fieldMapping2.GetName() != "string to bool" {
			return fmt.Errorf("Field mapping name expected 'string to bool' got %s", fieldMapping2.GetName())
		}

		inputFields1 := fieldMapping1.GetInputFields()
		inputFields2 := fieldMapping2.GetInputFields()
		numFields1 := len(inputFields1)
		if numFields1 != 1 {
			return fmt.Errorf("Expected one input field on first field mapping, found %d.", numFields1)
		}
		numFields2 := len(inputFields2)
		if numFields1 != 1 {
			return fmt.Errorf("Expected one input field on second field mapping, found %d.", numFields2)
		}

		inputField1 := inputFields1[0]
		if inputField1.GetFieldName() != "population" {
			return fmt.Errorf("Expected first input FieldName to be 'population', found %s.", inputField1.GetFieldName())
		}
		inputField2 := inputFields2[0]
		if inputField2.GetFieldName() != "visited" {
			return fmt.Errorf("Expected second input FieldName to be 'visited', found %s.", inputField2.GetFieldName())
		}

		if inputField1.GetIfMissing() != "SKIP" {
			return fmt.Errorf("Expected first input IfMissing to be 'SKIP', found %s.", inputField1.GetIfMissing())
		}
		if inputField2.GetIfMissing() != "SKIP" {
			return fmt.Errorf("Expected second input IfMissing to be 'SKIP', found %s.", inputField2.GetIfMissing())
		}

		if inputField1.GetIsDrop() != false {
			return fmt.Errorf("Expected first input IsDrop to be false, found %t.", inputField1.GetIsDrop())
		}
		if inputField2.GetIsDrop() != false {
			return fmt.Errorf("Expected second input IsDrop to be false, found %t.", inputField2.GetIsDrop())
		}

		if inputField1.GetParam() != "pop" {
			return fmt.Errorf("Expected first input Param to be 'pop', found %s.", inputField1.GetParam())
		}

		if inputField2.GetParam() != "visited" {
			return fmt.Errorf("Expected second input Param to be 'visited', found %s.", inputField2.GetParam())
		}

		outputField1 := fieldMapping1.GetOutputField()
		if outputField1.GetFieldName() != "pop" {
			return fmt.Errorf("Expected first output FieldName to be 'pop', found %s.", outputField1.GetFieldName())
		}
		outputField2 := fieldMapping2.GetOutputField()
		if outputField1.GetFieldName() != "been there" {
			return fmt.Errorf("Expected second output FieldName to be 'been there', found %s.", outputField2.GetFieldName())
		}

		outputValue1 := outputField1.GetValue()
		outputSql1 := outputValue1.GetSql()
		if outputSql1 != "CAST(:pop as int)" {
			return fmt.Errorf("Expected first output Value.Sql to be 'CAST(:pop as int)', found %s.", outputSql1)
		}

		outputValue2 := outputField2.GetValue()
		outputSql2 := outputValue2.GetSql()
		if outputSql2 != "CAST(:visited as bool)" {
			return fmt.Errorf("Expected second output Value.Sql to be 'CAST(:visited as bool)', found %s.", outputSql2)
		}

		if outputField1.GetOnError() != "FAIL" {
			return fmt.Errorf("Expected first output FieldName to be 'FAIL', found %s.", outputField1.GetOnError())
		}

		if outputField2.GetOnError() != "SKIP" {
			return fmt.Errorf("Expected second output FieldName to be 'SKIP', found %s.", outputField2.GetOnError())
		}

		return nil
	}
}

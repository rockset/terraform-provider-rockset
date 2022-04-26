package rockset

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/rockset/rockset-go-client/openapi"
	"github.com/stretchr/testify/assert"
)

const testS3CollectionName = "terraform-provider-acceptance-tests-s3"
const testS3CollectionNameJson = "terraform-provider-acceptance-tests-s3-json"
const testS3CollectionWorkspace = "commons"
const testS3CollectionDescription = "Terraform provider acceptance tests."

func TestAccS3Collection_Basic(t *testing.T) {
	var collection openapi.Collection

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRocksetCollectionDestroy, // Reused from base collection
		Steps: []resource.TestStep{
			{
				Config: getHCL("s3_collection.tf"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRocksetCollectionExists("rockset_s3_collection.test", &collection), // Reused from base collection
					resource.TestCheckResourceAttr("rockset_s3_collection.test", "name", testS3CollectionName),
					resource.TestCheckResourceAttr("rockset_s3_collection.test", "workspace", testS3CollectionWorkspace),
					resource.TestCheckResourceAttr("rockset_s3_collection.test", "description", testS3CollectionDescription),
					testAccCheckRetentionSecsMatches(&collection, 3600),
					testAccCheckS3SourceExpected(t, &collection),
				),
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccS3Collection_Json(t *testing.T) {
	var collection openapi.Collection

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRocksetCollectionDestroy, // Reused from base collection
		Steps: []resource.TestStep{
			{
				Config: getHCL("s3_collection_json.tf"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRocksetCollectionExists("rockset_s3_collection.test", &collection), // Reused from base collection
					resource.TestCheckResourceAttr("rockset_s3_collection.test", "name", testS3CollectionNameJson),
					resource.TestCheckResourceAttr("rockset_s3_collection.test", "workspace", testS3CollectionWorkspace),
					resource.TestCheckResourceAttr("rockset_s3_collection.test", "description", testS3CollectionDescription),
					testAccCheckRetentionSecsMatches(&collection, 3600),
				),
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func testAccCheckS3SourceExpected(t *testing.T, collection *openapi.Collection) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		foundFieldMappings := collection.GetFieldMappings()
		numMappings := len(foundFieldMappings)

		assert.Equal(t, numMappings, 2, "Expected two field mappings.")

		fieldMapping1 := foundFieldMappings[0]
		fieldMapping2 := foundFieldMappings[1]

		assert.Equal(t, fieldMapping1.GetName(), "string to float", "First field mapping name didn't match.")
		assert.Equal(t, fieldMapping2.GetName(), "string to bool", "Second field mapping name didn't match.")

		inputFields1 := fieldMapping1.GetInputFields()
		inputFields2 := fieldMapping2.GetInputFields()

		assert.Equal(t, len(inputFields1), 1, "Expected one input field on first field mapping.")
		assert.Equal(t, len(inputFields2), 1, "Expected one input field on second field mapping.")

		inputField1 := inputFields1[0]
		inputField2 := inputFields2[0]
		assert.Equal(t, inputField1.GetFieldName(), "population", "First input FieldName didn't match.")
		assert.Equal(t, inputField1.GetIfMissing(), "SKIP", "First input IfMissing didn't match.")
		assert.False(t, inputField1.GetIsDrop(), "Expected first input IsDrop to be false.")
		assert.Equal(t, inputField1.GetParam(), "pop", "First input Param didn't match.")

		assert.Equal(t, inputField2.GetFieldName(), "visited", "Second input FieldName didn't match.")
		assert.Equal(t, inputField2.GetIfMissing(), "SKIP", "Second input IfMissing didn't match.")
		assert.False(t, inputField2.GetIsDrop(), "Expected second input IsDrop to be false.")
		assert.Equal(t, inputField2.GetParam(), "visited", "Second input Param didn't match.")

		outputField1 := fieldMapping1.GetOutputField()
		outputField2 := fieldMapping2.GetOutputField()

		assert.Equal(t, outputField1.GetFieldName(), "pop", "First output FieldName didn't match.")
		assert.Equal(t, outputField2.GetFieldName(), "been there", "Second output FieldName didn't match.")

		outputValue1 := outputField1.GetValue()
		outputValue2 := outputField2.GetValue()

		assert.Equal(t, outputValue1.GetSql(), "CAST(:pop as int)", "First output Value.Sql didn't match.")
		assert.Equal(t, outputValue2.GetSql(), "CAST(:visited as bool)", "Second output Value.Sql didn't match.")

		assert.Equal(t, outputField1.GetOnError(), "FAIL", "First output OnError didn't match.")
		assert.Equal(t, outputField2.GetOnError(), "SKIP", "Second output OnError didn't match.")

		sources := collection.GetSources()
		assert.Equal(t, 2, len(sources))

		assert.NotNil(t, sources[0].S3)
		assert.NotNil(t, sources[1].S3)

		var xmlIndex, csvIndex int
		if sources[0].FormatParams.Csv == nil && sources[1].FormatParams.Xml == nil {
			xmlIndex = 0
			csvIndex = 1
		} else if sources[1].FormatParams.Csv == nil && sources[0].FormatParams.Xml == nil {
			xmlIndex = 1
			csvIndex = 0
		} else {
			return fmt.Errorf("expected one CSV source and one XML source")
		}

		// Confirm our two sources are what we think.
		assert.NotNil(t, sources[xmlIndex].FormatParams.Xml)
		assert.NotNil(t, sources[csvIndex].FormatParams.Csv)

		// CSV fields
		assert.False(t, sources[csvIndex].FormatParams.Csv.GetFirstLineAsColumnNames())

		assert.Equal(t, "cities.csv", sources[csvIndex].S3.GetPattern())

		assert.Equal(t, 4, len(sources[csvIndex].FormatParams.Csv.GetColumnNames()))
		assert.Equal(t, "country", sources[csvIndex].FormatParams.Csv.GetColumnNames()[0])
		assert.Equal(t, "city", sources[csvIndex].FormatParams.Csv.GetColumnNames()[1])
		assert.Equal(t, "population", sources[csvIndex].FormatParams.Csv.GetColumnNames()[2])
		assert.Equal(t, "visited", sources[csvIndex].FormatParams.Csv.GetColumnNames()[3])

		assert.Equal(t, 4, len(sources[csvIndex].FormatParams.Csv.GetColumnTypes()))
		assert.Equal(t, "STRING", sources[csvIndex].FormatParams.Csv.GetColumnTypes()[0])
		assert.Equal(t, "STRING", sources[csvIndex].FormatParams.Csv.GetColumnTypes()[1])
		assert.Equal(t, "STRING", sources[csvIndex].FormatParams.Csv.GetColumnTypes()[2])
		assert.Equal(t, "STRING", sources[csvIndex].FormatParams.Csv.GetColumnTypes()[3])

		// XML fields
		assert.Equal(t, "cities.xml", sources[xmlIndex].S3.GetPattern())
		assert.Equal(t, "cities", sources[xmlIndex].FormatParams.Xml.GetRootTag())
		assert.Equal(t, "city", sources[xmlIndex].FormatParams.Xml.GetDocTag())
		assert.Equal(t, "UTF-8", sources[xmlIndex].FormatParams.Xml.GetEncoding())

		return nil
	}
}

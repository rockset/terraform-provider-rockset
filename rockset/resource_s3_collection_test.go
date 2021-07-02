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

func testAccCheckS3SourceExpected(t *testing.T, collection *openapi.Collection) resource.TestCheckFunc {
	assert := assert.New(t)

	return func(state *terraform.State) error {
		foundFieldMappings := collection.GetFieldMappings()
		numMappings := len(foundFieldMappings)

		assert.Equal(numMappings, 2, "Expected two field mappings.")

		fieldMapping1 := foundFieldMappings[0]
		fieldMapping2 := foundFieldMappings[1]

		assert.Equal(fieldMapping1.GetName(), "string to float", "First field mapping name didn't match.")
		assert.Equal(fieldMapping2.GetName(), "string to bool", "Second field mapping name didn't match.")

		inputFields1 := fieldMapping1.GetInputFields()
		inputFields2 := fieldMapping2.GetInputFields()

		assert.Equal(len(inputFields1), 1, "Expected one input field on first field mapping.")
		assert.Equal(len(inputFields2), 1, "Expected one input field on second field mapping.")

		inputField1 := inputFields1[0]
		inputField2 := inputFields2[0]
		assert.Equal(inputField1.GetFieldName(), "population", "First input FieldName didn't match.")
		assert.Equal(inputField1.GetIfMissing(), "SKIP", "First input IfMissing didn't match.")
		assert.False(inputField1.GetIsDrop(), "Expected first input IsDrop to be false.")
		assert.Equal(inputField1.GetParam(), "pop", "First input Param didn't match.")

		assert.Equal(inputField2.GetFieldName(), "visited", "Second input FieldName didn't match.")
		assert.Equal(inputField2.GetIfMissing(), "SKIP", "Second input IfMissing didn't match.")
		assert.False(inputField2.GetIsDrop(), "Expected second input IsDrop to be false.")
		assert.Equal(inputField2.GetParam(), "visited", "Second input Param didn't match.")

		outputField1 := fieldMapping1.GetOutputField()
		outputField2 := fieldMapping2.GetOutputField()

		assert.Equal(outputField1.GetFieldName(), "pop", "First output FieldName didn't match.")
		assert.Equal(outputField2.GetFieldName(), "been there", "Second output FieldName didn't match.")

		outputValue1 := outputField1.GetValue()
		outputValue2 := outputField2.GetValue()

		assert.Equal(outputValue1.GetSql(), "CAST(:pop as int)", "First output Value.Sql didn't match.")
		assert.Equal(outputValue2.GetSql(), "CAST(:visited as bool)", "Second output Value.Sql didn't match.")

		assert.Equal(outputField1.GetOnError(), "FAIL", "First output OnError didn't match.")
		assert.Equal(outputField2.GetOnError(), "SKIP", "Second output OnError didn't match.")

		sources := collection.GetSources()
		assert.Equal(2, len(sources))

		assert.NotNil(sources[0].S3)
		assert.NotNil(sources[1].S3)

		var xmlIndex, csvIndex int
		if sources[0].FormatParams.Csv == nil && sources[1].FormatParams.Xml == nil {
			xmlIndex = 0
			csvIndex = 1
		} else if sources[1].FormatParams.Csv == nil && sources[0].FormatParams.Xml == nil {
			xmlIndex = 1
			csvIndex = 0
		} else {
			return fmt.Errorf("Expected one CSV source and one XML source.")
		}

		// Confirm our two sources are what we think.
		assert.NotNil(sources[xmlIndex].FormatParams.Xml)
		assert.NotNil(sources[csvIndex].FormatParams.Csv)

		// CSV fields
		assert.False(sources[csvIndex].FormatParams.Csv.GetFirstLineAsColumnNames())

		assert.Equal("cities.csv", sources[csvIndex].S3.GetPattern())

		assert.Equal(4, len(sources[csvIndex].FormatParams.Csv.GetColumnNames()))
		assert.Equal("country", sources[csvIndex].FormatParams.Csv.GetColumnNames()[0])
		assert.Equal("city", sources[csvIndex].FormatParams.Csv.GetColumnNames()[1])
		assert.Equal("population", sources[csvIndex].FormatParams.Csv.GetColumnNames()[2])
		assert.Equal("visited", sources[csvIndex].FormatParams.Csv.GetColumnNames()[3])

		assert.Equal(4, len(sources[csvIndex].FormatParams.Csv.GetColumnTypes()))
		assert.Equal("STRING", sources[csvIndex].FormatParams.Csv.GetColumnTypes()[0])
		assert.Equal("STRING", sources[csvIndex].FormatParams.Csv.GetColumnTypes()[1])
		assert.Equal("STRING", sources[csvIndex].FormatParams.Csv.GetColumnTypes()[2])
		assert.Equal("STRING", sources[csvIndex].FormatParams.Csv.GetColumnTypes()[3])

		// XML fields
		assert.Equal("cities.xml", sources[xmlIndex].S3.GetPattern())
		assert.Equal("cities", sources[xmlIndex].FormatParams.Xml.GetRootTag())
		assert.Equal("city", sources[xmlIndex].FormatParams.Xml.GetDocTag())
		assert.Equal("UTF-8", sources[xmlIndex].FormatParams.Xml.GetEncoding())

		return nil
	}
}

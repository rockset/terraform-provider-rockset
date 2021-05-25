package rockset

import (
	"log"
	"path/filepath"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/rockset/rockset-go-client/openapi"
	"github.com/stretchr/testify/assert"
)

const testS3CollectionNameCsv = "terraform-provider-acceptance-tests-s3-csv"
const testS3CollectionNameXml = "terraform-provider-acceptance-tests-s3-xml"
const testS3CollectionWorkspace = "commons"
const testS3CollectionDescription = "Terraform provider acceptance tests."
const testS3CollectionIntegrationRoleArn = "arn:aws:iam::469279130686:role/terraform-provider-rockset-tests"
const testS3CollectionBucket = "terraform-provider-rockset-tests"
const testS3ColletionPatternCsv = "cities.csv"

func TestAccS3Collection_Basic(t *testing.T) {
	var collection openapi.Collection

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRocksetCollectionDestroy, // Reused from base collection
		Steps: []resource.TestStep{
			{
				Config: testAccCheckS3CollectionCsv(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRocksetCollectionExists("rockset_s3_collection.test", &collection), // Reused from base collection
					resource.TestCheckResourceAttr("rockset_s3_collection.test", "name", testS3CollectionNameCsv),
					resource.TestCheckResourceAttr("rockset_s3_collection.test", "workspace", testS3CollectionWorkspace),
					resource.TestCheckResourceAttr("rockset_s3_collection.test", "description", testS3CollectionDescription),
					testAccCheckRetentionSecsMatches(&collection, 3600),
					testAccCheckS3SourceExpectedCsv(t, &collection),
				),
				ExpectNonEmptyPlan: false,
			},
			{
				Config: testAccCheckS3CollectionXml(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRocksetCollectionExists("rockset_s3_collection.test", &collection), // Reused from base collection
					resource.TestCheckResourceAttr("rockset_s3_collection.test", "name", testS3CollectionNameXml),
					resource.TestCheckResourceAttr("rockset_s3_collection.test", "workspace", testS3CollectionWorkspace),
					resource.TestCheckResourceAttr("rockset_s3_collection.test", "description", testS3CollectionDescription),
					testAccCheckRetentionSecsMatches(&collection, 3600),
					testAccCheckS3SourceExpectedXml(t, &collection),
				),
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func testAccCheckS3CollectionCsv() string {
	hclPath := filepath.Join("..", "testdata", "s3_collection_csv.tf")
	s3CollectionHCL, err := getFileContents(hclPath)
	if err != nil {
		log.Fatalf("Unexpected error loading test data %s", hclPath)
	}

	return s3CollectionHCL
}

func testAccCheckS3CollectionXml() string {
	hclPath := filepath.Join("..", "testdata", "s3_collection_xml.tf")
	s3CollectionHCL, err := getFileContents(hclPath)
	if err != nil {
		log.Fatalf("Unexpected error loading test data %s", hclPath)
	}

	return s3CollectionHCL
}

func testAccCheckS3SourceExpectedCsv(t *testing.T, collection *openapi.Collection) resource.TestCheckFunc {
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

		return nil
	}
}

func testAccCheckS3SourceExpectedXml(t *testing.T, collection *openapi.Collection) resource.TestCheckFunc {
	assert := assert.New(t)

	return func(state *terraform.State) error {
		assert.Equal(len(collection.GetFieldMappings()), 0, "Expected no field mappings.")

		sources := collection.GetSources()
		assert.Equal(len(sources), 1, "Expected one source.")

		xmlSource := sources[0]
		assert.Equal(*xmlSource.FormatParams.Xml.RootTag, "note")
		assert.Equal(*xmlSource.FormatParams.Xml.Encoding, "UTF-8")
		assert.Equal(*xmlSource.FormatParams.Xml.DocTag, "note")

		return nil
	}
}

package rockset

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/rockset/rockset-go-client"
	"github.com/rockset/rockset-go-client/openapi"
	"github.com/rockset/rockset-go-client/wait"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccS3Collection_Basic(t *testing.T) {
	var collection openapi.Collection

	name := randomName("s3")
	values := Values{
		Name:        name,
		Collection:  name,
		Workspace:   name,
		Description: description(),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRocksetCollectionDestroy, // Reused from base collection
		Steps: []resource.TestStep{
			{
				Config: getHCLTemplate("s3_collection.tf", values),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRocksetCollectionExists("rockset_s3_collection.test", &collection), // Reused from base collection
					resource.TestCheckResourceAttr("rockset_s3_collection.test", "name", values.Collection),
					resource.TestCheckResourceAttr("rockset_s3_collection.test", "workspace", values.Workspace),
					resource.TestCheckResourceAttr("rockset_s3_collection.test", "description", values.Description),
					resource.TestCheckResourceAttr("rockset_s3_collection.test", "retention_secs", "3600"),
				),
			},
		},
	})
}

func TestAccS3Collection_Json(t *testing.T) {
	var collection openapi.Collection

	name := randomName("s3-json")
	values := Values{
		Name:        name,
		Collection:  name,
		Workspace:   name,
		Description: description(),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRocksetCollectionDestroy, // Reused from base collection
		Steps: []resource.TestStep{
			{
				Config: getHCLTemplate("s3_collection_json.tf", values),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRocksetCollectionExists("rockset_s3_collection.test", &collection), // Reused from base collection
					resource.TestCheckResourceAttr("rockset_s3_collection.test", "name", values.Name),
					resource.TestCheckResourceAttr("rockset_s3_collection.test", "workspace", values.Workspace),
					resource.TestCheckResourceAttr("rockset_s3_collection.test", "description", values.Description),
					resource.TestCheckResourceAttr("rockset_s3_collection.test", "retention_secs", "3600"),
					resource.TestCheckResourceAttr("rockset_s3_collection.test", "source.0.integration_name", values.Name),
					resource.TestCheckResourceAttr("rockset_s3_collection.test", "source.0.bucket", "terraform-provider-rockset-tests"),
					resource.TestCheckResourceAttr("rockset_s3_collection.test", "source.0.pattern", "cities.json"),
					resource.TestCheckResourceAttr("rockset_s3_collection.test", "source.0.format", "json"),
				),
			},
			{
				PreConfig: func() {
					triggerWriteAPISourceAdd(t, values.Workspace, values.Collection)
				},
				Config: getHCLTemplate("s3_collection.tf", values),
				Check: resource.ComposeTestCheckFunc(
					// check that we still just have two sources
					resource.TestCheckResourceAttr("rockset_s3_collection.test", "source.#", "2"),
				),
			},
		},
	})
}

func triggerWriteAPISourceAdd(t *testing.T, workspace, collection string) {
	ctx := context.Background()
	rs := testAccProvider.Meta().(*rockset.RockClient)

	doc := map[string]interface{}{"foo": "bar"}

	// write a document to the collection to trigger a write api source to be added
	response, err := rs.AddDocumentsWithOffset(ctx, workspace, collection, []interface{}{doc})
	require.NoError(t, err)
	assert.Equal(t, 1, len(response.GetData()))

	// wait for the document to be available, which also means the write api source has been added
	w := wait.New(rs)
	err = w.UntilQueryable(ctx, workspace, collection, []string{response.GetLastOffset()})
	assert.NoError(t, err)
}

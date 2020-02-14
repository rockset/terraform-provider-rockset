package rockset

import (
	"errors"
	"github.com/davecgh/go-spew/spew"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/rockset/rockset-go-client"
	models "github.com/rockset/rockset-go-client/lib/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"log"
	"testing"
	//models "github.com/rockset/rockset-go-client/lib/go"
)

func TestFoo(t *testing.T) {
	rc, err := rockset.NewClient(rockset.FromEnv())
	assert.Nil(t, err)

	resp, httpResp, err := rc.Integration.Get("test")
	if err != nil {
		var gsw models.GenericSwaggerError
		if errors.As(err, &gsw) {
			log.Printf("msg: %s", gsw.Body())
		}
		log.Printf("http status: %d", httpResp.StatusCode)
		t.Fail()
		return
	}

	spew.Dump(resp.Data)
}

func TestResourceS3IntegrationRead(t *testing.T) {
	t.Skip("not working")
	d := &schema.ResourceData{}
	err := d.Set("name", "foobar")
	require.Nil(t, err)
	rc, err := rockset.NewClient(rockset.FromEnv())
	require.Nil(t, err)

	err = resourceS3IntegrationRead(d, rc)
	assert.Nil(t, err)
}

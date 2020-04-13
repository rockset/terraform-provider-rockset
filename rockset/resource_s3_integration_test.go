package rockset

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/rockset/rockset-go-client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

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

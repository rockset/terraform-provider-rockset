package rockset

import (
	"github.com/rockset/rockset-go-client"
	models "github.com/rockset/rockset-go-client/lib/go"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestWorkspaceID(t *testing.T) {
	a, b := workspaceID("foo:bar")
	assert.Equal(t, "foo", a)
	assert.Equal(t, "bar", b)
}

func TestFlattenFieldMappings(t *testing.T) {
	fm := []models.FieldMappingV2{
		{
			Name: "fm1",
			InputFields: []models.InputField{
				{
					FieldName: "in1",
					IfMissing: "foo",
					IsDrop:    false,
					Param:     "in1",
				},
				{
					FieldName: "in2",
					IfMissing: "bar",
					IsDrop:    true,
					Param:     "in2",
				},
			},
			OutputField: &models.OutputField{
				FieldName: "out1",
				Value:     &models.SqlExpression{Sql: "SELECT * FROM foo"},
				OnError:   "fail",
			},
		},
	}

	flat := flattenFieldMappings(fm)
	assert.Len(t, flat, 1)

	f := flat[0].(map[string]interface{})

	of := makeOutputField(f["output_field"])
	assert.Equal(t, "out1", of.FieldName)

	ifs := makeInputFields(f["input_fields"])
	assert.Len(t, ifs, 2)
}

func TestGetCollection(t *testing.T) {
	rc, err := rockset.NewClient(rockset.FromEnv())
	assert.Nil(t, err)

	_, _, err = rc.Collection.Get("commons", "cities")
	assert.Nil(t, err)

	_, _, err = rc.Collection.Get("commons", "foobar")
	assert.NotNil(t, err)
	e := asSwaggerMessage(err)
	assert.Equal(t, "Collection 'foobar' not found in workspace 'commons'", e.Error())
}

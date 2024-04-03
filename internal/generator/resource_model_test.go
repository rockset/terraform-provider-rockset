package generator_test

import (
	"bytes"
	"testing"

	"github.com/docker/docker/pkg/ioutils"
	"github.com/stretchr/testify/assert"

	"github.com/rockset/terraform-provider-rockset/internal/generator"
)

func TestGenerateResourceModel(t *testing.T) {
	g := generator.New("openapi.Collection")
	g.Package = "generator"
	g.Name = "Collection"
	g.FieldOverrides["Rrn"] = "RRN"

	buf := &bytes.Buffer{}
	ncw := ioutils.NopWriteCloser(buf)
	assert.NoError(t, g.Generate(ncw))
	assert.Equal(t, `name: RRN - tag: , type: string`, buf.String())
}

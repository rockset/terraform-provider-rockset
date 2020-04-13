package rockset

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRocksetNameValidator(t *testing.T) {
	warnings, errors := rocksetNameValidator("s3 integration", "name")
	assert.Len(t, warnings, 0)
	assert.Len(t, errors, 1)
}

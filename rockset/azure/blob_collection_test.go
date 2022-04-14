package azure_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestBlobCollectionSuite(t *testing.T) {
	s := new(BlobCollectionSuite)
	suite.Run(t, s)
}

type BlobCollectionSuite struct {
	suite.Suite
}

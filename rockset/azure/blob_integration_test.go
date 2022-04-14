package azure_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestBlobIntegrationSuite(t *testing.T) {
	s := new(BlobIntegrationSuite)
	suite.Run(t, s)
}

type BlobIntegrationSuite struct {
	suite.Suite
}

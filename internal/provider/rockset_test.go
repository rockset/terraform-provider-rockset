package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"rockset": providerserver.NewProtocol6WithError(New("test")()),
}

func testAccPreCheck(t *testing.T) {
	if os.Getenv("ROCKSET_APIKEY") == "" {
		t.Fatal("ROCKSET_APIKEY must be set for acceptance tests")
	}
	if os.Getenv("ROCKSET_APISERVER") == "" {
		t.Fatal("ROCKSET_APISERVER must be set for acceptance tests")
	}
}

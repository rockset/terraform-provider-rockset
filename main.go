package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"github.com/rockset/terraform-provider-rockset/rockset"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: rockset.Provider,
	})
}

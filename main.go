package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6/tf6server"
	"github.com/hashicorp/terraform-plugin-mux/tf5to6server"
	"github.com/hashicorp/terraform-plugin-mux/tf6muxserver"

	"github.com/rockset/terraform-provider-rockset/internal/provider"
	"github.com/rockset/terraform-provider-rockset/rockset"
)

// setup provider muxing as specified in https://developer.hashicorp.com/terraform/plugin/framework/migrating/mux

var version = "dev"

const FullyQualifiedProviderName = "registry.terraform.io/rockset/rockset"

func main() {
	ctx := context.Background()

	var debug bool
	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	// we're in process of migrating from the SDKv2 to the plugin framework, so we need to mux the two

	// create an instance of the old SDKv2 provider
	upgradedSdkServer, err := tf5to6server.UpgradeServer(ctx, rockset.Provider().GRPCProvider)
	if err != nil {
		log.Fatal(err)
	}

	providers := []func() tfprotov6.ProviderServer{
		// add the new plugin framework provider to the provider list
		providerserver.NewProtocol6(provider.New(version)()),
		// add the old SDKv2 provider to the provider list
		func() tfprotov6.ProviderServer {
			return upgradedSdkServer
		},
	}

	muxServer, err := tf6muxserver.NewMuxServer(ctx, providers...)
	if err != nil {
		log.Fatal(err)
	}

	var serveOpts []tf6server.ServeOpt
	if debug {
		serveOpts = append(serveOpts, tf6server.WithManagedDebug())
	}

	if err = tf6server.Serve(FullyQualifiedProviderName, muxServer.ProviderServer, serveOpts...); err != nil {
		log.Fatal(err)
	}
}

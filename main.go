package main

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

func main() {
	if err := providerserver.Serve(context.Background(), newProvider, providerserver.ServeOpts{
		Address: "registry.terraform.io/singingbuddy/tiktok",
	}); err != nil {
		log.Fatal(err)
	}
}

package main

import (
	"context"
	"log"

	"github.com/eyesofnetwork/terraform-provider-eon/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

var version = "0.1.0"

func main() {
	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/eyesofnetwork/eon",
	}
	if err := providerserver.Serve(context.Background(), provider.New(version), opts); err != nil {
		log.Fatal(err)
	}
}

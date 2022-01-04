package main

import (
	"context"
	"log"
	"os"

	"github.com/massdriver-cloud/terraform-provider-cache/cache"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5/tf5server"
	tfmux "github.com/hashicorp/terraform-plugin-mux"
)

func main() {

	cache := cache.Provider()

	ctx := context.Background()
	factory, err := tfmux.NewSchemaServerFactory(ctx, cache)
	if err != nil {
		log.Println(err.Error())
		os.Exit(1)
	}

	tf5server.Serve("registry.terraform.io/massdriver-cloud/cache", func() tfprotov5.ProviderServer {
		return factory.Server()
	})
}

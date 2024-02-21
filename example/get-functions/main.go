package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/openfaas/faas-provider/types"
	"github.com/openfaas/go-sdk"
)

func main() {
	// NOTE: You can have any name for environment variables. below defined variables names are not standard names
	username := os.Getenv("OPENFAAS_USERNAME")
	password := os.Getenv("OPENFAAS_PASSWORD")

	gatewayURL, _ := url.Parse(os.Getenv("OPENFAAS_GATEWAY_URL"))
	auth := &sdk.BasicAuth{
		Username: username,
		Password: password,
	}

	client := sdk.NewClient(gatewayURL, auth, http.DefaultClient)

	fns, err := client.GetFunctions(context.Background(), "openfaas-fn")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Get Failed: %s", err)
		os.Exit(1)
	}
	fmt.Printf("No Of Functions: %d\n", len(fns))

	status, err := client.Deploy(context.Background(), types.FunctionDeployment{
		Service:    "env-store-test",
		Image:      "ghcr.io/openfaas/alpine:latest",
		Namespace:  "openfaas-fn",
		EnvProcess: "env",
		Labels: &map[string]string{
			"purpose": "test",
		},
	})
	// non 200 status value will have some error
	if err != nil {
		fmt.Fprintf(os.Stderr, "Status: %d Deploy Failed: %s", status, err)
		os.Exit(1)
	}

	fns, err = client.GetFunctions(context.Background(), "openfaas-fn")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Get Failed: %s", err)
		os.Exit(1)
	}
	fmt.Printf("No Of Functions: %d\n", len(fns))

	err = client.DeleteFunction(context.Background(), "env-store-test", "openfaas-fn")
	// non 200 status value will have some error
	if err != nil {
		fmt.Fprintf(os.Stderr, "Delete Failed: %s", err)
		os.Exit(1)
	}
}

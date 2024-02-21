package main

import (
	"context"
	"fmt"
	"log"
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

	ns, err := client.GetNamespaces(context.Background())
	if err != nil {
		log.Printf("Get Failed: %s", err)
	}
	fmt.Printf("No Of Namespaces: %d\n", len(ns))

	status, err := client.CreateNamespace(context.Background(), types.FunctionNamespace{
		Name: "test-namespace",
	})
	// non 200 status value will have some error
	if err != nil {
		log.Printf("Status: %d Create Failed: %s", status, err)
	}

	ns, err = client.GetNamespaces(context.Background())
	if err != nil {
		log.Printf("Get Failed: %s", err)
	}
	fmt.Printf("No Of Namespaces: %d\n", len(ns))

	// delete namespace
	err = client.DeleteNamespace(context.Background(), "test-namespace")
	// non 200 status value will have some error
	if err != nil {
		log.Printf("Delete Failed: %s", err)
	}
}

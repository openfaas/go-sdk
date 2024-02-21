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

	status, err := client.CreateSecret(context.Background(), types.Secret{
		Name:      "env-store-test",
		Namespace: "openfaas-fn",
		// secret support both binary and string values
		// Use Value field to store string values
		RawValue: []byte("this is secret"),
	})
	// non 200 status value will have some error
	if err != nil {
		fmt.Fprintf(os.Stderr, "Status: %d Create Failed: %s", status, err)
		os.Exit(1)
	}

	status, err = client.UpdateSecret(context.Background(), types.Secret{
		Name:      "env-store-test",
		Namespace: "openfaas-fn",
		// secret support both binary and string values
		// Use Value field to store string values
		RawValue: []byte("update secret value"),
	})
	// non 200 status value will have some error
	if err != nil {
		fmt.Fprintf(os.Stderr, "Status: %d Update Failed: %s", status, err)
		os.Exit(1)
	}

	err = client.DeleteSecret(context.Background(), "env-store-test", "openfaas-fn")
	// non 200 status value will have some error
	if err != nil {
		fmt.Fprintf(os.Stderr, "Delete Failed: %s", err)
		os.Exit(1)
	}
}

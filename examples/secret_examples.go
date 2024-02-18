package examples

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

func CreateSecret() {
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
		log.Printf("Status: %d Create Failed: %s", status, err)
	}
}

func UpdateSecret() {
	// NOTE: You can have any name for environment variables. below defined variables names are not standard names
	username := os.Getenv("OPENFAAS_USERNAME")
	password := os.Getenv("OPENFAAS_PASSWORD")

	gatewayURL, _ := url.Parse(os.Getenv("OPENFAAS_GATEWAY_URL"))
	auth := &sdk.BasicAuth{
		Username: username,
		Password: password,
	}

	client := sdk.NewClient(gatewayURL, auth, http.DefaultClient)

	status, err := client.UpdateSecret(context.Background(), types.Secret{
		Name:      "env-store-test",
		Namespace: "openfaas-fn",
		// secret support both binary and string values
		// Use Value field to store string values
		RawValue: []byte("update secret value"),
	})

	// non 200 status value will have some error
	if err != nil {
		log.Printf("Status: %d Update Failed: %s", status, err)
	}
}

func DeleteSecret() {
	// NOTE: You can have any name for environment variables. below defined variables names are not standard names
	username := os.Getenv("OPENFAAS_USERNAME")
	password := os.Getenv("OPENFAAS_PASSWORD")

	gatewayURL, _ := url.Parse(os.Getenv("OPENFAAS_GATEWAY_URL"))
	auth := &sdk.BasicAuth{
		Username: username,
		Password: password,
	}

	client := sdk.NewClient(gatewayURL, auth, http.DefaultClient)

	err := client.DeleteSecret(context.Background(),
		"env-store-test",
		"openfaas-fn")

	// non 200 status value will have some error
	if err != nil {
		log.Printf("Delete Failed: %s", err)
	}
}

func GetSecrets() {
	// NOTE: You can have any name for environment variables. below defined variables names are not standard names
	username := os.Getenv("OPENFAAS_USERNAME")
	password := os.Getenv("OPENFAAS_PASSWORD")

	gatewayURL, _ := url.Parse(os.Getenv("OPENFAAS_GATEWAY_URL"))
	auth := &sdk.BasicAuth{
		Username: username,
		Password: password,
	}

	client := sdk.NewClient(gatewayURL, auth, http.DefaultClient)

	secrets, err := client.GetSecrets(context.Background(), "openfaas-fn")

	// non 200 status value will have some error
	if err != nil {
		log.Printf("Get Failed: %s", err)
	}

	for _, s := range secrets {
		fmt.Printf("Secret: %v \n", s)
	}
}

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

func CreateNamespace() {
	// NOTE: You can have any name for environment variables. below defined variables names are not standard names
	username := os.Getenv("OPENFAAS_USERNAME")
	password := os.Getenv("OPENFAAS_PASSWORD")

	gatewayURL, _ := url.Parse(os.Getenv("OPENFAAS_GATEWAY_URL"))
	auth := &sdk.BasicAuth{
		Username: username,
		Password: password,
	}

	client := sdk.NewClient(gatewayURL, auth, http.DefaultClient)

	status, err := client.CreateNamespace(context.Background(), types.FunctionNamespace{
		Name: "test-namespace",
		Labels: map[string]string{
			"env": "dev",
		},
		Annotations: map[string]string{
			"imageregistry": "https://hub.docker.com/",
		},
	})

	// non 200 status value will have some error
	if err != nil {
		log.Printf("Status: %d Create Failed: %s", status, err)
	}
}

func UpdateNamespace() {
	// NOTE: You can have any name for environment variables. below defined variables names are not standard names
	username := os.Getenv("OPENFAAS_USERNAME")
	password := os.Getenv("OPENFAAS_PASSWORD")

	gatewayURL, _ := url.Parse(os.Getenv("OPENFAAS_GATEWAY_URL"))
	auth := &sdk.BasicAuth{
		Username: username,
		Password: password,
	}

	client := sdk.NewClient(gatewayURL, auth, http.DefaultClient)

	status, err := client.UpdateNamespace(context.Background(), types.FunctionNamespace{
		Name: "test-namespace",
		Labels: map[string]string{
			"env": "dev",
		},
		Annotations: map[string]string{
			"imageregistry": "https://private.registry.com/",
		},
	})

	// non 200 status value will have some error
	if err != nil {
		log.Printf("Status: %d Update Failed: %s", status, err)
	}
}

func GetNamespace() {
	// NOTE: You can have any name for environment variables. below defined variables names are not standard names
	username := os.Getenv("OPENFAAS_USERNAME")
	password := os.Getenv("OPENFAAS_PASSWORD")

	gatewayURL, _ := url.Parse(os.Getenv("OPENFAAS_GATEWAY_URL"))
	auth := &sdk.BasicAuth{
		Username: username,
		Password: password,
	}

	client := sdk.NewClient(gatewayURL, auth, http.DefaultClient)

	ns, err := client.GetNamespace(context.Background(), "test-namespace")
	if err != nil {
		log.Printf("Get Failed: %s", err)
	}

	fmt.Printf("Namespace: %v \n", ns)
}

func GetNamespaces() {
	// NOTE: You can have any name for environment variables. below defined variables names are not standard names
	username := os.Getenv("OPENFAAS_USERNAME")
	password := os.Getenv("OPENFAAS_PASSWORD")

	gatewayURL, _ := url.Parse(os.Getenv("OPENFAAS_GATEWAY_URL"))
	auth := &sdk.BasicAuth{
		Username: username,
		Password: password,
	}

	client := sdk.NewClient(gatewayURL, auth, http.DefaultClient)

	namespaces, err := client.GetNamespaces(context.Background())
	if err != nil {
		log.Printf("Get Failed: %s", err)
	}

	for _, ns := range namespaces {
		fmt.Printf("Namespace: %v \n", ns)
	}
}

func DeleteNamespace() {
	// NOTE: You can have any name for environment variables. below defined variables names are not standard names
	username := os.Getenv("OPENFAAS_USERNAME")
	password := os.Getenv("OPENFAAS_PASSWORD")

	gatewayURL, _ := url.Parse(os.Getenv("OPENFAAS_GATEWAY_URL"))
	auth := &sdk.BasicAuth{
		Username: username,
		Password: password,
	}

	client := sdk.NewClient(gatewayURL, auth, http.DefaultClient)

	err := client.DeleteNamespace(context.Background(), "test-namespace")

	// non 200 status value will have some error
	if err != nil {
		log.Printf("Delete Failed: %s", err)
	}
}

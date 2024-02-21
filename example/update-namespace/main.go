package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

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

	fmt.Println("Wait for 15 seconds....")
	fmt.Println("Get Namespace Before Update")
	time.Sleep(15 * time.Second)
	ns, err := client.GetNamespace(context.Background(), "test-namespace")
	if err != nil {
		log.Printf("Get Failed: %s", err)
	}
	fmt.Printf("Namespace: %v \n", ns)

	status, err = client.UpdateNamespace(context.Background(), types.FunctionNamespace{
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

	fmt.Println("Wait for 15 seconds....")
	fmt.Println("Get Namespace After Update")
	time.Sleep(15 * time.Second)
	ns, err = client.GetNamespace(context.Background(), "test-namespace")
	if err != nil {
		log.Printf("Get Failed: %s", err)
	}
	fmt.Printf("Namespace: %v \n", ns)

	// delete namespace
	err = client.DeleteNamespace(context.Background(), "test-namespace")
	// non 200 status value will have some error
	if err != nil {
		log.Printf("Delete Failed: %s", err)
	}
}

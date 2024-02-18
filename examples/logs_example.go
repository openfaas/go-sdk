package examples

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/openfaas/go-sdk"
)

func GetLogs() {

	// NOTE: You can have any name for environment variables. below defined variables names are not standard names
	username := os.Getenv("OPENFAAS_USERNAME")
	password := os.Getenv("OPENFAAS_PASSWORD")

	gatewayURL, _ := url.Parse(os.Getenv("OPENFAAS_GATEWAY_URL"))
	auth := &sdk.BasicAuth{
		Username: username,
		Password: password,
	}

	client := sdk.NewClient(gatewayURL, auth, http.DefaultClient)

	// Follow is allows the user to request a stream of logs until the timeout
	follow := false
	// Tail sets the maximum number of log messages to return, <=0 means unlimited
	tail := 5
	// Since is the optional datetime value to start the logs from
	since := time.Now().Add(-30 * time.Second)

	logsChan, err := client.GetLogs(context.Background(), "env-store-test", "openfaas-fn", follow, tail, &since)
	if err != nil {
		log.Printf("Get Logs Failed: %s", err)
	}

	fmt.Println("Logs Received....")
	for line := range logsChan {
		fmt.Println(line)
	}

}

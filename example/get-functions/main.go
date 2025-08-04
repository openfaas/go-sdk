package main

import (
	"context"
	b64 "encoding/base64"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	execute "github.com/alexellis/go-execute/pkg/v2"
	"github.com/openfaas/go-sdk"
)

// main will:

// Get all functions, and print the number found.
// Iterate each, and print the various details available.
func main() {
	var (
		gateway,
		username,
		password string
	)

	flag.StringVar(&gateway, "gateway", "http://127.0.0.1:8080", "The URL of the OpenFaaS gateway")
	flag.StringVar(&username, "username", "admin", "The username to use for authentication")
	flag.StringVar(&password, "password", "", "The password to use for authentication")

	flag.Parse()

	if len(password) == 0 {
		fmt.Println("No --password provided, looking up via kubectl")
		password = lookupPasswordViaKubectl()
	}

	gatewayURL, err := url.Parse(gateway)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid gateway URL: %s", err)
		os.Exit(1)
	}

	auth := &sdk.BasicAuth{
		Username: username,
		Password: password,
	}

	client := sdk.NewClient(gatewayURL, auth, http.DefaultClient)

	functions, err := client.GetFunctions(context.Background(), "openfaas-fn")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Get Failed: %s", err)
		os.Exit(1)
	}
	fmt.Printf("Found %d functions\n", len(functions))
	fmt.Println("--------------------------------")

	for _, function := range functions {
		fmt.Printf("Function: %s\n", function.Name)
		fmt.Printf("Image: %s\n", function.Image)
		if function.Labels != nil {
			fmt.Printf("Labels:\n")
			for key, value := range *function.Labels {
				fmt.Printf("  %s: %s\n", key, value)
			}
		} else {
			fmt.Printf("Labels: <none>\n")
		}
		if function.Annotations != nil {
			fmt.Printf("Annotations:\n")
			for key, value := range *function.Annotations {
				fmt.Printf("  %s: %s\n", key, value)
			}
		} else {
			fmt.Printf("Annotations: <none>\n")
		}

		if function.EnvVars != nil {
			fmt.Printf("EnvVars:\n")
			for key, value := range function.EnvVars {
				fmt.Printf("  %s: %s\n", key, value)
			}
		} else {
			fmt.Printf("EnvVars: <none>\n")
		}

		if function.Secrets != nil {
			fmt.Printf("Secrets:\n")
			for _, value := range function.Secrets {
				fmt.Printf("  -%s\n", value)
			}
		} else {
			fmt.Printf("Secrets: <none>\n")
		}

		if function.Requests != nil {
			fmt.Printf("Requests: %+v\n", function.Requests)
		} else {
			fmt.Printf("Requests: <none>\n")
		}

		if function.Limits != nil {
			fmt.Printf("Limits: %+v\n", function.Limits)
		} else {
			fmt.Printf("Limits: <none>\n")
		}
		fmt.Println("--------------------------------")
	}
}

func lookupPasswordViaKubectl() string {

	cmd := execute.ExecTask{
		Command:      "kubectl",
		Args:         []string{"get", "secret", "-n", "openfaas", "basic-auth", "-o", "jsonpath='{.data.basic-auth-password}'"},
		StreamStdio:  false,
		PrintCommand: false,
	}

	res, err := cmd.Execute(context.Background())
	if err != nil {
		panic(err)
	}

	if res.ExitCode != 0 {
		panic("Non-zero exit code: " + res.Stderr)
	}
	resOut := strings.Trim(res.Stdout, "\\'")

	decoded, err := b64.StdEncoding.DecodeString(resOut)
	if err != nil {
		panic(err)
	}

	password := strings.TrimSpace(string(decoded))

	return password
}

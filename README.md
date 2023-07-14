## go-sdk

A lightweight Go SDK for use within OpenFaaS functions and to control the OpenFaaS gateway.

For use within any Go code (not just OpenFaaS Functions):

* Client - A client for the OpenFaaS REST API

For use within functions:

* ReadSecret() - Read a named secret from within an OpenFaaS Function
* ReadSecrets() - Read all available secrets returning a queryable map

## Usage

```go
import "github.com/openfaas/go-sdk"
```

Construct a new OpenFaaS client and use it to access the OpenFaaS gateway API.

```go
gatewayURL, _ := url.Parse("http://127.0.0.1:8080")
auth := &sdk.BasicAuth{
    Username: username,
    Password: password,
}

client := sdk.NewClient(gatewayURL, auth, http.DefaultClient)

namespace, err := client.GetNamespaces(context.Background())
```

### Authentication with IAM

To authenticate with an OpenFaaS deployment that has [Identity and Access Management (IAM)](https://docs.openfaas.com/openfaas-pro/iam/overview/) enabled, the client needs to exchange an ID token for an OpenFaaS ID token.

To get a token that can be exchanged for an OpenFaaS token you need to implement the `TokenSource` interface.

This is an example of a token source that gets a service account token mounted into a pod with [ServiceAccount token volume projection](https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/#serviceaccount-token-volume-projection).

```go
type ServiceAccountTokenSource struct{}

func (ts *ServiceAccountTokenSource) Token() (string, error) {
	tokenMountPath := getEnv("token_mount_path", "/var/secrets/tokens")
	if len(tokenMountPath) == 0 {
		return "", fmt.Errorf("invalid token_mount_path specified for reading the service account token")
	}

	idTokenPath := path.Join(tokenMountPath, "openfaas-token")
	idToken, err := os.ReadFile(idTokenPath)
	if err != nil {
		return "", fmt.Errorf("unable to load service account token: %s", err)
	}

	return string(idToken), nil
}
```

The service account token returned by the `TokenSource` is automatically exchanged for an OpenFaaS token that is then used in the Authorization header for all requests made to the API.

If the OpenFaaS token is expired the `TokenSource` is asked for a token and the token exchange will run again.

```go
gatewayURL, _ := url.Parse("https://gw.openfaas.example.com")

auth := &sdk.TokenAuth{
    TokenURL "https://gw.openfaas.example.com/oauth/token",
    TokenSource: &ServiceAccountTokenSource{}
}

client := sdk.NewClient(gatewayURL, auth, http.DefaultClient)
```

License: MIT

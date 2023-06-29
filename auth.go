package sdk

import (
	"net/http"
	"sync"
)

// BasicAuth basic authentication for the the OpenFaaS client
type BasicAuth struct {
	Username string
	Password string
}

// Set Authorization Basic header on request
func (auth *BasicAuth) Set(req *http.Request) error {
	req.SetBasicAuth(auth.Username, auth.Password)
	return nil
}

// A TokenSource is anything that can return an OIDC ID token that can be exchanged for
// an OpenFaaS token.
type TokenSource interface {
	// Token returns a token or an error.
	Token() (string, error)
}

// TokenAuth bearer token authentication for OpenFaaS deployments with OpenFaaS IAM
// enabled.
type TokenAuth struct {
	// TokenURL represents the OpenFaaS gateways token endpoint URL.
	TokenURL string

	// TokenSource used to get an ID token that can be exchanged for an OpenFaaS ID token.
	TokenSource TokenSource

	lock  sync.Mutex // guards token
	token *Token
}

// Set Authorization Bearer header on request.
// Set validates the token expiry on each call. If it's expired it will exchange
// an ID token from the TokenSource for a new OpenFaaS token.
func (a *TokenAuth) Set(req *http.Request) error {
	a.lock.Lock()
	defer a.lock.Unlock()

	if a.token == nil || a.token.Expired() {
		idToken, err := a.TokenSource.Token()
		if err != nil {
			return err
		}

		token, err := ExchangeIDToken(a.TokenURL, idToken)
		if err != nil {
			return err
		}
		a.token = token
	}

	req.Header.Add("Authorization", "Bearer "+a.token.IDToken)
	return nil
}

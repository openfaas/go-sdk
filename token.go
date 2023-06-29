package sdk

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// expiryDelta determines how much earlier a token should be considered
// expired than its actual expiration time. It is used to avoid late
// expirations due to client-server time mismatches.
const expiryDelta = 10 * time.Second

// Token represents an OpenFaaS ID token
type Token struct {
	// IDToken is the OIDC access token that authorizes and authenticates
	// the requests.
	IDToken string

	// Expiry is the expiration time of the access token.
	//
	// A zero value means the token never expires.
	Expiry time.Time
}

// Expired reports whether the token is expired.
func (t *Token) Expired() bool {
	if t.Expiry.IsZero() {
		return false
	}

	return t.Expiry.Round(0).Add(-expiryDelta).Before(time.Now())
}

type tokenJSON struct {
	AccessToken string `json:"access_token"`
	IDToken     string `json:"id_token"`

	TokenType string `json:"token_type"`

	ExpiresIn int `json:"expires_in"`
}

func (t *tokenJSON) expiry() (exp time.Time) {
	if v := t.ExpiresIn; v != 0 {
		return time.Now().Add(time.Duration(v) * time.Second)
	}
	return
}

// Exchange an ID Token for OpenFaaS token
func ExchangeIDToken(tokenURL, rawIDToken string) (*Token, error) {
	v := url.Values{}
	v.Set("grant_type", "urn:ietf:params:oauth:grant-type:token-exchange")
	v.Set("subject_token_type", "urn:ietf:params:oauth:token-type:id_token")
	v.Set("subject_token", rawIDToken)

	u, _ := url.Parse(tokenURL)

	req, err := http.NewRequest(http.MethodPost, u.String(), strings.NewReader(v.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("cannot fetch token: %v", err)
	}

	if code := res.StatusCode; code < 200 || code > 299 {
		return nil, fmt.Errorf("cannot fetch token: %v\nResponse: %s", res.Status, body)
	}

	tj := &tokenJSON{}
	if err := json.Unmarshal(body, tj); err != nil {
		return nil, fmt.Errorf("unable to unmarshal token: %s", err)
	}

	return &Token{
		IDToken: tj.AccessToken,
		Expiry:  tj.expiry(),
	}, nil
}

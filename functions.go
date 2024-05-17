package sdk

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const DefaultNamespace = "openfaas-fn"

func (c *Client) InvokeFunction(ctx context.Context, name, namespace string, method string, header http.Header, query url.Values, body io.Reader, async bool, auth bool) (*http.Response, error) {
	fnEndpoint := "/function"
	if async {
		fnEndpoint = "/async-function"
	}

	if len(namespace) == 0 {
		namespace = DefaultNamespace
	}

	u, _ := url.Parse(c.GatewayURL.String())
	u.Path = fmt.Sprintf("%s/%s.%s", fnEndpoint, name, namespace)
	u.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, method, u.String(), body)
	if err != nil {
		return nil, err
	}

	for key, values := range header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	if auth && c.FunctionTokenSource != nil {
		tokenURL := fmt.Sprintf("%s/oauth/token", c.GatewayURL.String())
		scope := []string{"function"}
		audience := []string{fmt.Sprintf("%s:%s", namespace, name)}

		idToken, err := c.FunctionTokenSource.Token()
		if err != nil {
			return nil, fmt.Errorf("failed to get function access token: %w", err)
		}

		// Consider caching the token in memory as long as the token is valid
		// to prevent having to do a token exchange each time the function is invoked.
		functionToken, err := ExchangeIDToken(tokenURL, idToken, WithScope(scope), WithAudience(audience))

		var authError *OAuthError
		if errors.As(err, &authError) {
			return nil, fmt.Errorf("failed to get function access token: %s", authError.Description)
		}
		if err != nil {
			return nil, fmt.Errorf("failed to get function access token: %w", err)
		}

		req.Header.Add("Authorization", "Bearer "+functionToken.IDToken)
	}

	return c.do(req)
}

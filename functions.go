package sdk

import (
	"context"
	"crypto/sha256"
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
		idToken, err := c.FunctionTokenSource.Token()
		if err != nil {
			return nil, fmt.Errorf("failed to get function access token: %w", err)
		}

		tokenURL := fmt.Sprintf("%s/oauth/token", c.GatewayURL.String())
		scope := []string{"function"}
		audience := []string{fmt.Sprintf("%s:%s", namespace, name)}

		var bearer string
		if c.fnTokenCache != nil {
			// Function access tokens are cached as long as the token is valid
			// to prevent having to do a token exchange each time the function is invoked.
			cacheKey := getFunctionTokenCacheKey(idToken, fmt.Sprintf("%s.%s", name, namespace))

			token, ok := c.fnTokenCache.Get(cacheKey)
			if !ok {
				token, err = ExchangeIDToken(tokenURL, idToken, WithScope(scope), WithAudience(audience))
				if err != nil {
					return nil, fmt.Errorf("failed to get function access token: %w", err)
				}

				c.fnTokenCache.Set(cacheKey, token)
			}

			bearer = token.IDToken
		} else {
			token, err := ExchangeIDToken(tokenURL, idToken, WithScope(scope), WithAudience(audience))
			if err != nil {
				return nil, fmt.Errorf("failed to get function access token: %w", err)
			}

			bearer = token.IDToken
		}

		req.Header.Add("Authorization", "Bearer "+bearer)
	}

	return c.do(req)
}

// getFunctionTokenCacheKey computes a cache key for caching a function access token based
// on the original id token that is exchanged for the function access token and the function
// name e.g. figlet.openfaas-fn.
// The original token is included in the hash to avoid cache hits for a function when the
// source token changes.
func getFunctionTokenCacheKey(idToken string, serviceName string) string {
	hash := sha256.New()
	hash.Write([]byte(idToken))
	hash.Write([]byte(serviceName))

	sum := hash.Sum(nil)
	return fmt.Sprintf("%x", sum)
}

package sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"

	"github.com/openfaas/faas-provider/types"
)

// Client is used to manage OpenFaaS functions
type Client struct {
	GatewayURL *url.URL
	Client     *http.Client
	ClientAuth ClientAuth
}

// ClientAuth an interface for client authentication.
// to add authentication to the client implement this interface
type ClientAuth interface {
	Set(req *http.Request) error
}

// NewClient creates an Client for managing OpenFaaS
func NewClient(gatewayURL *url.URL, auth ClientAuth, client *http.Client) *Client {
	return &Client{
		GatewayURL: gatewayURL,
		Client:     http.DefaultClient,
		ClientAuth: auth,
	}
}

// GetNamespaces get openfaas namespaces
func (s *Client) GetNamespaces() ([]string, error) {
	u := s.GatewayURL
	namespaces := []string{}
	u.Path = "/system/namespaces"

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return namespaces, fmt.Errorf("unable to create request: %s, error: %w", u.String(), err)
	}

	if s.ClientAuth != nil {
		if err := s.ClientAuth.Set(req); err != nil {
			return namespaces, fmt.Errorf("unable to set Authorization header: %w", err)
		}
	}

	res, err := s.Client.Do(req)
	if err != nil {
		return namespaces, fmt.Errorf("unable to make request: %w", err)
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	bytesOut, err := io.ReadAll(res.Body)
	if err != nil {
		return namespaces, err
	}

	if res.StatusCode == http.StatusUnauthorized {
		return namespaces, fmt.Errorf("check authorization, status code: %d", res.StatusCode)
	}

	if len(bytesOut) == 0 {
		return namespaces, nil
	}

	if err := json.Unmarshal(bytesOut, &namespaces); err != nil {
		return namespaces, fmt.Errorf("unable to marshal to JSON: %s, error: %w", string(bytesOut), err)
	}

	return namespaces, err
}

// GetFunctions lists all functions
func (s *Client) GetFunctions(namespace string) ([]types.FunctionStatus, error) {
	u := s.GatewayURL

	u.Path = "/system/functions"

	if len(namespace) > 0 {
		query := u.Query()
		query.Set("namespace", namespace)
		u.RawQuery = query.Encode()
	}

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return []types.FunctionStatus{}, fmt.Errorf("unable to create request for %s, error: %w", u.String(), err)
	}

	if s.ClientAuth != nil {
		if err := s.ClientAuth.Set(req); err != nil {
			return []types.FunctionStatus{}, fmt.Errorf("unable to set Authorization header: %w", err)
		}
	}

	res, err := s.Client.Do(req)
	if err != nil {
		return []types.FunctionStatus{}, fmt.Errorf("unable to make HTTP request: %w", err)
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	body, _ := io.ReadAll(res.Body)

	functions := []types.FunctionStatus{}
	if err := json.Unmarshal(body, &functions); err != nil {
		return []types.FunctionStatus{},
			fmt.Errorf("unable to unmarshal value: %q, error: %w", string(body), err)
	}

	return functions, nil
}

func (s *Client) GetInfo() (SystemInfo, error) {
	u := s.GatewayURL

	u.Path = "/system/info"

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return SystemInfo{}, fmt.Errorf("unable to create request for %s, error: %w", u.String(), err)
	}

	if s.ClientAuth != nil {
		if err := s.ClientAuth.Set(req); err != nil {
			return SystemInfo{}, fmt.Errorf("unable to set Authorization header: %w", err)
		}
	}

	res, err := s.Client.Do(req)
	if err != nil {
		return SystemInfo{}, fmt.Errorf("unable to make HTTP request: %w", err)
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	body, _ := io.ReadAll(res.Body)

	info := SystemInfo{}
	if err := json.Unmarshal(body, &info); err != nil {
		return SystemInfo{},
			fmt.Errorf("unable to unmarshal value: %q, error: %w", string(body), err)
	}

	return info, nil
}

// GetFunction gives a richer payload than GetFunctions, but for a specific function
func (s *Client) GetFunction(name, namespace string) (types.FunctionDeployment, error) {
	u := s.GatewayURL

	u.Path = "/system/function/" + name

	if len(namespace) > 0 {
		query := u.Query()
		query.Set("namespace", namespace)
		u.RawQuery = query.Encode()
	}

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return types.FunctionDeployment{}, fmt.Errorf("unable to create request for %s, error: %w", u.String(), err)
	}

	if s.ClientAuth != nil {
		if err := s.ClientAuth.Set(req); err != nil {
			return types.FunctionDeployment{}, fmt.Errorf("unable to set Authorization header: %w", err)
		}
	}

	res, err := s.Client.Do(req)
	if err != nil {
		return types.FunctionDeployment{}, fmt.Errorf("unable to make HTTP request: %w", err)
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	body, _ := io.ReadAll(res.Body)

	functions := types.FunctionDeployment{}
	if err := json.Unmarshal(body, &functions); err != nil {
		return types.FunctionDeployment{},
			fmt.Errorf("unable to unmarshal value: %q, error: %w", string(body), err)
	}

	return functions, nil
}

func (s *Client) Deploy(spec types.FunctionDeployment) (int, error) {
	return s.deploy(http.MethodPost, spec)

}

func (s *Client) Update(spec types.FunctionDeployment) (int, error) {
	return s.deploy(http.MethodPut, spec)
}

func (s *Client) deploy(method string, spec types.FunctionDeployment) (int, error) {

	bodyBytes, err := json.Marshal(spec)
	if err != nil {
		return http.StatusBadRequest, err
	}

	bodyReader := bytes.NewReader(bodyBytes)

	u := s.GatewayURL
	u.Path = "/system/functions"

	req, err := http.NewRequest(method, u.String(), bodyReader)
	if err != nil {
		return http.StatusBadGateway, err
	}

	if s.ClientAuth != nil {
		if err := s.ClientAuth.Set(req); err != nil {
			return http.StatusInternalServerError, fmt.Errorf("unable to set Authorization header: %w", err)
		}
	}

	res, err := s.Client.Do(req)
	if err != nil {
		return http.StatusBadGateway, err
	}

	var body []byte
	if res.Body != nil {
		defer res.Body.Close()
		body, _ = io.ReadAll(res.Body)
	}

	if res.StatusCode != http.StatusAccepted {
		return res.StatusCode, fmt.Errorf("unexpected status code: %d, message: %s", res.StatusCode, string(body))
	}

	return res.StatusCode, nil
}

// ScaleFunction scales a function to a number of replicas
func (s *Client) ScaleFunction(ctx context.Context, functionName, namespace string, replicas uint64) error {

	scaleReq := types.ScaleServiceRequest{
		ServiceName: functionName,
		Replicas:    replicas,
		Namespace:   namespace,
	}

	var err error

	bodyBytes, _ := json.Marshal(scaleReq)
	bodyReader := bytes.NewReader(bodyBytes)

	u := s.GatewayURL

	functionPath := filepath.Join("/system/scale-function", functionName)

	u.Path = functionPath

	req, err := http.NewRequest(http.MethodPost, u.String(), bodyReader)
	if err != nil {
		return fmt.Errorf("cannot connect to OpenFaaS on URL: %s, error: %s", u.String(), err)
	}

	if s.ClientAuth != nil {
		if err := s.ClientAuth.Set(req); err != nil {
			return fmt.Errorf("unable to set Authorization header: %w", err)
		}
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("cannot connect to OpenFaaS on URL: %s, error: %s", s.GatewayURL, err)

	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	switch res.StatusCode {
	case http.StatusAccepted, http.StatusOK, http.StatusCreated:
		break

	case http.StatusNotFound:
		return fmt.Errorf("function %s not found", functionName)

	case http.StatusUnauthorized:
		return fmt.Errorf("unauthorized action, please setup authentication for this server")

	default:
		var err error
		bytesOut, err := io.ReadAll(res.Body)
		if err != nil {
			return err
		}

		return fmt.Errorf("server returned unexpected status code %d, message: %q", res.StatusCode, string(bytesOut))
	}
	return nil
}

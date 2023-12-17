package sdk

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/openfaas/faas-provider/types"
)

func TestSdk_GetNamespaces_TwoNamespaces(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {

		rw.Write([]byte(`["openfaas-fn","dev"]`))
	}))

	sU, _ := url.Parse(s.URL)

	client := NewClient(sU, nil, http.DefaultClient)
	ns, err := client.GetNamespaces(context.Background())
	if err != nil {
		t.Fatalf("wanted no error, but got: %s", err)
	}
	want := 2
	if len(ns) != want {
		t.Fatalf("want %d namespaces, got: %d", want, len(ns))
	}
	wantNS := []string{"openfaas-fn", "dev"}
	gotNS := 0

	for _, n := range ns {
		for _, w := range wantNS {
			if n == w {
				gotNS++
			}
		}
	}
	if gotNS != len(wantNS) {
		t.Fatalf("want %d namespaces, got: %d", len(wantNS), gotNS)
	}
}

func TestSdk_GetNamespaces_NoNamespaces(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {

		rw.Write([]byte(`[]`))
	}))

	sU, _ := url.Parse(s.URL)

	client := NewClient(sU, nil, http.DefaultClient)
	ns, err := client.GetNamespaces(context.Background())
	if err != nil {
		t.Fatalf("wanted no error, but got: %s", err)
	}
	want := 0
	if len(ns) != want {
		t.Fatalf("want %d namespaces, got: %d", want, len(ns))
	}
	wantNS := []string{}
	gotNS := 0

	for _, n := range ns {
		for _, w := range wantNS {
			if n == w {
				gotNS++
			}
		}
	}
	if gotNS != len(wantNS) {
		t.Fatalf("want %d namespaces, got: %d", len(wantNS), gotNS)
	}
}

func TestSdk_DeployFunction(t *testing.T) {
	funcName := "funct1"
	nsName := "ns1"
	tests := []struct {
		name         string
		functionName string
		namespace    string
		err          error
		handler      func(rw http.ResponseWriter, req *http.Request)
	}{
		{
			name:         "function deployed",
			functionName: funcName,
			namespace:    nsName,
			handler: func(rw http.ResponseWriter, req *http.Request) {
				rw.WriteHeader(http.StatusOK)
			},
		},
		{
			name:         "function will bedeployed",
			functionName: funcName,
			namespace:    nsName,
			handler: func(rw http.ResponseWriter, req *http.Request) {
				rw.WriteHeader(http.StatusAccepted)
			},
		},
		{
			name:         "client not authorized",
			functionName: funcName,
			namespace:    nsName,
			handler: func(rw http.ResponseWriter, req *http.Request) {
				rw.WriteHeader(http.StatusUnauthorized)
			},
			err: fmt.Errorf("unauthorized action, please setup authentication for this server"),
		},
		{
			name:         "unknown error",
			functionName: funcName,
			namespace:    nsName,
			handler: func(rw http.ResponseWriter, req *http.Request) {
				http.Error(rw, "unknown error", http.StatusInternalServerError)
			},
			err: fmt.Errorf("unexpected status code: %d, message: %q", http.StatusInternalServerError, "unknown error\n"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := httptest.NewServer(http.HandlerFunc(test.handler))

			sU, _ := url.Parse(s.URL)

			client := NewClient(sU, nil, http.DefaultClient)

			_, err := client.Deploy(context.Background(), types.FunctionDeployment{
				Service:   funcName,
				Image:     fmt.Sprintf("docker.io/openfaas/%s:latest", funcName),
				Namespace: nsName,
			})

			if !errors.Is(err, test.err) && err.Error() != test.err.Error() {
				t.Fatalf("wanted %s, but got: %s", test.err, err)
			}
		})
	}
}

func TestSdk_DeleteFunction(t *testing.T) {
	funcName := "funct1"
	nsName := "ns1"
	tests := []struct {
		name         string
		functionName string
		namespace    string
		err          error
		handler      func(rw http.ResponseWriter, req *http.Request)
	}{
		{
			name:         "function deleted",
			functionName: funcName,
			namespace:    nsName,
			handler: func(rw http.ResponseWriter, req *http.Request) {
				rw.WriteHeader(http.StatusAccepted)
			},
		},
		{
			name:         "function not found",
			functionName: funcName,
			namespace:    nsName,
			handler: func(rw http.ResponseWriter, req *http.Request) {
				rw.WriteHeader(http.StatusNotFound)
			},
			err: fmt.Errorf("function %s not found", funcName),
		},
		{
			name:         "client not authorized",
			functionName: funcName,
			namespace:    nsName,
			handler: func(rw http.ResponseWriter, req *http.Request) {
				rw.WriteHeader(http.StatusUnauthorized)
			},
			err: fmt.Errorf("unauthorized action, please setup authentication for this server"),
		},
		{
			name:         "unknown error",
			functionName: funcName,
			namespace:    nsName,
			handler: func(rw http.ResponseWriter, req *http.Request) {
				http.Error(rw, "unknown error", http.StatusInternalServerError)
			},
			err: fmt.Errorf("server returned unexpected status code %d, message: %q", http.StatusInternalServerError, string("unknown error\n")),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := httptest.NewServer(http.HandlerFunc(test.handler))

			sU, _ := url.Parse(s.URL)

			client := NewClient(sU, nil, http.DefaultClient)
			err := client.DeleteFunction(context.Background(), test.functionName, test.namespace)

			if !errors.Is(err, test.err) && err.Error() != test.err.Error() {
				t.Fatalf("wanted %s, but got: %s", test.err, err)
			}
		})
	}
}

func TestSdk_GetNamespace(t *testing.T) {
	nsName := "ns1"
	tests := []struct {
		name      string
		namespace string
		err       error
		handler   func(rw http.ResponseWriter, req *http.Request)
	}{
		{
			name:      "namespace not found",
			namespace: nsName,
			handler: func(rw http.ResponseWriter, req *http.Request) {
				rw.WriteHeader(http.StatusNotFound)
			},
			err: fmt.Errorf("namespace %s not found", nsName),
		},
		{
			name:      "client not authorized",
			namespace: nsName,
			handler: func(rw http.ResponseWriter, req *http.Request) {
				rw.WriteHeader(http.StatusUnauthorized)
			},
			err: fmt.Errorf("unauthorized action, please setup authentication for this server"),
		},
		{
			name:      "unknown error",
			namespace: nsName,
			handler: func(rw http.ResponseWriter, req *http.Request) {
				http.Error(rw, string("unknown error"), http.StatusInternalServerError)
			},
			err: fmt.Errorf("unexpected status code: %d, message: %q", http.StatusInternalServerError, string("unknown error\n")),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := httptest.NewServer(http.HandlerFunc(test.handler))

			sU, _ := url.Parse(s.URL)

			client := NewClient(sU, nil, http.DefaultClient)
			_, err := client.GetNamespace(context.Background(), test.namespace)

			if !errors.Is(err, test.err) && err.Error() != test.err.Error() {
				t.Fatalf("wanted %s, but got: %s", test.err, err)
			}
		})
	}
}

func TestSdk_CreateNamespace(t *testing.T) {
	nsName := "ns1"
	tests := []struct {
		name    string
		req     types.FunctionNamespace
		err     error
		handler func(rw http.ResponseWriter, req *http.Request)
	}{
		{
			name: "namespace created with no label and annotation",
			req: types.FunctionNamespace{
				Name: nsName,
			},
			handler: func(rw http.ResponseWriter, req *http.Request) {
				rw.WriteHeader(http.StatusOK)
			},
		},
		{
			name: "namespace created with no label and annotation",
			req: types.FunctionNamespace{
				Name:        nsName,
				Labels:      map[string]string{},
				Annotations: map[string]string{},
			},
			handler: func(rw http.ResponseWriter, req *http.Request) {
				rw.WriteHeader(http.StatusOK)
			},
		},
		{
			name: "client not authorized",
			req: types.FunctionNamespace{
				Name: nsName,
			},
			handler: func(rw http.ResponseWriter, req *http.Request) {
				rw.WriteHeader(http.StatusUnauthorized)
			},
			err: fmt.Errorf("unauthorized action, please setup authentication for this server"),
		},
		{
			name: "unknown error",
			req: types.FunctionNamespace{
				Name: nsName,
			},
			handler: func(rw http.ResponseWriter, req *http.Request) {
				http.Error(rw, string("unknown error"), http.StatusInternalServerError)
			},
			err: fmt.Errorf("unexpected status code: %d, message: %q", http.StatusInternalServerError, string("unknown error\n")),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := httptest.NewServer(http.HandlerFunc(test.handler))

			sU, _ := url.Parse(s.URL)

			client := NewClient(sU, nil, http.DefaultClient)
			_, err := client.CreateNamespace(context.Background(), test.req)

			if !errors.Is(err, test.err) && err.Error() != test.err.Error() {
				t.Fatalf("wanted %s, but got: %s", test.err, err)
			}
		})
	}
}

func TestSdk_UpdateNamespace(t *testing.T) {
	nsName := "ns1"
	tests := []struct {
		name    string
		req     types.FunctionNamespace
		err     error
		handler func(rw http.ResponseWriter, req *http.Request)
	}{
		{
			name: "namespace updated with no label and annotation",
			req: types.FunctionNamespace{
				Name: nsName,
			},
			handler: func(rw http.ResponseWriter, req *http.Request) {
				rw.WriteHeader(http.StatusOK)
			},
		},
		{
			name: "namespace updated with no label and annotation",
			req: types.FunctionNamespace{
				Name:        nsName,
				Labels:      map[string]string{},
				Annotations: map[string]string{},
			},
			handler: func(rw http.ResponseWriter, req *http.Request) {
				rw.WriteHeader(http.StatusOK)
			},
		},
		{
			name: "namespace not found",
			req: types.FunctionNamespace{
				Name:        nsName,
				Labels:      map[string]string{},
				Annotations: map[string]string{},
			},
			handler: func(rw http.ResponseWriter, req *http.Request) {
				rw.WriteHeader(http.StatusNotFound)
			},
			err: fmt.Errorf("namespace %s not found", nsName),
		},
		{
			name: "client not authorized",
			req: types.FunctionNamespace{
				Name: nsName,
			},
			handler: func(rw http.ResponseWriter, req *http.Request) {
				rw.WriteHeader(http.StatusUnauthorized)
			},
			err: fmt.Errorf("unauthorized action, please setup authentication for this server"),
		},
		{
			name: "unknown error",
			req: types.FunctionNamespace{
				Name: nsName,
			},
			handler: func(rw http.ResponseWriter, req *http.Request) {
				http.Error(rw, string("unknown error"), http.StatusInternalServerError)
			},
			err: fmt.Errorf("unexpected status code: %d, message: %q", http.StatusInternalServerError, string("unknown error\n")),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := httptest.NewServer(http.HandlerFunc(test.handler))

			sU, _ := url.Parse(s.URL)

			client := NewClient(sU, nil, http.DefaultClient)
			_, err := client.UpdateNamespace(context.Background(), test.req)

			if !errors.Is(err, test.err) && err.Error() != test.err.Error() {
				t.Fatalf("wanted %s, but got: %s", test.err, err)
			}
		})
	}
}

func TestSdk_DeleteNamespace(t *testing.T) {
	nsName := "ns1"
	tests := []struct {
		name      string
		namespace string
		err       error
		handler   func(rw http.ResponseWriter, req *http.Request)
	}{
		{
			name:      "namespace deleted",
			namespace: nsName,
			handler: func(rw http.ResponseWriter, req *http.Request) {
				rw.WriteHeader(http.StatusOK)
			},
		},
		{
			name:      "namespace not found",
			namespace: nsName,
			handler: func(rw http.ResponseWriter, req *http.Request) {
				rw.WriteHeader(http.StatusNotFound)
			},
			err: fmt.Errorf("namespace %s not found", nsName),
		},
		{
			name:      "client not authorized",
			namespace: nsName,
			handler: func(rw http.ResponseWriter, req *http.Request) {
				rw.WriteHeader(http.StatusUnauthorized)
			},
			err: fmt.Errorf("unauthorized action, please setup authentication for this server"),
		},
		{
			name:      "unknown error",
			namespace: nsName,
			handler: func(rw http.ResponseWriter, req *http.Request) {
				http.Error(rw, string("unknown error"), http.StatusInternalServerError)
			},
			err: fmt.Errorf("unexpected status code: %d, message: %q", http.StatusInternalServerError, string("unknown error\n")),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := httptest.NewServer(http.HandlerFunc(test.handler))

			sU, _ := url.Parse(s.URL)

			client := NewClient(sU, nil, http.DefaultClient)
			err := client.DeleteNamespace(context.Background(), test.namespace)

			if !errors.Is(err, test.err) && err.Error() != test.err.Error() {
				t.Fatalf("wanted %s, but got: %s", test.err, err)
			}
		})
	}
}

func TestSdk_CreateSecret(t *testing.T) {
	nsName := "ns1"
	secretName := "secret1"
	secretValue := "secretValue1"
	tests := []struct {
		name    string
		req     types.Secret
		err     error
		handler func(rw http.ResponseWriter, req *http.Request)
	}{
		{
			name: "secret created",
			req: types.Secret{
				Name:      secretName,
				Namespace: nsName,
				Value:     secretValue,
			},
			handler: func(rw http.ResponseWriter, req *http.Request) {
				rw.WriteHeader(http.StatusOK)
			},
		},
		{
			name: "client not authorized",
			req: types.Secret{
				Name:      secretName,
				Namespace: nsName,
				Value:     secretValue,
			},
			handler: func(rw http.ResponseWriter, req *http.Request) {
				rw.WriteHeader(http.StatusUnauthorized)
			},
			err: fmt.Errorf("unauthorized action, please setup authentication for this server"),
		},
		{
			name: "unknown error",
			req: types.Secret{
				Name:      secretName,
				Namespace: nsName,
				Value:     secretValue,
			},
			handler: func(rw http.ResponseWriter, req *http.Request) {
				http.Error(rw, string("unknown error"), http.StatusInternalServerError)
			},
			err: fmt.Errorf("unexpected status code: %d, message: %q", http.StatusInternalServerError, string("unknown error\n")),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := httptest.NewServer(http.HandlerFunc(test.handler))

			sU, _ := url.Parse(s.URL)

			client := NewClient(sU, nil, http.DefaultClient)
			_, err := client.CreateSecret(context.Background(), test.req)

			if !errors.Is(err, test.err) && err.Error() != test.err.Error() {
				t.Fatalf("wanted %s, but got: %s", test.err, err)
			}
		})
	}
}

func TestSdk_UpdateSecret(t *testing.T) {
	nsName := "ns1"
	secretName := "secret1"
	secretValue := "secretValue1"
	tests := []struct {
		name    string
		req     types.Secret
		err     error
		handler func(rw http.ResponseWriter, req *http.Request)
	}{
		{
			name: "secret updated",
			req: types.Secret{
				Name:      secretName,
				Namespace: nsName,
				Value:     secretValue,
			},
			handler: func(rw http.ResponseWriter, req *http.Request) {
				rw.WriteHeader(http.StatusOK)
			},
		},
		{
			name: "client not authorized",
			req: types.Secret{
				Name:      secretName,
				Namespace: nsName,
				Value:     secretValue,
			},
			handler: func(rw http.ResponseWriter, req *http.Request) {
				rw.WriteHeader(http.StatusUnauthorized)
			},
			err: fmt.Errorf("unauthorized action, please setup authentication for this server"),
		},
		{
			name: "unknown error",
			req: types.Secret{
				Name:      secretName,
				Namespace: nsName,
				Value:     secretValue,
			},
			handler: func(rw http.ResponseWriter, req *http.Request) {
				http.Error(rw, string("unknown error"), http.StatusInternalServerError)
			},
			err: fmt.Errorf("unexpected status code: %d, message: %q", http.StatusInternalServerError, string("unknown error\n")),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := httptest.NewServer(http.HandlerFunc(test.handler))

			sU, _ := url.Parse(s.URL)

			client := NewClient(sU, nil, http.DefaultClient)
			_, err := client.UpdateSecret(context.Background(), test.req)

			if !errors.Is(err, test.err) && err.Error() != test.err.Error() {
				t.Fatalf("wanted %s, but got: %s", test.err, err)
			}
		})
	}
}

func TestSdk_DeleteSecret(t *testing.T) {
	secretName := "secret1"
	nsName := "ns1"
	tests := []struct {
		name       string
		secretName string
		namespace  string
		err        error
		handler    func(rw http.ResponseWriter, req *http.Request)
	}{
		{
			name:       "secret deleted",
			secretName: secretName,
			namespace:  nsName,
			handler: func(rw http.ResponseWriter, req *http.Request) {
				rw.WriteHeader(http.StatusAccepted)
			},
		},
		{
			name:       "secret not found",
			secretName: secretName,
			namespace:  nsName,
			handler: func(rw http.ResponseWriter, req *http.Request) {
				rw.WriteHeader(http.StatusNotFound)
			},
			err: fmt.Errorf("secret %s not found", secretName),
		},
		{
			name:       "client not authorized",
			secretName: secretName,
			namespace:  nsName,
			handler: func(rw http.ResponseWriter, req *http.Request) {
				rw.WriteHeader(http.StatusUnauthorized)
			},
			err: fmt.Errorf("unauthorized action, please setup authentication for this server"),
		},
		{
			name:       "unknown error",
			secretName: secretName,
			namespace:  nsName,
			handler: func(rw http.ResponseWriter, req *http.Request) {
				http.Error(rw, "unknown error", http.StatusInternalServerError)
			},
			err: fmt.Errorf("server returned unexpected status code %d, message: %q", http.StatusInternalServerError, string("unknown error\n")),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := httptest.NewServer(http.HandlerFunc(test.handler))

			sU, _ := url.Parse(s.URL)

			client := NewClient(sU, nil, http.DefaultClient)
			err := client.DeleteSecret(context.Background(), test.secretName, test.namespace)

			if !errors.Is(err, test.err) && err.Error() != test.err.Error() {
				t.Fatalf("wanted %s, but got: %s", test.err, err)
			}
		})
	}
}

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
			err: createHttpError(fmt.Errorf("unauthorized action, please setup authentication for this server"), http.StatusUnauthorized),
		},
		{
			name:         "unknown error",
			functionName: funcName,
			namespace:    nsName,
			handler: func(rw http.ResponseWriter, req *http.Request) {
				http.Error(rw, "unknown error", http.StatusInternalServerError)
			},
			err: createHttpError(fmt.Errorf("unexpected status code: %d, message: %q", http.StatusInternalServerError, "unknown error\n"), http.StatusInternalServerError),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := httptest.NewServer(http.HandlerFunc(test.handler))

			sU, _ := url.Parse(s.URL)

			client := NewClient(sU, nil, http.DefaultClient)

			// _, err := client.Deploy(context.Background(), types.FunctionDeployment{
			err := client.Deploy(context.Background(), types.FunctionDeployment{
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
			err: createHttpError(fmt.Errorf("function %s not found", funcName), http.StatusNotFound),
		},
		{
			name:         "client not authorized",
			functionName: funcName,
			namespace:    nsName,
			handler: func(rw http.ResponseWriter, req *http.Request) {
				rw.WriteHeader(http.StatusUnauthorized)
			},
			err: createHttpError(fmt.Errorf("unauthorized action, please setup authentication for this server"), http.StatusUnauthorized),
		},
		{
			name:         "unknown error",
			functionName: funcName,
			namespace:    nsName,
			handler: func(rw http.ResponseWriter, req *http.Request) {
				http.Error(rw, "unknown error", http.StatusInternalServerError)
			},
			err: createHttpError(fmt.Errorf("server returned unexpected status code %d, message: %q", http.StatusInternalServerError, string("unknown error\n")), http.StatusUnauthorized),
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

func TestSdk_ScaleFunction(t *testing.T) {
	funcName := "funct1"
	nsName := "ns1"
	tests := []struct {
		name         string
		functionName string
		namespace    string
		replicas     uint64
		err          error
		handler      func(rw http.ResponseWriter, req *http.Request)
	}{
		{
			name:         "scale request accepted",
			functionName: funcName,
			namespace:    nsName,
			replicas:     0,
			handler: func(rw http.ResponseWriter, req *http.Request) {
				rw.WriteHeader(http.StatusAccepted)
			},
		},
		{
			name:         "function not found",
			functionName: funcName,
			namespace:    nsName,
			replicas:     0,
			handler: func(rw http.ResponseWriter, req *http.Request) {
				rw.WriteHeader(http.StatusNotFound)
			},
			err: createHttpError(fmt.Errorf("function %s not found", funcName), http.StatusNotFound),
		},
		{
			name:         "client not authorized",
			functionName: funcName,
			namespace:    nsName,
			replicas:     0,
			handler: func(rw http.ResponseWriter, req *http.Request) {
				rw.WriteHeader(http.StatusUnauthorized)
			},
			err: createHttpError(fmt.Errorf("unauthorized action, please setup authentication for this server"), http.StatusUnauthorized),
		},
		{
			name:         "unknown error",
			functionName: funcName,
			namespace:    nsName,
			replicas:     0,
			handler: func(rw http.ResponseWriter, req *http.Request) {
				http.Error(rw, "unknown error", http.StatusInternalServerError)
			},
			err: createHttpError(fmt.Errorf("server returned unexpected status code %d, message: %q", http.StatusInternalServerError, string("unknown error\n")), http.StatusUnauthorized),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := httptest.NewServer(http.HandlerFunc(test.handler))

			sU, _ := url.Parse(s.URL)

			client := NewClient(sU, nil, http.DefaultClient)
			err := client.ScaleFunction(context.Background(), test.functionName, test.namespace, test.replicas)

			if !errors.Is(err, test.err) && err.Error() != test.err.Error() {
				t.Fatalf("wanted %s, but got: %s", test.err, err)
			}
		})
	}
}

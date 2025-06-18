package httpclient

import (
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func Test_dumpRequest(t *testing.T) {
	tests := []struct {
		name string
		req  *http.Request
		want string
	}{
		{
			name: "request without body",
			req: &http.Request{
				Method: http.MethodPost,
				URL: &url.URL{
					Scheme: "https",
					Host:   "gw.example.com",
					Path:   "/function/env.openfaas-fn",
				},
			},
			want: "POST https://gw.example.com/function/env.openfaas-fn\n",
		},
		{
			name: "request with headers",
			req: &http.Request{
				Method: http.MethodPost,
				URL: &url.URL{
					Scheme: "https",
					Host:   "gw.example.com",
					Path:   "/function/env.openfaas-fn",
				},
				Header: http.Header{
					"Content-Type": []string{"text/plain"},
					"User-Agent":   []string{"openfaas-go-sdk"},
				},
			},
			want: "POST https://gw.example.com/function/env.openfaas-fn\n" +
				"Content-Type: [text/plain]\n" +
				"User-Agent: [openfaas-go-sdk]\n",
		},
		{
			name: "request with body",
			req: &http.Request{
				Method: http.MethodPost,
				URL: &url.URL{
					Scheme: "https",
					Host:   "gw.example.com",
					Path:   "/function/env.openfaas-fn",
				},
				Header: http.Header{
					"Content-Type": []string{"text/plain"},
				},
				Body: io.NopCloser(strings.NewReader("Hello OpenFaaS!!")),
			},
			want: "POST https://gw.example.com/function/env.openfaas-fn\n" +
				"Content-Type: [text/plain]\n" +
				"Hello OpenFaaS!!\n",
		},
		{
			name: "request with bearer auth",
			req: &http.Request{
				Method: http.MethodPost,
				URL: &url.URL{
					Scheme: "https",
					Host:   "gw.example.com",
					Path:   "/function/env.openfaas-fn",
				},
				Header: http.Header{
					"Authorization": []string{"Bearer secret openfaas-token"},
				},
			},
			want: "POST https://gw.example.com/function/env.openfaas-fn\n" +
				"Authorization: Bearer [REDACTED]\n",
		},
		{
			name: "request with basic auth",
			req: &http.Request{
				Method: http.MethodPost,
				URL: &url.URL{
					Scheme: "https",
					Host:   "gw.example.com",
					Path:   "/function/env.openfaas-fn",
				},
				Header: http.Header{
					"Authorization": []string{"Basic username:password"},
				},
			},
			want: "POST https://gw.example.com/function/env.openfaas-fn\n" +
				"Authorization: Basic [REDACTED]\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := DumpRequest(test.req)

			if err != nil {
				t.Errorf("want %s, but got error: %s", test.want, err)
			}

			if test.want != got {
				t.Errorf("want %s, but got: %s", test.want, got)
			}
		})
	}
}

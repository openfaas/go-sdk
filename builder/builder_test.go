package builder

import (
	"archive/tar"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	hmac "github.com/alexellis/hmac/v2"
	"github.com/google/go-cmp/cmp"
	"github.com/openfaas/go-sdk/seal"
)

// readCloser wraps an io.ReadCloser and tracks when it's closed
type readCloser struct {
	io.ReadCloser
	closed bool
}

func (r *readCloser) Close() error {
	r.closed = true
	return r.ReadCloser.Close()
}

func Test_BuildResultStream_Results(t *testing.T) {
	// Open the test file
	file, err := os.Open("../testdata/buildlogs.ndjson")
	if err != nil {
		t.Fatalf("Failed to open test file: %v", err)
	}
	defer file.Close()

	// Create a BuildResultStream with the test file
	stream := &BuildResultStream{r: file}

	// Collect results from the stream
	var results []BuildResult
	for result, err := range stream.Results() {
		if err != nil {
			t.Fatalf("Unexpected error from stream: %v", err)
		}
		results = append(results, result)
	}

	// Verify we got the expected number of results
	if len(results) != 40 {
		t.Errorf("Expected 40 results, got %d", len(results))
	}

	// Verify the first result
	wantFirst := BuildResult{
		Log:    []string{"v: 2025-06-13T20:16:16Z [internal] load build definition from Dockerfile"},
		Status: "in_progress",
	}
	if diff := cmp.Diff(wantFirst, results[0]); diff != "" {
		t.Errorf("First result mismatch:\n%s", diff)
	}

	// Verify the last result
	wantLast := BuildResult{
		Image:  "ttl.sh/openfaas/test-image-hello:10m",
		Status: "success",
	}
	if diff := cmp.Diff(wantLast, results[39]); diff != "" {
		t.Errorf("Last result mismatch:\n%s", diff)
	}
}

func Test_BuildResultStream_ReaderClosed(t *testing.T) {
	t.Run("reader closed after normal completion", func(t *testing.T) {
		// Open the test file
		file, err := os.Open("../testdata/buildlogs.ndjson")
		if err != nil {
			t.Fatalf("Failed to open test file: %v", err)
		}
		defer file.Close()

		// Wrap the file in our readCloser
		cr := &readCloser{ReadCloser: file}

		// Create a BuildResultStream with the wrapped reader
		stream := &BuildResultStream{r: cr}

		// Iterate over results
		for result, err := range stream.Results() {
			if err != nil {
				t.Fatalf("Unexpected error from stream: %v", err)
			}
			// Verify we got a valid result
			if result.Status == "" {
				t.Error("Expected non-empty status in result")
			}
		}

		// Verify the reader was closed
		if !cr.closed {
			t.Error("Expected reader to be closed")
		}
	})

	t.Run("reader closed after early break", func(t *testing.T) {
		// Open the test file
		file, err := os.Open("../testdata/buildlogs.ndjson")
		if err != nil {
			t.Fatalf("Failed to open test file: %v", err)
		}
		defer file.Close()

		// Wrap the file in our readCloser
		cr := &readCloser{ReadCloser: file}

		// Create a BuildResultStream with the wrapped reader
		stream := &BuildResultStream{r: cr}

		// Iterate over results
		count := 0
		for result, err := range stream.Results() {
			if err != nil {
				t.Fatalf("Unexpected error from stream: %v", err)
			}
			// Verify we got a valid result
			if result.Status == "" {
				t.Error("Expected non-empty status in result")
			}
			count++

			if count >= 5 {
				break
			}
		}

		// Verify the reader was closed
		if !cr.closed {
			t.Error("Expected reader to be closed")
		}
	})
}

func TestBuildWithSecrets(t *testing.T) {
	pub, priv, err := seal.GenerateKeyPair()
	if err != nil {
		t.Fatalf("seal.GenerateKeyPair: %v", err)
	}

	buildTar := createTestTar(t)

	tmpFile, err := os.CreateTemp(t.TempDir(), "build-*.tar")
	if err != nil {
		t.Fatalf("os.CreateTemp returned error: %v", err)
	}
	if _, err := tmpFile.Write(buildTar); err != nil {
		t.Fatalf("tmpFile.Write returned error: %v", err)
	}
	tmpFile.Close()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("io.ReadAll returned error: %v", err)
		}

		// Verify HMAC
		wantDigest := hmac.Sign(body, []byte("payload-secret"), sha256.New)
		gotDigest := r.Header.Get("X-Build-Signature")
		if gotDigest != "sha256="+hex.EncodeToString(wantDigest) {
			t.Fatalf("unexpected signature: %s", gotDigest)
		}

		// Body should be a tar, not multipart
		if ct := r.Header.Get("Content-Type"); ct != "application/octet-stream" {
			t.Fatalf("unexpected content-type: %s", ct)
		}

		// Extract sealed secrets from tar
		tr := tar.NewReader(bytes.NewReader(body))
		var sealedData []byte
		for {
			hdr, err := tr.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Fatalf("tar.Next returned error: %v", err)
			}
			if hdr.Name == BuildSecretsFileName {
				sealedData, err = io.ReadAll(tr)
				if err != nil {
					t.Fatalf("io.ReadAll sealed secrets: %v", err)
				}
			}
		}

		if sealedData == nil {
			t.Fatal("sealed secrets file not found in tar")
		}

		// Unseal and verify
		secrets, err := seal.Unseal(priv, sealedData)
		if err != nil {
			t.Fatalf("seal.Unseal returned error: %v", err)
		}

		if got := string(secrets["pip_token"]); got != "s3cr3t" {
			t.Fatalf("want pip_token to be %q, got %q", "s3cr3t", got)
		}

		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"status":"success","image":"ttl.sh/test:latest"}`)
	}))
	defer server.Close()

	serverURL, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("url.Parse returned error: %v", err)
	}

	builder := NewFunctionBuilder(serverURL, http.DefaultClient,
		WithHmacAuth("payload-secret"),
		WithBuildSecretsKey(pub))

	result, err := builder.BuildWithSecrets(tmpFile.Name(), map[string]string{
		"pip_token": "s3cr3t",
	})
	if err != nil {
		t.Fatalf("BuildWithSecrets returned error: %v", err)
	}

	if result.Status != BuildSuccess {
		t.Fatalf("want status %q, got %q", BuildSuccess, result.Status)
	}
}

// createTestTar creates a minimal valid tar for testing.
func createTestTar(t *testing.T) []byte {
	t.Helper()
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	data := []byte(`{"image":"test:latest"}`)
	if err := tw.WriteHeader(&tar.Header{
		Name: BuilderConfigFileName,
		Mode: 0600,
		Size: int64(len(data)),
	}); err != nil {
		t.Fatalf("tar.WriteHeader: %v", err)
	}
	if _, err := tw.Write(data); err != nil {
		t.Fatalf("tar.Write: %v", err)
	}
	if err := tw.Close(); err != nil {
		t.Fatalf("tar.Close: %v", err)
	}
	return buf.Bytes()
}

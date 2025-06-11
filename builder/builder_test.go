package builder

import (
	"io"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
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

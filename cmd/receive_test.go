package cmd

import (
	"net/http"
	"testing"
)

func TestDeleteItem(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		setupViper(t)

		var method, path string
		newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			method = r.Method
			path = r.URL.Path
			w.WriteHeader(200)
			_, _ = w.Write([]byte(`{}`))
		})

		err := deleteItem(42, &http.Client{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if method != "DELETE" {
			t.Errorf("method = %q, want DELETE", method)
		}
		if path != "/v1/receive/42" {
			t.Errorf("path = %q, want /v1/receive/42", path)
		}
	})

	t.Run("not found", func(t *testing.T) {
		setupViper(t)
		newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(404)
			_, _ = w.Write([]byte("not found"))
		})

		err := deleteItem(999, &http.Client{})
		if err == nil {
			t.Fatal("expected error for 404, got nil")
		}
	})

	t.Run("server error", func(t *testing.T) {
		setupViper(t)
		newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(500)
			_, _ = w.Write([]byte("internal error"))
		})

		err := deleteItem(1, &http.Client{})
		if err == nil {
			t.Fatal("expected error for 500, got nil")
		}
	})
}

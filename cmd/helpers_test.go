package cmd

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

func setupViper(t *testing.T) {
	t.Helper()
	viper.Reset()
	viper.Set("host", "https://api.invoicetronic.com")
	viper.Set("apikey", "test-api-key")
	viper.Set("verbose", false)
}

func newTestServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	viper.Set("host", server.URL)
	return server
}

func createTempFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return p
}

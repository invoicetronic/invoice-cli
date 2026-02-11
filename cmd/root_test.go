package cmd

import (
	"bytes"
	"encoding/base64"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

// --- Phase 1: pure function tests ---

func TestBasicAuth(t *testing.T) {
	tests := []struct {
		name   string
		apikey string
		want   string
	}{
		{"simple key", "mykey", base64.StdEncoding.EncodeToString([]byte("mykey:"))},
		{"empty key", "", base64.StdEncoding.EncodeToString([]byte(":"))},
		{"key with special chars", "abc!@#$%", base64.StdEncoding.EncodeToString([]byte("abc!@#$%:"))},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Set("apikey", tt.apikey)
			got := BasicAuth()
			if got != tt.want {
				t.Errorf("BasicAuth() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBuildEndpointUrl(t *testing.T) {
	tests := []struct {
		name string
		host string
		args []string
		want string
	}{
		{"send endpoint", "https://api.example.com", []string{"send"}, "https://api.example.com/v1/send"},
		{"receive endpoint", "https://api.example.com", []string{"receive"}, "https://api.example.com/v1/receive"},
		{"receive with id", "https://api.example.com", []string{"receive", "42"}, "https://api.example.com/v1/receive/42"},
		{"host with trailing slash", "https://api.example.com/", []string{"send"}, "https://api.example.com/v1/send"},
		{"no extra path", "https://api.example.com", nil, "https://api.example.com/v1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Set("host", tt.host)
			got := BuildEndpointUrl(tt.args...)
			if got.String() != tt.want {
				t.Errorf("BuildEndpointUrl(%v) = %q, want %q", tt.args, got.String(), tt.want)
			}
		})
	}
}

func TestGetFullFilePath(t *testing.T) {
	sep := string(os.PathSeparator)
	tests := []struct {
		name     string
		dest     string
		filename string
		want     string
	}{
		{"without trailing separator", "/tmp/out", "invoice.xml", filepath.Join("/tmp/out", "invoice.xml")},
		{"with trailing separator", "/tmp/out" + sep, "invoice.xml", filepath.Join("/tmp/out", "invoice.xml")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getFullFilePath(tt.dest, tt.filename)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("getFullFilePath(%q, %q) = %q, want %q", tt.dest, tt.filename, got, tt.want)
			}
		})
	}
}

func TestCreateDirectoryIfNotExists(t *testing.T) {
	t.Run("creates new directory", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "newdir")
		err := createDirectoryIfNotExists(dir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		info, err := os.Stat(dir)
		if err != nil {
			t.Fatalf("directory not created: %v", err)
		}
		if !info.IsDir() {
			t.Error("expected a directory")
		}
	})

	t.Run("existing directory is no-op", func(t *testing.T) {
		dir := t.TempDir()
		err := createDirectoryIfNotExists(dir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestVerbose(t *testing.T) {
	t.Run("verbose enabled", func(t *testing.T) {
		verbose = true
		defer func() { verbose = false }()

		var buf bytes.Buffer
		log.SetOutput(&buf)
		defer log.SetOutput(os.Stderr)

		Verbose("test message %s", "hello")

		if !strings.Contains(buf.String(), "test message hello") {
			t.Errorf("expected log output containing 'test message hello', got %q", buf.String())
		}
	})

	t.Run("verbose disabled", func(t *testing.T) {
		verbose = false

		var buf bytes.Buffer
		log.SetOutput(&buf)
		defer log.SetOutput(os.Stderr)

		Verbose("should not appear")

		if buf.Len() > 0 {
			t.Errorf("expected no output when verbose is disabled, got %q", buf.String())
		}
	})
}

// --- Phase 3: HTTP and File I/O tests ---

func TestPerformRequest_Success(t *testing.T) {
	setupViper(t)

	server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth == "" {
			t.Error("missing Authorization header")
		}
		if !strings.HasPrefix(auth, "Basic ") {
			t.Errorf("Authorization header should start with 'Basic ', got %q", auth)
		}
		ct := r.Header.Get("Content-Type")
		if ct != "application/json" {
			t.Errorf("Content-Type = %q, want application/json", ct)
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"ok":true}`))
	})

	req, _ := http.NewRequest("GET", server.URL+"/v1/test", nil)
	resp, body, err := performRequest(req, &http.Client{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
	if string(body) != `{"ok":true}` {
		t.Errorf("body = %q, want %q", string(body), `{"ok":true}`)
	}
}

func TestPerformRequest_HttpError(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
	}{
		{"bad request", 400},
		{"unauthorized", 401},
		{"internal server error", 500},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupViper(t)
			newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte("error details"))
			})

			req, _ := http.NewRequest("GET", viper.GetString("host")+"/v1/test", nil)
			_, _, err := performRequest(req, &http.Client{})
			if err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}

func TestPerformRequest_NetworkError(t *testing.T) {
	setupViper(t)
	server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {})
	server.Close()

	req, _ := http.NewRequest("GET", server.URL+"/v1/test", nil)
	_, _, err := performRequest(req, &http.Client{})
	if err == nil {
		t.Fatal("expected error for closed server, got nil")
	}
}

func TestToFile_Base64(t *testing.T) {
	setupViper(t)
	dir := t.TempDir()
	content := "hello, world!"
	encoded := base64.StdEncoding.EncodeToString([]byte(content))

	err := toFile("test.xml", encoded, "Base64", dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "test.xml"))
	if err != nil {
		t.Fatalf("error reading file: %v", err)
	}
	if string(data) != content {
		t.Errorf("file content = %q, want %q", string(data), content)
	}
}

func TestToFile_PlainText(t *testing.T) {
	setupViper(t)
	dir := t.TempDir()
	content := "<invoice>data</invoice>"

	err := toFile("test.xml", content, "None", dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "test.xml"))
	if err != nil {
		t.Fatalf("error reading file: %v", err)
	}
	if string(data) != content {
		t.Errorf("file content = %q, want %q", string(data), content)
	}
}

func TestToFile_InvalidBase64(t *testing.T) {
	setupViper(t)
	dir := t.TempDir()

	err := toFile("test.xml", "not-valid-base64!!!", "Base64", dir)
	if err == nil {
		t.Fatal("expected error for invalid base64, got nil")
	}
}

func TestToFile_CreatesDirectory(t *testing.T) {
	setupViper(t)
	dir := filepath.Join(t.TempDir(), "sub", "dir")

	err := toFile("test.xml", "content", "None", dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = os.Stat(filepath.Join(dir, "test.xml"))
	if err != nil {
		t.Errorf("file not created in nested directory: %v", err)
	}
}

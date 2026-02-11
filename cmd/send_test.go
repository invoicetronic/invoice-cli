package cmd

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/invoicetronic/invoice/models"
	"github.com/spf13/viper"
)

// --- Phase 1: pure function tests ---

func TestCapitalizeFirst(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"auto", "Auto"},
		{"apply", "Apply"},
		{"none", "None"},
		{"", ""},
		{"A", "A"},
		{"already", "Already"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := capitalizeFirst(tt.input)
			if got != tt.want {
				t.Errorf("capitalizeFirst(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// --- Phase 3: HTTP and integration tests ---

func TestBuildItems(t *testing.T) {
	t.Run("single file", func(t *testing.T) {
		dir := t.TempDir()
		content := "<invoice>test</invoice>"
		path := createTempFile(t, dir, "test.xml", content)

		items, err := buildItems([]string{path})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(items) != 1 {
			t.Fatalf("got %d items, want 1", len(items))
		}
		if items[0].FileName != "test.xml" {
			t.Errorf("FileName = %q, want %q", items[0].FileName, "test.xml")
		}
		decoded, _ := base64.StdEncoding.DecodeString(items[0].Payload)
		if string(decoded) != content {
			t.Errorf("decoded payload = %q, want %q", string(decoded), content)
		}
	})

	t.Run("glob pattern", func(t *testing.T) {
		dir := t.TempDir()
		createTempFile(t, dir, "a.xml", "content-a")
		createTempFile(t, dir, "b.xml", "content-b")
		createTempFile(t, dir, "c.txt", "content-c")

		items, err := buildItems([]string{filepath.Join(dir, "*.xml")})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(items) != 2 {
			t.Fatalf("got %d items, want 2", len(items))
		}
	})

	t.Run("nonexistent file returns empty", func(t *testing.T) {
		items, err := buildItems([]string{"/nonexistent/file.xml"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(items) != 0 {
			t.Errorf("got %d items, want 0", len(items))
		}
	})
}

func TestSendItems(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		setupViper(t)

		var received []models.SendItem
		server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			var item models.SendItem
			_ = json.Unmarshal(body, &item)
			received = append(received, item)
			w.WriteHeader(200)
			_, _ = w.Write([]byte(`{}`))
		})

		items := []models.SendItem{
			{FileName: "test1.xml", Payload: "dGVzdDE="},
			{FileName: "test2.xml", Payload: "dGVzdDI="},
		}

		err := sendItems(items, &http.Client{}, server.URL+"/v1/send", false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(received) != 2 {
			t.Errorf("server received %d requests, want 2", len(received))
		}
	})

	t.Run("server error", func(t *testing.T) {
		setupViper(t)
		newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(500)
			_, _ = w.Write([]byte("internal error"))
		})

		items := []models.SendItem{
			{FileName: "test.xml", Payload: "dGVzdA=="},
		}

		err := sendItems(items, &http.Client{}, viper.GetString("host")+"/v1/send", false)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("delete after send", func(t *testing.T) {
		setupViper(t)
		dir := t.TempDir()
		path := createTempFile(t, dir, "delete-me.xml", "content")

		newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
			_, _ = w.Write([]byte(`{}`))
		})

		items := []models.SendItem{
			{FileName: "delete-me.xml", FilePath: path, Payload: "dGVzdA=="},
		}

		err := sendItems(items, &http.Client{}, viper.GetString("host")+"/v1/send", true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Error("file should have been deleted after send")
		}
	})
}

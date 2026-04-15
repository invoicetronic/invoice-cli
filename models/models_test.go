package models

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestSendItemJSONRoundTrip(t *testing.T) {
	item := SendItem{
		Payload:  "dGVzdA==",
		FileName: "test.xml",
		FilePath: "/tmp/test.xml",
	}

	data, err := json.Marshal(item)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// FilePath is tagged json:"-" because it is a local-only field
	// that must not be sent to the API.
	want := item
	want.FilePath = ""

	var got SendItem
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if got != want {
		t.Errorf("round-trip mismatch: got %+v, want %+v", got, want)
	}
}

func TestSendItemJSONTags(t *testing.T) {
	item := SendItem{
		Payload:  "dGVzdA==",
		FileName: "test.xml",
		FilePath: "/tmp/test.xml",
	}

	data, err := json.Marshal(item)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	s := string(data)

	// The API uses snake_case property naming, so payload and file_name
	// must be emitted in snake_case. FilePath is local-only and must not
	// appear in the payload.
	if !strings.Contains(s, `"payload":"dGVzdA=="`) {
		t.Errorf("marshaled JSON should contain snake_case \"payload\": %s", s)
	}
	if !strings.Contains(s, `"file_name":"test.xml"`) {
		t.Errorf("marshaled JSON should contain snake_case \"file_name\": %s", s)
	}
	if strings.Contains(s, "FilePath") || strings.Contains(s, "file_path") {
		t.Errorf("marshaled JSON must not leak FilePath: %s", s)
	}
}

func TestReceiveItemJSONRoundTrip(t *testing.T) {
	item := ReceiveItem{
		Id:       42,
		Payload:  "dGVzdA==",
		FileName: "invoice.xml",
		Encoding: "Base64",
	}

	data, err := json.Marshal(item)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var got ReceiveItem
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if got != item {
		t.Errorf("round-trip mismatch: got %+v, want %+v", got, item)
	}
}

func TestReceiveItemJSONTags(t *testing.T) {
	jsonData := `{"Id":1,"payload":"dGVzdA==","file_name":"test.xml","encoding":"Base64"}`

	var item ReceiveItem
	if err := json.Unmarshal([]byte(jsonData), &item); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if item.Id != 1 {
		t.Errorf("Id = %d, want 1", item.Id)
	}
	if item.Payload != "dGVzdA==" {
		t.Errorf("Payload = %q, want %q", item.Payload, "dGVzdA==")
	}
	if item.FileName != "test.xml" {
		t.Errorf("FileName = %q, want %q", item.FileName, "test.xml")
	}
	if item.Encoding != "Base64" {
		t.Errorf("Encoding = %q, want %q", item.Encoding, "Base64")
	}
}

func TestReceiveItemIdFieldBehavior(t *testing.T) {
	// Id has no json tag — Go's json.Unmarshal matches case-insensitively,
	// so lowercase "id" from the API will match the Id field.
	jsonData := `{"id":42,"payload":"data","file_name":"test.xml","encoding":"Base64"}`

	var item ReceiveItem
	if err := json.Unmarshal([]byte(jsonData), &item); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if item.Id != 42 {
		t.Errorf("Id = %d, want 42 (lowercase 'id' should match via case-insensitive matching)", item.Id)
	}

	// Marshaling produces "Id" (capitalized, no json tag) not "id"
	data, _ := json.Marshal(item)
	if !strings.Contains(string(data), `"Id"`) {
		t.Errorf("marshaled JSON should contain 'Id' (capitalized, no json tag): %s", string(data))
	}
}

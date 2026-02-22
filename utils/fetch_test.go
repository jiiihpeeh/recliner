package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestFetch_GetJSON(t *testing.T) {
	// Mock server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"message": "hello"}`)
	}))
	defer ts.Close()

	// Test
	resp := FetchSync(ts.URL, nil)
	if !resp.Ok() {
		t.Errorf("Expected Ok(), got status %d", resp.StatusCode)
	}

	var data map[string]string
	err := resp.Json(&data)
	if err != nil {
		t.Errorf("Failed to parse JSON: %v", err)
	}

	if data["message"] != "hello" {
		t.Errorf("Expected message 'hello', got '%s'", data["message"])
	}
}

func TestFetch_PostJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}

		var reqData map[string]string
		json.NewDecoder(r.Body).Decode(&reqData)
		if reqData["name"] != "recliner" {
			t.Errorf("Expected name 'recliner', got '%s'", reqData["name"])
		}

		w.WriteHeader(http.StatusCreated)
		fmt.Fprintln(w, `{"status": "created"}`)
	}))
	defer ts.Close()

	opts := &FetchOptions{
		Method: "POST",
		Body:   map[string]string{"name": "recliner"},
	}
	resp := FetchSync(ts.URL, opts)

	if !resp.Ok() {
		t.Errorf("Expected Ok(), got status %d", resp.StatusCode)
	}

	var resData map[string]string
	resp.Json(&resData)
	if resData["status"] != "created" {
		t.Errorf("Expected status 'created', got '%s'", resData["status"])
	}
}

func TestFetch_Headers(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer token" {
			t.Errorf("Expected Authorization 'Bearer token', got '%s'", auth)
		}
	}))
	defer ts.Close()

	opts := &FetchOptions{
		Headers: map[string]string{
			"Authorization": "Bearer token",
		},
	}
	FetchSync(ts.URL, opts)
}

func TestFetch_BinaryResponse(t *testing.T) {
	expected := []byte{0xDE, 0xAD, 0xBE, 0xEF}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write(expected)
	}))
	defer ts.Close()

	resp := FetchSync(ts.URL, nil)

	// Test .Bytes()
	b, err := resp.Bytes()
	if err != nil {
		t.Fatalf("Bytes() error: %v", err)
	}
	if !reflect.DeepEqual(b, expected) {
		t.Errorf("Bytes(): expected %v, got %v", expected, b)
	}

	// Make a fresh request for .Blob() test because body is consumed
	resp2 := FetchSync(ts.URL, nil)
	blob, err := resp2.Blob()
	if err != nil {
		t.Fatalf("Blob() error: %v", err)
	}
	if blob.Type != "application/octet-stream" {
		t.Errorf("Blob type: expected application/octet-stream, got %s", blob.Type)
	}
	if !reflect.DeepEqual(blob.Data, expected) {
		t.Errorf("Blob data: expected %v, got %v", expected, blob.Data)
	}
}

func TestFetch_PostBinary(t *testing.T) {
	payload := []byte{1, 2, 3, 4}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if !reflect.DeepEqual(body, payload) {
			t.Errorf("Server received wrong bytes: %v", body)
		}
		if r.Header.Get("Content-Type") != "application/octet-stream" {
			t.Errorf("Wrong content type: %s", r.Header.Get("Content-Type"))
		}
	}))
	defer ts.Close()

	opts := &FetchOptions{
		Method: "POST",
		Body:   payload,
	}
	FetchSync(ts.URL, opts)
}

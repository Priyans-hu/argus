package ai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClient_IsAvailable_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/tags" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"models":[]}`))
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "test-model", 5*time.Second)
	if !client.IsAvailable(context.Background()) {
		t.Error("expected IsAvailable to return true")
	}
}

func TestClient_IsAvailable_ServerDown(t *testing.T) {
	client := NewClient("http://localhost:1", "test-model", 1*time.Second)
	if client.IsAvailable(context.Background()) {
		t.Error("expected IsAvailable to return false for unreachable server")
	}
}

func TestClient_Generate_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/generate" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			return
		}
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
			return
		}

		var req generateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode request: %v", err)
			return
		}

		if req.Model != "test-model" {
			t.Errorf("expected model 'test-model', got '%s'", req.Model)
		}
		if req.Stream {
			t.Error("expected stream=false")
		}

		resp := generateResponse{
			Response: "This is the AI response.",
			Done:     true,
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "test-model", 5*time.Second)
	result, err := client.Generate(context.Background(), "Hello")
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	if result != "This is the AI response." {
		t.Errorf("expected AI response, got %q", result)
	}
}

func TestClient_Generate_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("model not found"))
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "bad-model", 5*time.Second)
	_, err := client.Generate(context.Background(), "test")
	if err == nil {
		t.Error("expected error for 500 response")
	}
}

func TestClient_Generate_ContextCancelled(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second)
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	client := NewClient(srv.URL, "test-model", 5*time.Second)
	_, err := client.Generate(ctx, "test")
	if err == nil {
		t.Error("expected error for cancelled context")
	}
}

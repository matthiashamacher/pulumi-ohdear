package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientDo(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		if r.URL.Path == "/boom" {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"message":"nope"}`))
			return
		}
		_, _ = w.Write([]byte(`{"id":7}`))
	}))
	defer srv.Close()

	c := NewClient("tok123")
	c.baseURL = srv.URL

	var out struct {
		ID int `json:"id"`
	}
	if err := c.Do(context.Background(), http.MethodGet, "/ok", nil, &out); err != nil {
		t.Fatalf("Do: %v", err)
	}
	if gotAuth != "Bearer tok123" {
		t.Fatalf("Authorization = %q, want %q", gotAuth, "Bearer tok123")
	}
	if out.ID != 7 {
		t.Fatalf("out.ID = %d, want 7", out.ID)
	}

	if err := c.Do(context.Background(), http.MethodGet, "/boom", nil, nil); err == nil {
		t.Fatal("expected error on 401, got nil")
	}
}

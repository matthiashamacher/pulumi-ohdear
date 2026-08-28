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

func TestListEnvelopeTolerance(t *testing.T) {
	type item struct {
		ID int `json:"id"`
	}
	cases := map[string]string{
		"bare array":    `[{"id":1},{"id":2}]`,
		"data envelope": `{"data":[{"id":1},{"id":2}],"meta":{}}`,
	}
	for name, payload := range cases {
		t.Run(name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				_, _ = w.Write([]byte(payload))
			}))
			defer srv.Close()

			c := NewClient("t")
			c.baseURL = srv.URL

			got, err := List[item](context.Background(), c, "/x")
			if err != nil {
				t.Fatalf("List: %v", err)
			}
			if len(got) != 2 || got[0].ID != 1 || got[1].ID != 2 {
				t.Fatalf("got %+v, want ids [1 2]", got)
			}
		})
	}
}

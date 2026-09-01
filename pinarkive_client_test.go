package pinarkive

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewClientDefaults(t *testing.T) {
	c := NewClient("tok", "", "")
	if c.BaseURL != "https://api.pinarkive.com/api/v3" {
		t.Fatalf("unexpected base URL: %s", c.BaseURL)
	}
	if c.Token != "tok" {
		t.Fatalf("unexpected token: %s", c.Token)
	}
}

func TestGetMeAuthHeader(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		_ = json.NewEncoder(w).Encode(map[string]string{"ok": "true"})
	}))
	defer srv.Close()

	c := NewClient("abc", "", srv.URL)
	if _, err := c.GetMe(); err != nil {
		t.Fatal(err)
	}
	if gotAuth != "Bearer abc" {
		t.Fatalf("auth header = %q", gotAuth)
	}
}

func TestAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error":   "Unauthorized",
			"message": "bad",
			"code":    "unauthorized",
		})
	}))
	defer srv.Close()

	c := NewClient("x", "", srv.URL)
	_, err := c.GetMe()
	ae, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if ae.StatusCode != 401 || ae.Code != "unauthorized" {
		t.Fatalf("unexpected error: %#v", ae)
	}
}

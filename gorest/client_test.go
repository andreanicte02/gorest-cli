package gorest

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestClient(t *testing.T, h http.HandlerFunc) *Client {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	c := New("secret")
	c.BaseURL = srv.URL
	return c
}

func TestListUsers(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/users" {
			t.Errorf("got %s %s", r.Method, r.URL.Path)
		}
		if got := r.URL.Query().Get("page"); got != "2" {
			t.Errorf("page = %q, want 2", got)
		}
		if got := r.URL.Query().Get("per_page"); got != "5" {
			t.Errorf("per_page = %q, want 5", got)
		}
		json.NewEncoder(w).Encode([]User{{ID: 1, Name: "Ada"}, {ID: 2, Name: "Grace"}})
	})

	users, err := c.ListUsers(context.Background(), 2, 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(users) != 2 || users[1].Name != "Grace" {
		t.Fatalf("unexpected users: %+v", users)
	}
}

func TestCreateUserSendsTokenAndBody(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer secret" {
			t.Errorf("Authorization = %q", got)
		}
		var u User
		if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
			t.Fatal(err)
		}
		u.ID = 42
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(u)
	})

	u, err := c.CreateUser(context.Background(), User{Name: "Ada", Email: "ada@example.com", Gender: "female", Status: "active"})
	if err != nil {
		t.Fatal(err)
	}
	if u.ID != 42 || u.Email != "ada@example.com" {
		t.Fatalf("unexpected user: %+v", u)
	}
}

func TestUpdateUserOmitsEmptyFields(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch || r.URL.Path != "/users/7" {
			t.Errorf("got %s %s", r.Method, r.URL.Path)
		}
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		if len(body) != 1 || body["status"] != "inactive" {
			t.Errorf("body = %v, want only status", body)
		}
		json.NewEncoder(w).Encode(User{ID: 7, Status: "inactive"})
	})

	if _, err := c.UpdateUser(context.Background(), 7, User{Status: "inactive"}); err != nil {
		t.Fatal(err)
	}
}

func TestDeleteUserNoContent(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/users/7" {
			t.Errorf("got %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	if err := c.DeleteUser(context.Background(), 7); err != nil {
		t.Fatal(err)
	}
}

func TestErrors(t *testing.T) {
	tests := []struct {
		name       string
		status     int
		body       string
		wantFields int
		wantMsg    string
	}{
		{"not found", 404, `{"message":"Resource not found"}`, 0, "gorest: 404: Resource not found"},
		{"validation", 422, `[{"field":"email","message":"has already been taken"}]`, 1, "gorest: 422: email has already been taken"},
		{"empty body", 401, ``, 0, "gorest: 401: Unauthorized"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.status)
				w.Write([]byte(tt.body))
			})

			_, err := c.GetUser(context.Background(), 1)
			var apiErr *APIError
			if !errors.As(err, &apiErr) {
				t.Fatalf("err = %v, want *APIError", err)
			}
			if apiErr.StatusCode != tt.status || len(apiErr.Fields) != tt.wantFields {
				t.Errorf("got %+v", apiErr)
			}
			if err.Error() != tt.wantMsg {
				t.Errorf("Error() = %q, want %q", err.Error(), tt.wantMsg)
			}
		})
	}
}

func TestNetworkError(t *testing.T) {
	c := New("")
	c.BaseURL = "http://127.0.0.1:0"
	if _, err := c.GetUser(context.Background(), 1); err == nil {
		t.Fatal("expected error, got nil")
	}
}

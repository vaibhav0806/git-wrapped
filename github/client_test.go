// github/client_test.go
package github

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchUser(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/users/octocat" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(User{Login: "octocat", Name: "The Octocat"})
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "")
	user, err := c.FetchUser("octocat")
	if err != nil {
		t.Fatalf("FetchUser: %v", err)
	}
	if user.Login != "octocat" {
		t.Errorf("got login %q, want octocat", user.Login)
	}
}

func TestFetchEvents(t *testing.T) {
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			events := []Event{{Type: "PushEvent", Repo: EventRepo{Name: "octocat/hello"}}}
			w.Header().Set("Link", `<`+r.URL.Path+`?page=2>; rel="next"`)
			json.NewEncoder(w).Encode(events)
		} else {
			json.NewEncoder(w).Encode([]Event{})
		}
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "")
	events, err := c.FetchEvents("octocat")
	if err != nil {
		t.Fatalf("FetchEvents: %v", err)
	}
	if len(events) != 1 {
		t.Errorf("got %d events, want 1", len(events))
	}
}

func TestFetchRepos(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		repos := []Repo{{Name: "hello", Language: "Go", StargazersCount: 42}}
		json.NewEncoder(w).Encode(repos)
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "")
	repos, err := c.FetchRepos("octocat")
	if err != nil {
		t.Fatalf("FetchRepos: %v", err)
	}
	if len(repos) != 1 || repos[0].StargazersCount != 42 {
		t.Errorf("unexpected repos: %+v", repos)
	}
}

func TestFetchUserNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "")
	_, err := c.FetchUser("nobody")
	if err == nil {
		t.Fatal("expected error for 404")
	}
}

func TestFetchRateLimited(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "")
	_, err := c.FetchUser("octocat")
	if err == nil {
		t.Fatal("expected error for 403")
	}
}

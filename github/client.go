// github/client.go
package github

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type Client struct {
	baseURL string
	token   string
	http    *http.Client
}

func NewClient(baseURL, token string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		token:   token,
		http:    &http.Client{},
	}
}

func (c *Client) get(path string) (*http.Response, error) {
	req, err := http.NewRequest("GET", c.baseURL+path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusNotFound {
		resp.Body.Close()
		return nil, fmt.Errorf("not found")
	}
	if resp.StatusCode == http.StatusForbidden || resp.StatusCode == 429 {
		resp.Body.Close()
		return nil, fmt.Errorf("rate limited")
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
	return resp, nil
}

func (c *Client) FetchUser(username string) (User, error) {
	resp, err := c.get("/users/" + username)
	if err != nil {
		return User{}, err
	}
	defer resp.Body.Close()
	var u User
	return u, json.NewDecoder(resp.Body).Decode(&u)
}

func (c *Client) FetchEvents(username string) ([]Event, error) {
	var all []Event
	for page := 1; page <= 3; page++ {
		path := fmt.Sprintf("/users/%s/events/public?per_page=100&page=%d", username, page)
		resp, err := c.get(path)
		if err != nil {
			return nil, err
		}
		var events []Event
		err = json.NewDecoder(resp.Body).Decode(&events)
		resp.Body.Close()
		if err != nil {
			return nil, err
		}
		all = append(all, events...)
		if len(events) == 0 {
			break
		}
		// Check for next page via Link header
		link := resp.Header.Get("Link")
		if !strings.Contains(link, `rel="next"`) {
			break
		}
	}
	return all, nil
}

func (c *Client) FetchRepos(username string) ([]Repo, error) {
	var all []Repo
	for page := 1; ; page++ {
		path := fmt.Sprintf("/users/%s/repos?per_page=100&type=owner&page=%d", username, page)
		resp, err := c.get(path)
		if err != nil {
			return nil, err
		}
		var repos []Repo
		err = json.NewDecoder(resp.Body).Decode(&repos)
		resp.Body.Close()
		if err != nil {
			return nil, err
		}
		all = append(all, repos...)
		if len(repos) < 100 {
			break
		}
	}
	return all, nil
}

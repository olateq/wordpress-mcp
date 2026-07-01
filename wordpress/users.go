package wordpress

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// User represents a WordPress user.
type User struct {
	ID             int           `json:"id"`
	Username       string        `json:"username,omitempty"`
	Name           string        `json:"name,omitempty"`
	FirstName      string        `json:"first_name,omitempty"`
	LastName       string        `json:"last_name,omitempty"`
	Email          string        `json:"email,omitempty"`
	URL            string        `json:"url,omitempty"`
	Description    string        `json:"description,omitempty"`
	Link           string        `json:"link,omitempty"`
	Locale         string        `json:"locale,omitempty"`
	Nickname       string        `json:"nickname,omitempty"`
	Slug           string        `json:"slug,omitempty"`
	RegisteredDate string        `json:"registered_date,omitempty"`
	Roles          []string      `json:"roles,omitempty"`
	Capabilities   interface{}   `json:"capabilities,omitempty"`
	AvatarURLs     interface{}    `json:"avatar_urls,omitempty"`
	Meta           interface{}   `json:"meta,omitempty"`
}

// UserInput is used when creating/updating a user.
type UserInput struct {
	Username    string `json:"username,omitempty"`
	Name        string `json:"name,omitempty"`
	FirstName   string `json:"first_name,omitempty"`
	LastName    string `json:"last_name,omitempty"`
	Email       string `json:"email,omitempty"`
	URL         string `json:"url,omitempty"`
	Description string `json:"description,omitempty"`
	Locale      string `json:"locale,omitempty"`
	Nickname    string `json:"nickname,omitempty"`
	Slug        string `json:"slug,omitempty"`
	Roles       []string `json:"roles,omitempty"`
	Password    string `json:"password,omitempty"`
}

// ListUsers retrieves a list of users.
func (c *Client) ListUsers(params url.Values) ([]User, error) {
	u := c.apiURL("users")
	if params != nil {
		u += "?" + params.Encode()
	}
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}
	c.setAuth(req)
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}
	var users []User
	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		return nil, err
	}
	return users, nil
}

// GetUser retrieves a single user by ID.
func (c *Client) GetUser(id int) (*User, error) {
	u := c.apiURL(fmt.Sprintf("users/%d", id))
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}
	c.setAuth(req)
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}
	var user User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}
	return &user, nil
}

// CreateUser creates a new user.
func (c *Client) CreateUser(input UserInput) (*User, error) {
	body, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}
	u := c.apiURL("users")
	req, err := http.NewRequest("POST", u, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	c.setAuth(req)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(respBody))
	}
	var user User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateUser updates an existing user by ID.
func (c *Client) UpdateUser(id int, input UserInput) (*User, error) {
	body, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}
	u := c.apiURL(fmt.Sprintf("users/%d", id))
	req, err := http.NewRequest("POST", u, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	c.setAuth(req)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(respBody))
	}
	var user User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}
	return &user, nil
}

// DeleteUser deletes a user by ID.
func (c *Client) DeleteUser(id int, reassign int) error {
	u := c.apiURL(fmt.Sprintf("users/%d", id))
	if reassign > 0 {
		u += fmt.Sprintf("?reassign=%d", reassign)
	}
	req, err := http.NewRequest("DELETE", u, nil)
	if err != nil {
		return err
	}
	c.setAuth(req)
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}
	return nil
}
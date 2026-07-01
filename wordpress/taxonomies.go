package wordpress

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// Term represents a WordPress taxonomy term (category or tag).
type Term struct {
	ID          int      `json:"id"`
	Count       int      `json:"count,omitempty"`
	Description string   `json:"description,omitempty"`
	Link        string   `json:"link,omitempty"`
	Name        string   `json:"name,omitempty"`
	Slug        string   `json:"slug,omitempty"`
	Taxonomy    string   `json:"taxonomy,omitempty"`
	Parent      int      `json:"parent,omitempty"`
}

// TermInput is used when creating/updating a term.
type TermInput struct {
	Name        string `json:"name,omitempty"`
	Slug        string `json:"slug,omitempty"`
	Description string `json:"description,omitempty"`
	Parent      int    `json:"parent,omitempty"`
}

// ListCategories retrieves a list of categories.
func (c *Client) ListCategories(params url.Values) ([]Term, error) {
	u := c.apiURL("categories")
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
	var terms []Term
	if err := json.NewDecoder(resp.Body).Decode(&terms); err != nil {
		return nil, err
	}
	return terms, nil
}

// GetCategory retrieves a single category by ID.
func (c *Client) GetCategory(id int) (*Term, error) {
	u := c.apiURL(fmt.Sprintf("categories/%d", id))
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
	var term Term
	if err := json.NewDecoder(resp.Body).Decode(&term); err != nil {
		return nil, err
	}
	return &term, nil
}

// CreateCategory creates a new category.
func (c *Client) CreateCategory(input TermInput) (*Term, error) {
	body, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}
	u := c.apiURL("categories")
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
	var term Term
	if err := json.NewDecoder(resp.Body).Decode(&term); err != nil {
		return nil, err
	}
	return &term, nil
}

// UpdateCategory updates a category by ID.
func (c *Client) UpdateCategory(id int, input TermInput) (*Term, error) {
	body, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}
	u := c.apiURL(fmt.Sprintf("categories/%d", id))
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
	var term Term
	if err := json.NewDecoder(resp.Body).Decode(&term); err != nil {
		return nil, err
	}
	return &term, nil
}

// DeleteCategory deletes a category by ID.
func (c *Client) DeleteCategory(id int, force bool) error {
	u := c.apiURL(fmt.Sprintf("categories/%d", id))
	if force {
		u += "?force=true"
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

// ── Tags ─────────────────────────────────────────────────────────────────────

// ListTags retrieves a list of tags.
func (c *Client) ListTags(params url.Values) ([]Term, error) {
	u := c.apiURL("tags")
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
	var terms []Term
	if err := json.NewDecoder(resp.Body).Decode(&terms); err != nil {
		return nil, err
	}
	return terms, nil
}

// GetTag retrieves a single tag by ID.
func (c *Client) GetTag(id int) (*Term, error) {
	u := c.apiURL(fmt.Sprintf("tags/%d", id))
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
	var term Term
	if err := json.NewDecoder(resp.Body).Decode(&term); err != nil {
		return nil, err
	}
	return &term, nil
}

// CreateTag creates a new tag.
func (c *Client) CreateTag(input TermInput) (*Term, error) {
	body, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}
	u := c.apiURL("tags")
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
	var term Term
	if err := json.NewDecoder(resp.Body).Decode(&term); err != nil {
		return nil, err
	}
	return &term, nil
}

// UpdateTag updates a tag by ID.
func (c *Client) UpdateTag(id int, input TermInput) (*Term, error) {
	body, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}
	u := c.apiURL(fmt.Sprintf("tags/%d", id))
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
	var term Term
	if err := json.NewDecoder(resp.Body).Decode(&term); err != nil {
		return nil, err
	}
	return &term, nil
}

// DeleteTag deletes a tag by ID.
func (c *Client) DeleteTag(id int, force bool) error {
	u := c.apiURL(fmt.Sprintf("tags/%d", id))
	if force {
		u += "?force=true"
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
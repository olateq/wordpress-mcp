package wordpress

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// Page represents a WordPress page.
type Page struct {
	ID            int      `json:"id"`
	Date          string   `json:"date,omitempty"`
	DateGmt       string   `json:"date_gmt,omitempty"`
	Modified      string   `json:"modified,omitempty"`
	ModifiedGmt   string   `json:"modified_gmt,omitempty"`
	GUID          GUID     `json:"guid,omitempty"`
	Slug          string   `json:"slug,omitempty"`
	Link          string   `json:"link,omitempty"`
	Title         Rendered `json:"title"`
	Content       Rendered `json:"content"`
	Excerpt       Rendered `json:"excerpt,omitempty"`
	Author        int      `json:"author,omitempty"`
	FeaturedMedia int      `json:"featured_media,omitempty"`
	CommentStatus string   `json:"comment_status,omitempty"`
	PingStatus    string   `json:"ping_status,omitempty"`
	Status        string   `json:"status,omitempty"`
	Type          string   `json:"type,omitempty"`
	Template      string   `json:"template,omitempty"`
	Parent        int      `json:"parent,omitempty"`
	MenuOrder     int      `json:"menu_order,omitempty"`
}

// PageInput is used when creating/updating a page.
type PageInput struct {
	Title          string `json:"title,omitempty"`
	Content        string `json:"content,omitempty"`
	Excerpt        string `json:"excerpt,omitempty"`
	Slug           string `json:"slug,omitempty"`
	Status         string `json:"status,omitempty"`
	Author         int    `json:"author,omitempty"`
	FeaturedMedia  int    `json:"featured_media,omitempty"`
	CommentStatus  string `json:"comment_status,omitempty"`
	PingStatus     string `json:"ping_status,omitempty"`
	Template       string `json:"template,omitempty"`
	Parent         int    `json:"parent,omitempty"`
	MenuOrder      int    `json:"menu_order,omitempty"`
}

// ListPages retrieves a list of pages with optional query parameters.
func (c *Client) ListPages(params url.Values) ([]Page, error) {
	u := c.apiURL("pages")
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
	var pages []Page
	if err := json.NewDecoder(resp.Body).Decode(&pages); err != nil {
		return nil, err
	}
	return pages, nil
}

// GetPage retrieves a single page by ID.
func (c *Client) GetPage(id int) (*Page, error) {
	u := c.apiURL(fmt.Sprintf("pages/%d", id))
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
	var page Page
	if err := json.NewDecoder(resp.Body).Decode(&page); err != nil {
		return nil, err
	}
	return &page, nil
}

// CreatePage creates a new page.
func (c *Client) CreatePage(input PageInput) (*Page, error) {
	body, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}
	u := c.apiURL("pages")
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
	var page Page
	if err := json.NewDecoder(resp.Body).Decode(&page); err != nil {
		return nil, err
	}
	return &page, nil
}

// UpdatePage updates an existing page by ID.
func (c *Client) UpdatePage(id int, input PageInput) (*Page, error) {
	body, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}
	u := c.apiURL(fmt.Sprintf("pages/%d", id))
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
	var page Page
	if err := json.NewDecoder(resp.Body).Decode(&page); err != nil {
		return nil, err
	}
	return &page, nil
}

// DeletePage deletes a page by ID.
func (c *Client) DeletePage(id int, force bool) error {
	u := c.apiURL(fmt.Sprintf("pages/%d", id))
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
package wordpress

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client is the WordPress REST API client.
type Client struct {
	BaseURL  string
	Username string
	Password string // Application password
	HTTP     *http.Client
}

// NewClient creates a new WordPress REST API client.
func NewClient(baseURL, username, password string) *Client {
	baseURL = strings.TrimRight(baseURL, "/")
	return &Client{
		BaseURL:  baseURL,
		Username: username,
		Password: password,
		HTTP: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// apiURL builds the full API URL for a given path.
func (c *Client) apiURL(path string) string {
	return fmt.Sprintf("%s/wp-json/wp/v2/%s", c.BaseURL, strings.TrimLeft(path, "/"))
}

// setAuth sets Basic Authentication header.
func (c *Client) setAuth(req *http.Request) {
	req.SetBasicAuth(c.Username, c.Password)
}

// Rendered represents a WordPress rendered field (title, content, excerpt).
type Rendered struct {
	Rendered string `json:"rendered"`
	Raw      string `json:"raw,omitempty"`
}

// GUID represents the globally unique identifier object.
type GUID struct {
	Rendered string `json:"rendered"`
}

// ── Posts ──────────────────────────────────────────────────────────────────

type Post struct {
	ID             int      `json:"id"`
	Date           string   `json:"date,omitempty"`
	DateGmt        string   `json:"date_gmt,omitempty"`
	GUID           GUID     `json:"guid,omitempty"`
	Modified       string   `json:"modified,omitempty"`
	ModifiedGmt    string   `json:"modified_gmt,omitempty"`
	Slug           string   `json:"slug,omitempty"`
	Link           string   `json:"link,omitempty"`
	Title          Rendered `json:"title"`
	Content        Rendered `json:"content"`
	Excerpt        Rendered `json:"excerpt,omitempty"`
	Author         int      `json:"author,omitempty"`
	FeaturedMedia  int      `json:"featured_media,omitempty"`
	CommentStatus  string   `json:"comment_status,omitempty"`
	PingStatus     string   `json:"ping_status,omitempty"`
	Status         string   `json:"status,omitempty"`
	Type           string   `json:"type,omitempty"`
	Format         string   `json:"format,omitempty"`
	Sticky         bool     `json:"sticky,omitempty"`
	Categories     []int    `json:"categories,omitempty"`
	Tags           []int    `json:"tags,omitempty"`
	Template       string   `json:"template,omitempty"`
}

// PostInput is used when creating/updating a post.
type PostInput struct {
	Title          string `json:"title,omitempty"`
	Content        string `json:"content,omitempty"`
	Excerpt        string `json:"excerpt,omitempty"`
	Slug           string `json:"slug,omitempty"`
	Status         string `json:"status,omitempty"`
	Author         int    `json:"author,omitempty"`
	FeaturedMedia  int    `json:"featured_media,omitempty"`
	CommentStatus  string `json:"comment_status,omitempty"`
	PingStatus     string `json:"ping_status,omitempty"`
	Format         string `json:"format,omitempty"`
	Sticky         bool   `json:"sticky,omitempty"`
	Categories     []int  `json:"categories,omitempty"`
	Tags           []int  `json:"tags,omitempty"`
}

// ListPosts retrieves a list of posts with optional query parameters.
func (c *Client) ListPosts(params url.Values) ([]Post, error) {
	u := c.apiURL("posts")
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
	var posts []Post
	if err := json.NewDecoder(resp.Body).Decode(&posts); err != nil {
		return nil, err
	}
	return posts, nil
}

// GetPost retrieves a single post by ID.
func (c *Client) GetPost(id int) (*Post, error) {
	u := c.apiURL(fmt.Sprintf("posts/%d", id))
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
	var post Post
	if err := json.NewDecoder(resp.Body).Decode(&post); err != nil {
		return nil, err
	}
	return &post, nil
}

// CreatePost creates a new post.
func (c *Client) CreatePost(input PostInput) (*Post, error) {
	body, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}
	u := c.apiURL("posts")
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
	var post Post
	if err := json.NewDecoder(resp.Body).Decode(&post); err != nil {
		return nil, err
	}
	return &post, nil
}

// UpdatePost updates an existing post by ID.
func (c *Client) UpdatePost(id int, input PostInput) (*Post, error) {
	body, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}
	u := c.apiURL(fmt.Sprintf("posts/%d", id))
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
	var post Post
	if err := json.NewDecoder(resp.Body).Decode(&post); err != nil {
		return nil, err
	}
	return &post, nil
}

// DeletePost deletes a post by ID.
func (c *Client) DeletePost(id int, force bool) error {
	u := c.apiURL(fmt.Sprintf("posts/%d", id))
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
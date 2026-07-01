package wordpress

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// Comment represents a WordPress comment.
type Comment struct {
	ID             int      `json:"id"`
	Author         int      `json:"author,omitempty"`
	AuthorEmail    string   `json:"author_email,omitempty"`
	AuthorIP       string   `json:"author_ip,omitempty"`
	AuthorName     string   `json:"author_name,omitempty"`
	AuthorURL      string   `json:"author_url,omitempty"`
	AuthorUserAgent string  `json:"author_user_agent,omitempty"`
	Content        Rendered `json:"content,omitempty"`
	Date           string   `json:"date,omitempty"`
	DateGmt        string   `json:"date_gmt,omitempty"`
	Link           string   `json:"link,omitempty"`
	Parent         int      `json:"parent,omitempty"`
	Post           int      `json:"post,omitempty"`
	Status         string   `json:"status,omitempty"`
	Type           string   `json:"type,omitempty"`
}

// CommentInput is used when creating/updating a comment.
type CommentInput struct {
	Author         int    `json:"author,omitempty"`
	AuthorEmail    string `json:"author_email,omitempty"`
	AuthorName     string `json:"author_name,omitempty"`
	AuthorURL      string `json:"author_url,omitempty"`
	Content        string `json:"content,omitempty"`
	Parent         int    `json:"parent,omitempty"`
	Post           int    `json:"post,omitempty"`
	Status         string `json:"status,omitempty"`
}

// ListComments retrieves a list of comments.
func (c *Client) ListComments(params url.Values) ([]Comment, error) {
	u := c.apiURL("comments")
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
	var comments []Comment
	if err := json.NewDecoder(resp.Body).Decode(&comments); err != nil {
		return nil, err
	}
	return comments, nil
}

// GetComment retrieves a single comment by ID.
func (c *Client) GetComment(id int) (*Comment, error) {
	u := c.apiURL(fmt.Sprintf("comments/%d", id))
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
	var comment Comment
	if err := json.NewDecoder(resp.Body).Decode(&comment); err != nil {
		return nil, err
	}
	return &comment, nil
}

// CreateComment creates a new comment.
func (c *Client) CreateComment(input CommentInput) (*Comment, error) {
	body, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}
	u := c.apiURL("comments")
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
	var comment Comment
	if err := json.NewDecoder(resp.Body).Decode(&comment); err != nil {
		return nil, err
	}
	return &comment, nil
}

// UpdateComment updates a comment by ID.
func (c *Client) UpdateComment(id int, input CommentInput) (*Comment, error) {
	body, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}
	u := c.apiURL(fmt.Sprintf("comments/%d", id))
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
	var comment Comment
	if err := json.NewDecoder(resp.Body).Decode(&comment); err != nil {
		return nil, err
	}
	return &comment, nil
}

// DeleteComment deletes a comment by ID.
func (c *Client) DeleteComment(id int, force bool) error {
	u := c.apiURL(fmt.Sprintf("comments/%d", id))
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
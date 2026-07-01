package wordpress

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

// Media represents a WordPress media item.
type Media struct {
	ID          int      `json:"id"`
	Date        string   `json:"date,omitempty"`
	DateGmt     string   `json:"date_gmt,omitempty"`
	GUID        GUID     `json:"guid,omitempty"`
	Modified    string   `json:"modified,omitempty"`
	ModifiedGmt string   `json:"modified_gmt,omitempty"`
	Slug        string   `json:"slug,omitempty"`
	Link        string   `json:"link,omitempty"`
	Title       Rendered `json:"title"`
	Author      int      `json:"author,omitempty"`
	CommentStatus string `json:"comment_status,omitempty"`
	PingStatus    string `json:"ping_status,omitempty"`
	AltText     string   `json:"alt_text,omitempty"`
	Caption     Rendered `json:"caption,omitempty"`
	Description Rendered `json:"description,omitempty"`
	MediaType   string   `json:"media_type,omitempty"`
	MimeType    string   `json:"mime_type,omitempty"`
	MediaDetails interface{} `json:"media_details,omitempty"`
	Post        int      `json:"post,omitempty"`
	SourceURL   string   `json:"source_url,omitempty"`
}

// MediaInput is used when updating a media item.
type MediaInput struct {
	Title       string `json:"title,omitempty"`
	AltText     string `json:"alt_text,omitempty"`
	Caption     string `json:"caption,omitempty"`
	Description string `json:"description,omitempty"`
	Post        int    `json:"post,omitempty"`
}

// ListMedia retrieves a list of media items.
func (c *Client) ListMedia(params url.Values) ([]Media, error) {
	u := c.apiURL("media")
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
	var items []Media
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		return nil, err
	}
	return items, nil
}

// GetMedia retrieves a single media item by ID.
func (c *Client) GetMedia(id int) (*Media, error) {
	u := c.apiURL(fmt.Sprintf("media/%d", id))
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
	var m Media
	if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
		return nil, err
	}
	return &m, nil
}

// UploadMedia uploads a file as a new media item.
// filePath is the local path to the file to upload.
func (c *Client) UploadMedia(filePath string) (*Media, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	fw, err := w.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(fw, f); err != nil {
		return nil, err
	}
	w.Close()

	u := c.apiURL("media")
	req, err := http.NewRequest("POST", u, &buf)
	if err != nil {
		return nil, err
	}
	c.setAuth(req)
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filepath.Base(filePath)))

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}
	var m Media
	if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
		return nil, err
	}
	return &m, nil
}

// UpdateMedia updates a media item by ID.
func (c *Client) UpdateMedia(id int, input MediaInput) (*Media, error) {
	body, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}
	u := c.apiURL(fmt.Sprintf("media/%d", id))
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
	var m Media
	if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
		return nil, err
	}
	return &m, nil
}

// DeleteMedia deletes a media item by ID.
func (c *Client) DeleteMedia(id int, force bool) error {
	u := c.apiURL(fmt.Sprintf("media/%d", id))
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
package wordpress

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Settings represents WordPress site settings.
type Settings struct {
	Title                string `json:"title,omitempty"`
	Description          string `json:"description,omitempty"`
	URL                  string `json:"url,omitempty"`
	Email                string `json:"email,omitempty"`
	Timezone             string `json:"timezone,omitempty"`
	DateFormat           string `json:"date_format,omitempty"`
	TimeFormat           string `json:"time_format,omitempty"`
	StartOfWeek          int    `json:"start_of_week,omitempty"`
	Language             string `json:"language,omitempty"`
	UseSmilies           bool   `json:"use_smilies,omitempty"`
	DefaultCategory      int    `json:"default_category,omitempty"`
	DefaultPostFormat    string `json:"default_post_format,omitempty"`
	PostsPerPage         int    `json:"posts_per_page,omitempty"`
	DefaultPingStatus    string `json:"default_ping_status,omitempty"`
	DefaultCommentStatus string `json:"default_comment_status,omitempty"`
}

// SettingsInput is used when updating settings (all fields optional).
type SettingsInput = Settings

// GetSettings retrieves the site settings.
func (c *Client) GetSettings() (*Settings, error) {
	u := c.apiURL("settings")
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
	var s Settings
	if err := json.NewDecoder(resp.Body).Decode(&s); err != nil {
		return nil, err
	}
	return &s, nil
}

// UpdateSettings updates site settings.
func (c *Client) UpdateSettings(input SettingsInput) (*Settings, error) {
	body, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}
	u := c.apiURL("settings")
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
	var s Settings
	if err := json.NewDecoder(resp.Body).Decode(&s); err != nil {
		return nil, err
	}
	return &s, nil
}
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"

	wp "github.com/olateq/wordpress-mcp/wordpress"
)

// ── MCP JSON-RPC types ──────────────────────────────────────────────────────

type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *RPCError   `json:"error,omitempty"`
}

type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type Tool struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	InputSchema Schema `json:"inputSchema"`
}

type Schema struct {
	Type       string             `json:"type"`
	Properties map[string]PropDef `json:"properties"`
	Required   []string           `json:"required,omitempty"`
}

type PropDef struct {
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Default     interface{} `json:"default,omitempty"`
	Enum        []string `json:"enum,omitempty"`
}

type ToolCallParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

type ToolResult struct {
	Content []ContentBlock `json:"content"`
}

type ContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// ── Tool definitions ─────────────────────────────────────────────────────────

func toolsList() []Tool {
	return []Tool{
		{Name: "wp_list_posts", Description: "List WordPress posts. Optional filtering by page, per_page, search, status, author, categories, tags.",
			InputSchema: Schema{Type: "object",
				Properties: map[string]PropDef{
					"page":       {Type: "integer", Description: "Page number (default 1)"},
					"per_page":   {Type: "integer", Description: "Items per page (default 10, max 100)"},
					"search":     {Type: "string", Description: "Search term"},
					"status":     {Type: "string", Description: "Post status", Enum: []string{"publish", "future", "draft", "pending", "private"}},
					"author":     {Type: "integer", Description: "Author ID"},
					"categories": {Type: "string", Description: "Category IDs (comma-separated)"},
					"tags":       {Type: "string", Description: "Tag IDs (comma-separated)"},
					"orderby":    {Type: "string", Description: "Sort by", Enum: []string{"date", "id", "author", "modified", "slug", "title"}},
					"order":      {Type: "string", Description: "Sort order", Enum: []string{"asc", "desc"}},
				}}},
		{Name: "wp_get_post", Description: "Retrieve a single WordPress post by ID.",
			InputSchema: Schema{Type: "object",
				Properties: map[string]PropDef{
					"id": {Type: "integer", Description: "Post ID"},
				},
				Required: []string{"id"}}},
		{Name: "wp_create_post", Description: "Create a new WordPress post.",
			InputSchema: Schema{Type: "object",
				Properties: map[string]PropDef{
					"title":   {Type: "string", Description: "Post title"},
					"content": {Type: "string", Description: "Post content (HTML)"},
					"excerpt": {Type: "string", Description: "Post excerpt"},
					"slug":    {Type: "string", Description: "Post slug"},
					"status":  {Type: "string", Description: "Post status", Enum: []string{"publish", "future", "draft", "pending", "private"}},
					"author":  {Type: "integer", Description: "Author ID"},
					"categories": {Type: "string", Description: "Category IDs (comma-separated)"},
					"tags":    {Type: "string", Description: "Tag IDs (comma-separated)"},
					"featured_media": {Type: "integer", Description: "Featured media ID"},
				},
				Required: []string{"title", "content"}}},
		{Name: "wp_update_post", Description: "Update an existing WordPress post by ID.",
			InputSchema: Schema{Type: "object",
				Properties: map[string]PropDef{
					"id":      {Type: "integer", Description: "Post ID"},
					"title":   {Type: "string", Description: "Post title"},
					"content": {Type: "string", Description: "Post content (HTML)"},
					"excerpt": {Type: "string", Description: "Post excerpt"},
					"slug":    {Type: "string", Description: "Post slug"},
					"status":  {Type: "string", Description: "Post status", Enum: []string{"publish", "future", "draft", "pending", "private"}},
					"author":  {Type: "integer", Description: "Author ID"},
					"categories": {Type: "string", Description: "Category IDs (comma-separated)"},
					"tags":    {Type: "string", Description: "Tag IDs (comma-separated)"},
					"featured_media": {Type: "integer", Description: "Featured media ID"},
				},
				Required: []string{"id"}}},
		{Name: "wp_delete_post", Description: "Delete a WordPress post by ID.",
			InputSchema: Schema{Type: "object",
				Properties: map[string]PropDef{
					"id":    {Type: "integer", Description: "Post ID"},
					"force": {Type: "boolean", Description: "Skip trash and permanently delete"},
				},
				Required: []string{"id"}}},
		{Name: "wp_list_pages", Description: "List WordPress pages.",
			InputSchema: Schema{Type: "object",
				Properties: map[string]PropDef{
					"page":     {Type: "integer", Description: "Page number"},
					"per_page": {Type: "integer", Description: "Items per page"},
					"search":   {Type: "string", Description: "Search term"},
					"status":   {Type: "string", Description: "Page status"},
					"orderby":  {Type: "string", Description: "Sort by"},
					"order":    {Type: "string", Description: "Sort order"},
				}}},
		{Name: "wp_get_page", Description: "Retrieve a single WordPress page by ID.",
			InputSchema: Schema{Type: "object",
				Properties: map[string]PropDef{
					"id": {Type: "integer", Description: "Page ID"},
				},
				Required: []string{"id"}}},
		{Name: "wp_create_page", Description: "Create a new WordPress page.",
			InputSchema: Schema{Type: "object",
				Properties: map[string]PropDef{
					"title":   {Type: "string", Description: "Page title"},
					"content": {Type: "string", Description: "Page content (HTML)"},
					"slug":    {Type: "string", Description: "Page slug"},
					"status":  {Type: "string", Description: "Page status", Enum: []string{"publish", "draft", "pending", "private"}},
					"parent":  {Type: "integer", Description: "Parent page ID"},
				},
				Required: []string{"title", "content"}}},
		{Name: "wp_update_page", Description: "Update an existing WordPress page by ID.",
			InputSchema: Schema{Type: "object",
				Properties: map[string]PropDef{
					"id":      {Type: "integer", Description: "Page ID"},
					"title":   {Type: "string", Description: "Page title"},
					"content": {Type: "string", Description: "Page content (HTML)"},
					"slug":    {Type: "string", Description: "Page slug"},
					"status":  {Type: "string", Description: "Page status"},
					"parent":  {Type: "integer", Description: "Parent page ID"},
				},
				Required: []string{"id"}}},
		{Name: "wp_delete_page", Description: "Delete a WordPress page by ID.",
			InputSchema: Schema{Type: "object",
				Properties: map[string]PropDef{
					"id":    {Type: "integer", Description: "Page ID"},
					"force": {Type: "boolean", Description: "Skip trash"},
				},
				Required: []string{"id"}}},
		{Name: "wp_list_media", Description: "List WordPress media items.",
			InputSchema: Schema{Type: "object",
				Properties: map[string]PropDef{
					"page":     {Type: "integer", Description: "Page number"},
					"per_page": {Type: "integer", Description: "Items per page"},
					"search":   {Type: "string", Description: "Search term"},
					"author":   {Type: "integer", Description: "Author ID"},
				}}},
		{Name: "wp_get_media", Description: "Retrieve a single media item by ID.",
			InputSchema: Schema{Type: "object",
				Properties: map[string]PropDef{
					"id": {Type: "integer", Description: "Media ID"},
				},
				Required: []string{"id"}}},
		{Name: "wp_upload_media", Description: "Upload a file as a new media item.",
			InputSchema: Schema{Type: "object",
				Properties: map[string]PropDef{
					"file_path": {Type: "string", Description: "Local file path to upload"},
				},
				Required: []string{"file_path"}}},
		{Name: "wp_update_media", Description: "Update a media item (alt text, caption, etc).",
			InputSchema: Schema{Type: "object",
				Properties: map[string]PropDef{
					"id":          {Type: "integer", Description: "Media ID"},
					"alt_text":    {Type: "string", Description: "Alt text"},
					"caption":     {Type: "string", Description: "Caption"},
					"description": {Type: "string", Description: "Description"},
					"title":       {Type: "string", Description: "Title"},
					"post":        {Type: "integer", Description: "Associated post ID"},
				},
				Required: []string{"id"}}},
		{Name: "wp_delete_media", Description: "Delete a media item by ID.",
			InputSchema: Schema{Type: "object",
				Properties: map[string]PropDef{
					"id":    {Type: "integer", Description: "Media ID"},
					"force": {Type: "boolean", Description: "Skip trash"},
				},
				Required: []string{"id"}}},
		{Name: "wp_list_users", Description: "List WordPress users.",
			InputSchema: Schema{Type: "object",
				Properties: map[string]PropDef{
					"page":     {Type: "integer", Description: "Page number"},
					"per_page": {Type: "integer", Description: "Items per page"},
					"search":   {Type: "string", Description: "Search term"},
					"roles":    {Type: "string", Description: "Filter by role (comma-separated)"},
				}}},
		{Name: "wp_get_user", Description: "Retrieve a single user by ID.",
			InputSchema: Schema{Type: "object",
				Properties: map[string]PropDef{
					"id": {Type: "integer", Description: "User ID"},
				},
				Required: []string{"id"}}},
		{Name: "wp_create_user", Description: "Create a new WordPress user.",
			InputSchema: Schema{Type: "object",
				Properties: map[string]PropDef{
					"username": {Type: "string", Description: "Login name"},
					"email":    {Type: "string", Description: "Email address"},
					"password": {Type: "string", Description: "Password"},
					"name":     {Type: "string", Description: "Display name"},
					"roles":    {Type: "string", Description: "Roles (comma-separated, e.g. editor,author)"},
				},
				Required: []string{"username", "email", "password"}}},
		{Name: "wp_update_user", Description: "Update an existing user by ID.",
			InputSchema: Schema{Type: "object",
				Properties: map[string]PropDef{
					"id":         {Type: "integer", Description: "User ID"},
					"name":       {Type: "string", Description: "Display name"},
					"email":      {Type: "string", Description: "Email"},
					"description": {Type: "string", Description: "Description"},
					"roles":      {Type: "string", Description: "Roles (comma-separated)"},
				},
				Required: []string{"id"}}},
		{Name: "wp_delete_user", Description: "Delete a user by ID. Requires reassign parameter.",
			InputSchema: Schema{Type: "object",
				Properties: map[string]PropDef{
					"id":       {Type: "integer", Description: "User ID to delete"},
					"reassign": {Type: "integer", Description: "User ID to reassign posts to"},
				},
				Required: []string{"id", "reassign"}}},
		{Name: "wp_list_comments", Description: "List WordPress comments.",
			InputSchema: Schema{Type: "object",
				Properties: map[string]PropDef{
					"page":     {Type: "integer", Description: "Page number"},
					"per_page": {Type: "integer", Description: "Items per page"},
					"post":     {Type: "integer", Description: "Filter by post ID"},
					"status":   {Type: "string", Description: "Comment status", Enum: []string{"approve", "hold", "spam", "trash"}},
					"search":   {Type: "string", Description: "Search term"},
				}}},
		{Name: "wp_get_comment", Description: "Retrieve a single comment by ID.",
			InputSchema: Schema{Type: "object",
				Properties: map[string]PropDef{
					"id": {Type: "integer", Description: "Comment ID"},
				},
				Required: []string{"id"}}},
		{Name: "wp_create_comment", Description: "Create a new comment on a post.",
			InputSchema: Schema{Type: "object",
				Properties: map[string]PropDef{
					"post":         {Type: "integer", Description: "Post ID to comment on"},
					"author_name":  {Type: "string", Description: "Author name"},
					"author_email": {Type: "string", Description: "Author email"},
					"content":      {Type: "string", Description: "Comment content"},
					"status":       {Type: "string", Description: "Comment status"},
				},
				Required: []string{"post", "content"}}},
		{Name: "wp_update_comment", Description: "Update a comment by ID.",
			InputSchema: Schema{Type: "object",
				Properties: map[string]PropDef{
					"id":      {Type: "integer", Description: "Comment ID"},
					"content": {Type: "string", Description: "Comment content"},
					"status":  {Type: "string", Description: "Comment status", Enum: []string{"approve", "hold", "spam", "trash"}},
				},
				Required: []string{"id"}}},
		{Name: "wp_delete_comment", Description: "Delete a comment by ID.",
			InputSchema: Schema{Type: "object",
				Properties: map[string]PropDef{
					"id":    {Type: "integer", Description: "Comment ID"},
					"force": {Type: "boolean", Description: "Skip trash"},
				},
				Required: []string{"id"}}},
		{Name: "wp_list_categories", Description: "List WordPress categories.",
			InputSchema: Schema{Type: "object",
				Properties: map[string]PropDef{
					"page":     {Type: "integer", Description: "Page number"},
					"per_page": {Type: "integer", Description: "Items per page"},
					"search":   {Type: "string", Description: "Search term"},
					"hide_empty": {Type: "boolean", Description: "Hide categories with no posts"},
				}}},
		{Name: "wp_get_category", Description: "Retrieve a single category by ID.",
			InputSchema: Schema{Type: "object",
				Properties: map[string]PropDef{
					"id": {Type: "integer", Description: "Category ID"},
				},
				Required: []string{"id"}}},
		{Name: "wp_create_category", Description: "Create a new category.",
			InputSchema: Schema{Type: "object",
				Properties: map[string]PropDef{
					"name":        {Type: "string", Description: "Category name"},
					"slug":        {Type: "string", Description: "Category slug"},
					"description": {Type: "string", Description: "Category description"},
					"parent":      {Type: "integer", Description: "Parent category ID"},
				},
				Required: []string{"name"}}},
		{Name: "wp_update_category", Description: "Update a category by ID.",
			InputSchema: Schema{Type: "object",
				Properties: map[string]PropDef{
					"id":          {Type: "integer", Description: "Category ID"},
					"name":        {Type: "string", Description: "Category name"},
					"slug":        {Type: "string", Description: "Category slug"},
					"description": {Type: "string", Description: "Category description"},
				},
				Required: []string{"id"}}},
		{Name: "wp_delete_category", Description: "Delete a category by ID.",
			InputSchema: Schema{Type: "object",
				Properties: map[string]PropDef{
					"id":    {Type: "integer", Description: "Category ID"},
					"force": {Type: "boolean", Description: "Force delete"},
				},
				Required: []string{"id"}}},
		{Name: "wp_list_tags", Description: "List WordPress tags.",
			InputSchema: Schema{Type: "object",
				Properties: map[string]PropDef{
					"page":     {Type: "integer", Description: "Page number"},
					"per_page": {Type: "integer", Description: "Items per page"},
					"search":   {Type: "string", Description: "Search term"},
				}}},
		{Name: "wp_get_tag", Description: "Retrieve a single tag by ID.",
			InputSchema: Schema{Type: "object",
				Properties: map[string]PropDef{
					"id": {Type: "integer", Description: "Tag ID"},
				},
				Required: []string{"id"}}},
		{Name: "wp_create_tag", Description: "Create a new tag.",
			InputSchema: Schema{Type: "object",
				Properties: map[string]PropDef{
					"name":        {Type: "string", Description: "Tag name"},
					"slug":        {Type: "string", Description: "Tag slug"},
					"description": {Type: "string", Description: "Tag description"},
				},
				Required: []string{"name"}}},
		{Name: "wp_update_tag", Description: "Update a tag by ID.",
			InputSchema: Schema{Type: "object",
				Properties: map[string]PropDef{
					"id":          {Type: "integer", Description: "Tag ID"},
					"name":        {Type: "string", Description: "Tag name"},
					"slug":        {Type: "string", Description: "Tag slug"},
					"description": {Type: "string", Description: "Tag description"},
				},
				Required: []string{"id"}}},
		{Name: "wp_delete_tag", Description: "Delete a tag by ID.",
			InputSchema: Schema{Type: "object",
				Properties: map[string]PropDef{
					"id": {Type: "integer", Description: "Tag ID"},
				},
				Required: []string{"id"}}},
		{Name: "wp_get_settings", Description: "Retrieve WordPress site settings.",
			InputSchema: Schema{Type: "object",
				Properties: map[string]PropDef{}}},
		{Name: "wp_update_settings", Description: "Update WordPress site settings.",
			InputSchema: Schema{Type: "object",
				Properties: map[string]PropDef{
					"title":       {Type: "string", Description: "Site title"},
					"description": {Type: "string", Description: "Site tagline"},
					"email":       {Type: "string", Description: "Admin email"},
					"language":    {Type: "string", Description: "Site language"},
				}}},
	}
}

// ── Helpers ──────────────────────────────────────────────────────────────────

func getClient() *wp.Client {
	baseURL := os.Getenv("WP_BASE_URL")
	username := os.Getenv("WP_USERNAME")
	password := os.Getenv("WP_APP_PASSWORD")
	if baseURL == "" {
		baseURL = "http://localhost"
	}
	return wp.NewClient(baseURL, username, password)
}

func toInt(v interface{}) int {
	switch t := v.(type) {
	case float64:
		return int(t)
	case int:
		return t
	case string:
		n, _ := strconv.Atoi(t)
		return n
	}
	return 0
}

func toBool(v interface{}) bool {
	switch t := v.(type) {
	case bool:
		return t
	case string:
		return t == "true" || t == "1"
	case float64:
		return t != 0
	}
	return false
}

func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%v", v)
}

func parseIntSlice(v interface{}) []int {
	s := toString(v)
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	var result []int
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if n, err := strconv.Atoi(p); err == nil {
			result = append(result, n)
		}
	}
	return result
}

func parseStringSlice(v interface{}) []string {
	s := toString(v)
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	var result []string
	for _, p := range parts {
		result = append(result, strings.TrimSpace(p))
	}
	return result
}

func buildParams(args map[string]interface{}) url.Values {
	q := url.Values{}
	for k, v := range args {
		if v == nil {
			continue
		}
		s := toString(v)
		if s != "" {
			q.Set(k, s)
		}
	}
	return q
}

func jsonText(v interface{}) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Sprintf("Error marshalling result: %v", err)
	}
	return string(b)
}

func okResult(text string) ToolResult {
	return ToolResult{Content: []ContentBlock{{Type: "text", Text: text}}}
}

func errResult(text string) ToolResult {
	return ToolResult{Content: []ContentBlock{{Type: "text", Text: "Error: " + text}}}
}

// ── Tool dispatch ────────────────────────────────────────────────────────────

func handleToolCall(params ToolCallParams) ToolResult {
	client := getClient()

	switch params.Name {
	// ── Posts ──
	case "wp_list_posts":
		posts, err := client.ListPosts(buildParams(params.Arguments))
		if err != nil {
			return errResult(err.Error())
		}
		return okResult(jsonText(posts))
	case "wp_get_post":
		post, err := client.GetPost(toInt(params.Arguments["id"]))
		if err != nil {
			return errResult(err.Error())
		}
		return okResult(jsonText(post))
	case "wp_create_post":
		input := wp.PostInput{
			Title:   toString(params.Arguments["title"]),
			Content: toString(params.Arguments["content"]),
			Excerpt: toString(params.Arguments["excerpt"]),
			Slug:    toString(params.Arguments["slug"]),
			Status:  toString(params.Arguments["status"]),
			Author:  toInt(params.Arguments["author"]),
			FeaturedMedia: toInt(params.Arguments["featured_media"]),
			Categories: parseIntSlice(params.Arguments["categories"]),
			Tags:     parseIntSlice(params.Arguments["tags"]),
		}
		post, err := client.CreatePost(input)
		if err != nil {
			return errResult(err.Error())
		}
		return okResult(jsonText(post))
	case "wp_update_post":
		input := wp.PostInput{
			Title:   toString(params.Arguments["title"]),
			Content: toString(params.Arguments["content"]),
			Excerpt: toString(params.Arguments["excerpt"]),
			Slug:    toString(params.Arguments["slug"]),
			Status:  toString(params.Arguments["status"]),
			Author:  toInt(params.Arguments["author"]),
			FeaturedMedia: toInt(params.Arguments["featured_media"]),
			Categories: parseIntSlice(params.Arguments["categories"]),
			Tags:     parseIntSlice(params.Arguments["tags"]),
		}
		post, err := client.UpdatePost(toInt(params.Arguments["id"]), input)
		if err != nil {
			return errResult(err.Error())
		}
		return okResult(jsonText(post))
	case "wp_delete_post":
		err := client.DeletePost(toInt(params.Arguments["id"]), toBool(params.Arguments["force"]))
		if err != nil {
			return errResult(err.Error())
		}
		return okResult("Post deleted successfully")

	// ── Pages ──
	case "wp_list_pages":
		pages, err := client.ListPages(buildParams(params.Arguments))
		if err != nil {
			return errResult(err.Error())
		}
		return okResult(jsonText(pages))
	case "wp_get_page":
		page, err := client.GetPage(toInt(params.Arguments["id"]))
		if err != nil {
			return errResult(err.Error())
		}
		return okResult(jsonText(page))
	case "wp_create_page":
		input := wp.PageInput{
			Title:   toString(params.Arguments["title"]),
			Content: toString(params.Arguments["content"]),
			Slug:    toString(params.Arguments["slug"]),
			Status:  toString(params.Arguments["status"]),
			Parent:  toInt(params.Arguments["parent"]),
		}
		page, err := client.CreatePage(input)
		if err != nil {
			return errResult(err.Error())
		}
		return okResult(jsonText(page))
	case "wp_update_page":
		input := wp.PageInput{
			Title:   toString(params.Arguments["title"]),
			Content: toString(params.Arguments["content"]),
			Slug:    toString(params.Arguments["slug"]),
			Status:  toString(params.Arguments["status"]),
			Parent:  toInt(params.Arguments["parent"]),
		}
		page, err := client.UpdatePage(toInt(params.Arguments["id"]), input)
		if err != nil {
			return errResult(err.Error())
		}
		return okResult(jsonText(page))
	case "wp_delete_page":
		err := client.DeletePage(toInt(params.Arguments["id"]), toBool(params.Arguments["force"]))
		if err != nil {
			return errResult(err.Error())
		}
		return okResult("Page deleted successfully")

	// ── Media ──
	case "wp_list_media":
		media, err := client.ListMedia(buildParams(params.Arguments))
		if err != nil {
			return errResult(err.Error())
		}
		return okResult(jsonText(media))
	case "wp_get_media":
		m, err := client.GetMedia(toInt(params.Arguments["id"]))
		if err != nil {
			return errResult(err.Error())
		}
		return okResult(jsonText(m))
	case "wp_upload_media":
		m, err := client.UploadMedia(toString(params.Arguments["file_path"]))
		if err != nil {
			return errResult(err.Error())
		}
		return okResult(jsonText(m))
	case "wp_update_media":
		input := wp.MediaInput{
			Title:       toString(params.Arguments["title"]),
			AltText:     toString(params.Arguments["alt_text"]),
			Caption:     toString(params.Arguments["caption"]),
			Description: toString(params.Arguments["description"]),
			Post:        toInt(params.Arguments["post"]),
		}
		m, err := client.UpdateMedia(toInt(params.Arguments["id"]), input)
		if err != nil {
			return errResult(err.Error())
		}
		return okResult(jsonText(m))
	case "wp_delete_media":
		err := client.DeleteMedia(toInt(params.Arguments["id"]), toBool(params.Arguments["force"]))
		if err != nil {
			return errResult(err.Error())
		}
		return okResult("Media deleted successfully")

	// ── Users ──
	case "wp_list_users":
		users, err := client.ListUsers(buildParams(params.Arguments))
		if err != nil {
			return errResult(err.Error())
		}
		return okResult(jsonText(users))
	case "wp_get_user":
		user, err := client.GetUser(toInt(params.Arguments["id"]))
		if err != nil {
			return errResult(err.Error())
		}
		return okResult(jsonText(user))
	case "wp_create_user":
		input := wp.UserInput{
			Username: toString(params.Arguments["username"]),
			Email:    toString(params.Arguments["email"]),
			Password: toString(params.Arguments["password"]),
			Name:     toString(params.Arguments["name"]),
			Roles:    parseStringSlice(params.Arguments["roles"]),
		}
		user, err := client.CreateUser(input)
		if err != nil {
			return errResult(err.Error())
		}
		return okResult(jsonText(user))
	case "wp_update_user":
		input := wp.UserInput{
			Name:        toString(params.Arguments["name"]),
			Email:       toString(params.Arguments["email"]),
			Description: toString(params.Arguments["description"]),
			Roles:       parseStringSlice(params.Arguments["roles"]),
		}
		user, err := client.UpdateUser(toInt(params.Arguments["id"]), input)
		if err != nil {
			return errResult(err.Error())
		}
		return okResult(jsonText(user))
	case "wp_delete_user":
		err := client.DeleteUser(toInt(params.Arguments["id"]), toInt(params.Arguments["reassign"]))
		if err != nil {
			return errResult(err.Error())
		}
		return okResult("User deleted successfully")

	// ── Comments ──
	case "wp_list_comments":
		comments, err := client.ListComments(buildParams(params.Arguments))
		if err != nil {
			return errResult(err.Error())
		}
		return okResult(jsonText(comments))
	case "wp_get_comment":
		comment, err := client.GetComment(toInt(params.Arguments["id"]))
		if err != nil {
			return errResult(err.Error())
		}
		return okResult(jsonText(comment))
	case "wp_create_comment":
		input := wp.CommentInput{
			Post:       toInt(params.Arguments["post"]),
			AuthorName: toString(params.Arguments["author_name"]),
			AuthorEmail: toString(params.Arguments["author_email"]),
			Content:    toString(params.Arguments["content"]),
			Status:     toString(params.Arguments["status"]),
		}
		comment, err := client.CreateComment(input)
		if err != nil {
			return errResult(err.Error())
		}
		return okResult(jsonText(comment))
	case "wp_update_comment":
		input := wp.CommentInput{
			Content: toString(params.Arguments["content"]),
			Status:  toString(params.Arguments["status"]),
		}
		comment, err := client.UpdateComment(toInt(params.Arguments["id"]), input)
		if err != nil {
			return errResult(err.Error())
		}
		return okResult(jsonText(comment))
	case "wp_delete_comment":
		err := client.DeleteComment(toInt(params.Arguments["id"]), toBool(params.Arguments["force"]))
		if err != nil {
			return errResult(err.Error())
		}
		return okResult("Comment deleted successfully")

	// ── Categories ──
	case "wp_list_categories":
		terms, err := client.ListCategories(buildParams(params.Arguments))
		if err != nil {
			return errResult(err.Error())
		}
		return okResult(jsonText(terms))
	case "wp_get_category":
		term, err := client.GetCategory(toInt(params.Arguments["id"]))
		if err != nil {
			return errResult(err.Error())
		}
		return okResult(jsonText(term))
	case "wp_create_category":
		input := wp.TermInput{
			Name:        toString(params.Arguments["name"]),
			Slug:        toString(params.Arguments["slug"]),
			Description: toString(params.Arguments["description"]),
			Parent:      toInt(params.Arguments["parent"]),
		}
		term, err := client.CreateCategory(input)
		if err != nil {
			return errResult(err.Error())
		}
		return okResult(jsonText(term))
	case "wp_update_category":
		input := wp.TermInput{
			Name:        toString(params.Arguments["name"]),
			Slug:        toString(params.Arguments["slug"]),
			Description: toString(params.Arguments["description"]),
		}
		term, err := client.UpdateCategory(toInt(params.Arguments["id"]), input)
		if err != nil {
			return errResult(err.Error())
		}
		return okResult(jsonText(term))
	case "wp_delete_category":
		err := client.DeleteCategory(toInt(params.Arguments["id"]), toBool(params.Arguments["force"]))
		if err != nil {
			return errResult(err.Error())
		}
		return okResult("Category deleted successfully")

	// ── Tags ──
	case "wp_list_tags":
		terms, err := client.ListTags(buildParams(params.Arguments))
		if err != nil {
			return errResult(err.Error())
		}
		return okResult(jsonText(terms))
	case "wp_get_tag":
		term, err := client.GetTag(toInt(params.Arguments["id"]))
		if err != nil {
			return errResult(err.Error())
		}
		return okResult(jsonText(term))
	case "wp_create_tag":
		input := wp.TermInput{
			Name:        toString(params.Arguments["name"]),
			Slug:        toString(params.Arguments["slug"]),
			Description: toString(params.Arguments["description"]),
		}
		term, err := client.CreateTag(input)
		if err != nil {
			return errResult(err.Error())
		}
		return okResult(jsonText(term))
	case "wp_update_tag":
		input := wp.TermInput{
			Name:        toString(params.Arguments["name"]),
			Slug:        toString(params.Arguments["slug"]),
			Description: toString(params.Arguments["description"]),
		}
		term, err := client.UpdateTag(toInt(params.Arguments["id"]), input)
		if err != nil {
			return errResult(err.Error())
		}
		return okResult(jsonText(term))
	case "wp_delete_tag":
		err := client.DeleteTag(toInt(params.Arguments["id"]), false)
		if err != nil {
			return errResult(err.Error())
		}
		return okResult("Tag deleted successfully")

	// ── Settings ──
	case "wp_get_settings":
		s, err := client.GetSettings()
		if err != nil {
			return errResult(err.Error())
		}
		return okResult(jsonText(s))
	case "wp_update_settings":
		input := wp.SettingsInput{
			Title:       toString(params.Arguments["title"]),
			Description: toString(params.Arguments["description"]),
			Email:       toString(params.Arguments["email"]),
			Language:    toString(params.Arguments["language"]),
		}
		s, err := client.UpdateSettings(input)
		if err != nil {
			return errResult(err.Error())
		}
		return okResult(jsonText(s))

	default:
		return errResult(fmt.Sprintf("Unknown tool: %s", params.Name))
	}
}

// ── MCP server (JSON-RPC over stdio) ─────────────────────────────────────────

func handleRequest(req JSONRPCRequest) JSONRPCResponse {
	switch req.Method {
	case "initialize":
		return JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: map[string]interface{}{
				"protocolVersion": "2024-11-05",
				"capabilities": map[string]interface{}{
					"tools": map[string]interface{}{},
				},
				"serverInfo": map[string]interface{}{
					"name":    "wordpress-mcp",
					"version": "1.0.0",
				},
			},
		}

	case "tools/list":
		return JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: map[string]interface{}{
				"tools": toolsList(),
			},
		}

	case "tools/call":
		var params ToolCallParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			return JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Error:   &RPCError{Code: -32602, Message: "Invalid params: " + err.Error()},
			}
		}
		result := handleToolCall(params)
		return JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  result,
		}

	case "ping":
		return JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  map[string]interface{}{},
		}

	default:
		return JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &RPCError{Code: -32601, Message: "Method not found: " + req.Method},
		}
	}
}

// ── CLI ──────────────────────────────────────────────────────────────────────

func main() {
	log.SetOutput(os.Stderr)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Parse CLI flags
	mode := "stdio"
	addr := ":8080"
	apiKey := ""

	args := os.Args[1:]
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--mode", "-m":
			if i+1 < len(args) {
				mode = args[i+1]
				i++
			}
		case "--addr", "-a":
			if i+1 < len(args) {
				addr = args[i+1]
				i++
			}
		case "--api-key", "-k":
			if i+1 < len(args) {
				apiKey = args[i+1]
				i++
			}
		case "--help", "-h":
			printUsage()
			return
		}
	}

	// API key can also come from env var
	if apiKey == "" {
		apiKey = os.Getenv("MCP_API_KEY")
	}

	switch mode {
	case "stdio":
		runStdio()
	case "sse", "http":
		runSSE(addr, apiKey)
	default:
		log.Fatalf("Unknown mode: %s (use 'stdio' or 'sse')", mode)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `
wordpress-mcp — WordPress REST API MCP Server

Usage:
  wordpress-mcp [flags]

Flags:
  --mode, -m <mode>       Transport mode: "stdio" (default) or "sse"
  --addr, -a <addr>       SSE listen address (default ":8080")
  --api-key, -k <key>     Bearer API key for SSE mode (or set MCP_API_KEY env var)
  --help, -h              Show this help

Environment variables:
  WP_BASE_URL             WordPress site URL (e.g. https://your-site.com)
  WP_USERNAME             WordPress username
  WP_APP_PASSWORD         WordPress Application Password
  MCP_API_KEY             Bearer API key for SSE mode (if --api-key not set)

Modes:
  stdio   JSON-RPC over stdin/stdout (default, for local MCP clients)
  sse     HTTP Server-Sent Events (for remote/network MCP clients)

SSE endpoints:
  GET  /sse        Opens SSE connection, returns endpoint URL in first event
  POST /messages   Sends JSON-RPC requests (with ?session_id=...)

Examples:
  wordpress-mcp                                    # stdio mode
  wordpress-mcp --mode sse --addr :8080            # SSE on port 8080, no auth
  wordpress-mcp --mode sse --api-key secret123     # SSE with Bearer auth
  MCP_API_KEY=secret wordpress-mcp -m sse -a :9090 # SSE via env vars

`)
}

// runStdio runs the MCP server in stdio mode (original behavior).
func runStdio() {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 1024*1024), 10*1024*1024)

	writer := bufio.NewWriter(os.Stdout)
	defer writer.Flush()

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var req JSONRPCRequest
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			log.Printf("Failed to parse request: %v", err)
			continue
		}

		resp := handleRequest(req)
		respBytes, err := json.Marshal(resp)
		if err != nil {
			log.Printf("Failed to marshal response: %v", err)
			continue
		}

		fmt.Fprintln(writer, string(respBytes))
		writer.Flush()
	}

	if err := scanner.Err(); err != nil && err != io.EOF {
		log.Printf("Scanner error: %v", err)
	}
}

// runSSE runs the MCP server in SSE/HTTP mode with optional Bearer auth.
func runSSE(addr, apiKey string) {
	server := NewSSEServer(apiKey)
	if err := server.Start(addr); err != nil {
		log.Fatalf("SSE server error: %v", err)
	}
}
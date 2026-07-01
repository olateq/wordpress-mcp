# wordpress-mcp

MCP (Model Context Protocol) server for WordPress REST API.

Exposes **37 tools** covering Posts, Pages, Media, Users, Comments, Categories, Tags, and Settings — accessible via **stdio** (local) or **HTTP/SSE** (remote) transport.

## Features

- **37 MCP tools** — full CRUD for all major WordPress resources
- **Two transport modes** — stdio (default) and SSE/HTTP
- **Bearer authentication** — optional API key for SSE mode
- **Zero external dependencies** — Go standard library + `google/uuid` only
- **CORS support** — ready for browser-based MCP clients
- **Session management** — concurrent SSE sessions with heartbeat

## Project structure

```
wordpress-mcp/
├── main.go              # MCP server (JSON-RPC) + tool dispatch + CLI
├── server_sse.go        # SSE/HTTP server with Bearer auth + session management
├── go.mod               # Module definition
├── Taskfile.yml         # Build, cross-compile, test, vet, fmt, clean
├── wordpress-mcp.service # systemd unit file for production deployment
├── README.md            # This file
└── wordpress/
    ├── client.go        # HTTP client + types + CRUD posts
    ├── pages.go         # CRUD pages
    ├── media.go         # CRUD media + upload multipart
    ├── users.go         # CRUD users
    ├── comments.go      # CRUD comments
    ├── taxonomies.go    # CRUD categories + tags
    └── settings.go      # Get/update site settings
```

## Configuration

### Environment variables

| Variable | Description | Required |
|---|---|---|
| `WP_BASE_URL` | WordPress site URL (e.g. `https://your-site.com`) | ✅ |
| `WP_USERNAME` | WordPress username | ✅ |
| `WP_APP_PASSWORD` | WordPress Application Password | ✅ |
| `MCP_API_KEY` | Bearer API key for SSE mode (if `--api-key` not set) | SSE only |

## Usage

### Stdio mode (default — for local MCP clients)

```bash
# Build
task build

# Configure
export WP_BASE_URL="https://your-site.com"
export WP_USERNAME="admin"
export WP_APP_PASSWORD="xxxx xxxx xxxx xxxx"

# Run
./wordpress-mcp
```

**MCP client config (stdio):**
```json
{
  "mcpServers": {
    "wordpress": {
      "command": "/path/to/wordpress-mcp",
      "env": {
        "WP_BASE_URL": "https://your-site.com",
        "WP_USERNAME": "admin",
        "WP_APP_PASSWORD": "your-app-password"
      }
    }
  }
}
```

### SSE/HTTP mode (for remote/network MCP clients)

```bash
# SSE without auth (not recommended for production)
./wordpress-mcp --mode sse --addr :8080

# SSE with Bearer authentication
./wordpress-mcp --mode sse --addr :8080 --api-key your-secret-key

# SSE via environment variable
MCP_API_KEY=your-secret-key ./wordpress-mcp -m sse -a :8080
```

**SSE endpoints:**

| Method | Path | Description |
|---|---|---|
| `GET` | `/sse` | Opens SSE connection, returns endpoint URL in first event |
| `POST` | `/messages?session_id=...` | Sends JSON-RPC requests |

**Authentication:**

All requests must include an `Authorization: Bearer <your-api-key>` header (when `--api-key` or `MCP_API_KEY` is set).

**MCP client config (SSE):**
```json
{
  "mcpServers": {
    "wordpress": {
      "url": "http://your-server:8080/sse",
      "headers": {
        "Authorization": "Bearer your-secret-key"
      }
    }
  }
}
```

### How SSE sessions work

1. Client opens `GET /sse` → server responds with `event: endpoint` containing the POST URL with a `session_id`
2. Client sends JSON-RPC requests via `POST /messages?session_id=...`
3. Server responds with `202 Accepted` and pushes the JSON-RPC response through the SSE channel
4. Heartbeat every 30s keeps the connection alive


### systemd service (production deployment)

A ready-to-use systemd unit file is provided (`wordpress-mcp.service`) for running the server as a background service.

**Setup:**

```bash
# 1. Build and install the binary
 task build
sudo cp wordpress-mcp /usr/local/bin/

# 2. Create a dedicated user
sudo useradd -r -s /usr/sbin/nologin -M wordpress-mcp

# 3. Copy the service file
sudo cp wordpress-mcp.service /etc/systemd/system/

# 4. Edit the environment variables in the unit file
#    (or use an EnvironmentFile at /etc/wordpress-mcp/env)
sudo systemctl edit --full wordpress-mcp

# 5. Reload and start
sudo systemctl daemon-reload
sudo systemctl enable --now wordpress-mcp
```

**Alternatively, use an env file** (recommended for secrets):

```bash
sudo mkdir /etc/wordpress-mcp
sudo tee /etc/wordpress-mcp/env << 'EOF'
WP_BASE_URL=https://your-site.com
WP_USERNAME=admin
WP_APP_PASSWORD=xxxx xxxx xxxx xxxx
MCP_API_KEY=your-secret-key
EOF
sudo chmod 600 /etc/wordpress-mcp/env
sudo chown wordpress-mcp:wordpress-mcp /etc/wordpress-mcp/env
```

Then uncomment the `EnvironmentFile=` line in the unit file and remove the individual `Environment=` lines.

**Service management:**

```bash
sudo systemctl status wordpress-mcp   # check status
sudo systemctl restart wordpress-mcp  # restart
journalctl -u wordpress-mcp -f        # view logs
```

The unit file includes security hardening (`NoNewPrivileges`, `ProtectSystem=strict`, `ProtectHome`, `PrivateTmp`) and automatic restart on failure.

## CLI flags

```
wordpress-mcp — WordPress REST API MCP Server

Flags:
  --mode, -m <mode>       Transport mode: "stdio" (default) or "sse"
  --addr, -a <addr>       SSE listen address (default ":8080")
  --api-key, -k <key>     Bearer API key for SSE mode (or set MCP_API_KEY env var)
  --help, -h              Show this help
```

## Available tools (37)

| Domain | Tools |
|---|---|
| Posts (5) | `wp_list_posts`, `wp_get_post`, `wp_create_post`, `wp_update_post`, `wp_delete_post` |
| Pages (5) | `wp_list_pages`, `wp_get_page`, `wp_create_page`, `wp_update_page`, `wp_delete_page` |
| Media (5) | `wp_list_media`, `wp_get_media`, `wp_upload_media`, `wp_update_media`, `wp_delete_media` |
| Users (5) | `wp_list_users`, `wp_get_user`, `wp_create_user`, `wp_update_user`, `wp_delete_user` |
| Comments (5) | `wp_list_comments`, `wp_get_comment`, `wp_create_comment`, `wp_update_comment`, `wp_delete_comment` |
| Categories (5) | `wp_list_categories`, `wp_get_category`, `wp_create_category`, `wp_update_category`, `wp_delete_category` |
| Tags (5) | `wp_list_tags`, `wp_get_tag`, `wp_create_tag`, `wp_update_tag`, `wp_delete_tag` |
| Settings (2) | `wp_get_settings`, `wp_update_settings` |

## Build

```bash
# Build
task build

# Cross-compile (linux/darwin × amd64/arm64)
task build-all

# Run tests
task test

# Vet
task vet

# Format
task fmt

# Clean
task clean
```

## Binary

- **Size**: ~7 MB (stripped with `-ldflags="-s -w"`)
- **Protocol**: MCP `2024-11-05` (initialize, tools/list, tools/call, ping)
- **Auth**: WordPress Application Passwords (Basic Auth) + Bearer API key (SSE mode)

## License

This project is licensed under the **Sustainable Use License v1.0** (SUL-1.0).

Copyright (c) 2026 Martin Urbain

See the [LICENSE](LICENSE.md) file for the full text.

SPDX-License-Identifier: SUL-1.0

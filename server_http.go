package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// ── HTTP Session ──────────────────────────────────────────────────────────────

type HTTPSession struct {
	ID      string
	Created time.Time
	mu      sync.Mutex
}

func NewHTTPSession() *HTTPSession {
	return &HTTPSession{
		ID:      uuid.NewString(),
		Created: time.Now(),
	}
}

// ── HTTP Server (Streamable HTTP transport) ───────────────────────────────────

type HTTPServer struct {
	Sessions sync.Map // map[string]*HTTPSession
	APIKey   string
}

func NewHTTPServer(apiKey string) *HTTPServer {
	return &HTTPServer{APIKey: apiKey}
}

// bearerAuth checks for a valid Bearer token.
func (s *HTTPServer) bearerAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.APIKey == "" {
			next(w, r)
			return
		}
		auth := r.Header.Get("Authorization")
		if auth == "" {
			http.Error(w, `{"error":"missing Authorization header"}`, http.StatusUnauthorized)
			return
		}
		const prefix = "Bearer "
		if len(auth) < len(prefix) || auth[:len(prefix)] != prefix {
			http.Error(w, `{"error":"invalid Authorization header, expected Bearer token"}`, http.StatusUnauthorized)
			return
		}
		token := auth[len(prefix):]
		if token != s.APIKey {
			http.Error(w, `{"error":"invalid API key"}`, http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}

// corsMiddlewareHTTP adds CORS headers for browser-based clients.
func corsMiddlewareHTTP(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Mcp-Session-Id, Accept")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next(w, r)
	}
}

// validateSession checks the Mcp-Session-Id header and returns the session if valid.
func (s *HTTPServer) validateSession(r *http.Request) (*HTTPSession, bool) {
	sessionID := r.Header.Get("Mcp-Session-Id")
	if sessionID == "" {
		return nil, false
	}
	val, ok := s.Sessions.Load(sessionID)
	if !ok {
		return nil, false
	}
	return val.(*HTTPSession), true
}

// handleMCPPost handles POST /mcp — receives JSON-RPC requests and returns responses.
func (s *HTTPServer) handleMCPPost(w http.ResponseWriter, r *http.Request) {
	var req JSONRPCRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, `{"error":"invalid JSON: %s"}`, err.Error())
		return
	}

	// Check if this is an initialize request
	isInitialize := req.Method == "initialize"

	if isInitialize {
		// Create a new session
		session := NewHTTPSession()
		s.Sessions.Store(session.ID, session)
		w.Header().Set("Mcp-Session-Id", session.ID)
	} else {
		// Validate session for non-initialize requests
		if _, ok := s.validateSession(r); !ok {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, `{"error":"Invalid or missing Mcp-Session-Id header. Send an 'initialize' request first."}`)
			return
		}
	}

	resp := handleRequest(req)

	// Check if this is a notification (no ID field)
	isNotification := len(req.ID) == 0

	if isNotification {
		w.WriteHeader(http.StatusAccepted)
		return
	}

	// Determine response format based on Accept header
	accept := r.Header.Get("Accept")
	if strings.Contains(accept, "text/event-stream") && !strings.Contains(accept, "application/json") {
		// Client only accepts SSE — wrap response in SSE format
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		respBytes, _ := json.Marshal(resp)
		fmt.Fprintf(w, "data: %s\n\n", respBytes)
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
	} else {
		// Default: return JSON directly
		w.Header().Set("Content-Type", "application/json")
		respBytes, err := json.Marshal(resp)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, `{"error":"failed to marshal response"}`)
			return
		}
		w.Write(respBytes)
	}
}

// handleMCPGet handles GET /mcp — opens an SSE stream for server-to-client notifications.
func (s *HTTPServer) handleMCPGet(w http.ResponseWriter, r *http.Request) {
	// Validate session
	if _, ok := s.validateSession(r); !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"error":"Invalid or missing Mcp-Session-Id header"}`)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	// Heartbeat to keep connection alive
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			fmt.Fprintf(w, ": heartbeat\n\n")
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		}
	}
}

// handleMCPDelete handles DELETE /mcp — terminates a session.
func (s *HTTPServer) handleMCPDelete(w http.ResponseWriter, r *http.Request) {
	sessionID := r.Header.Get("Mcp-Session-Id")
	if sessionID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"error":"missing Mcp-Session-Id header"}`)
		return
	}

	if _, ok := s.Sessions.LoadAndDelete(sessionID); !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"error":"invalid or expired session"}`)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, `{"status":"session terminated"}`)
}

// handleMCP routes to the appropriate handler based on HTTP method.
func (s *HTTPServer) handleMCP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		s.handleMCPPost(w, r)
	case http.MethodGet:
		s.handleMCPGet(w, r)
	case http.MethodDelete:
		s.handleMCPDelete(w, r)
	default:
		w.Header().Set("Allow", "GET, POST, DELETE")
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// Start starts the HTTP server on the given address.
func (s *HTTPServer) Start(addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/mcp", corsMiddlewareHTTP(s.bearerAuth(s.handleMCP)))

	log.Printf("WordPress MCP HTTP server listening on %s", addr)
	log.Printf("MCP endpoint: http://%s/mcp", addr)
	if s.APIKey != "" {
		log.Printf("Auth: Bearer token required")
	} else {
		log.Printf("WARNING: No API key set — server is open to all connections!")
	}

	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	return srv.ListenAndServe()
}

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
)

// ── SSE Session ──────────────────────────────────────────────────────────────

type SSESession struct {
	ID       string
	Events   chan string
	Closed   bool
	mu       sync.Mutex
}

func NewSSESession() *SSESession {
	return &SSESession{
		ID:     uuid.NewString(),
		Events: make(chan string, 64),
	}
}

func (s *SSESession) Send(event string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.Closed {
		return
	}
	select {
	case s.Events <- event:
	default:
		// channel full, drop oldest
		select {
		case <-s.Events:
		default:
		}
		s.Events <- event
	}
}

func (s *SSESession) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.Closed {
		s.Closed = true
		close(s.Events)
	}
}

// ── SSE Server ───────────────────────────────────────────────────────────────

type SSEServer struct {
	Sessions sync.Map // map[string]*SSESession
	APIKey   string
}

func NewSSEServer(apiKey string) *SSEServer {
	return &SSEServer{APIKey: apiKey}
}

// bearerAuth is middleware that checks for a valid Bearer token.
func (s *SSEServer) bearerAuth(next http.HandlerFunc) http.HandlerFunc {
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

// corsMiddleware adds CORS headers for browser-based clients.
func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Mcp-Session-Id")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next(w, r)
	}
}

// handleSSE handles GET /sse — opens a persistent SSE connection.
func (s *SSEServer) handleSSE(w http.ResponseWriter, r *http.Request) {
	session := NewSSESession()
	s.Sessions.Store(session.ID, session)
	defer func() {
		s.Sessions.Delete(session.ID)
		session.Close()
	}()

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	// Send the endpoint event so the client knows where to POST messages.
	endpointURL := fmt.Sprintf("/messages?session_id=%s", session.ID)
	fmt.Fprintf(w, "event: endpoint\ndata: %s\n\n", endpointURL)
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}

	// Heartbeat ticker
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-session.Events:
			if !ok {
				return
			}
			fmt.Fprintf(w, "data: %s\n\n", event)
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		case <-ticker.C:
			fmt.Fprintf(w, ": heartbeat\n\n")
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		}
	}
}

// handleMessages handles POST /messages — receives JSON-RPC requests.
func (s *SSEServer) handleMessages(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		http.Error(w, `{"error":"missing session_id parameter"}`, http.StatusBadRequest)
		return
	}

	val, ok := s.Sessions.Load(sessionID)
	if !ok {
		http.Error(w, `{"error":"invalid or expired session"}`, http.StatusNotFound)
		return
	}
	session := val.(*SSESession)

	var req JSONRPCRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid JSON: `+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	resp := handleRequest(req)
	respBytes, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, `{"error":"failed to marshal response"}`, http.StatusInternalServerError)
		return
	}

	// Send the response back through the SSE channel.
	session.Send(string(respBytes))

	// Also return 202 Accepted to the POST request.
	w.WriteHeader(http.StatusAccepted)
}

// Start starts the SSE server on the given address.
func (s *SSEServer) Start(addr string) error {
	mux := http.NewServeMux()

	mux.HandleFunc("/sse", corsMiddleware(s.bearerAuth(s.handleSSE)))
	mux.HandleFunc("/messages", corsMiddleware(s.bearerAuth(s.handleMessages)))

	log.Printf("WordPress MCP SSE server listening on %s", addr)
	log.Printf("SSE endpoint:   http://%s/sse", addr)
	log.Printf("POST endpoint:  http://%s/messages", addr)
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
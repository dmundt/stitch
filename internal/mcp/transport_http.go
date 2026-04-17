package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/dmundt/stitch/css"
)

func (a *app) serveHTTP(ctx context.Context) error {
	mux := http.NewServeMux()

	assets, err := css.Assets()
	if err != nil {
		return err
	}
	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.FS(assets))))
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	})
	mux.HandleFunc("/sessions/", a.handleSessionHTTP)
	mux.HandleFunc(defaultMCPPath, a.handleMCPHTTP)
	mux.HandleFunc(defaultSSEPath, a.handleMCPSSE)

	srv := &http.Server{
		Addr:    listenAddr,
		Handler: mux,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()

	log.Printf("preview + MCP HTTP server listening at %s", DefaultHTTPEndpoint)
	err = srv.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func (a *app) handleSessionHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	parts := strings.Split(strings.Trim(path.Clean(r.URL.Path), "/"), "/")
	if len(parts) < 3 || parts[0] != "sessions" {
		http.NotFound(w, r)
		return
	}
	sessionID := parts[1]

	session, err := a.store.getSession(sessionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	switch {
	case len(parts) == 3 && parts[2] == "page":
		htmlText, err := a.renderFull(session)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(htmlText))
		return
	case len(parts) == 4 && parts[2] == "blocks":
		block := parts[3]
		fragment, err := a.renderBlock(session, block)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(fragment))
		return
	case len(parts) == 4 && parts[2] == "components":
		id := parts[3]
		fragment, err := a.renderComponent(session, id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(fragment))
		return
	default:
		http.NotFound(w, r)
	}
}

func (a *app) handleMCPSSE(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}
	_, _ = io.WriteString(w, "event: ready\ndata: {\"ok\":true}\n\n")
	flusher.Flush()

	ticker := time.NewTicker(25 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
			_, _ = io.WriteString(w, "event: ping\ndata: {}\n\n")
			flusher.Flush()
		}
	}
}

func (a *app) handleMCPHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req rpcRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	resp := a.handleRPC(req)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

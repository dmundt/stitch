package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/dmundt/stitch/css"
	"github.com/dmundt/stitch/internal/brand"
	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

func (a *app) serveHTTP(ctx context.Context) error {
	mux := http.NewServeMux()
	brand.MountRoutes(mux)

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

	sdkServer, err := a.newSDKServer()
	if err != nil {
		return err
	}
	streamable := &sdkmcp.StreamableServerTransport{
		SessionID: randomID("mcp-http"),
	}
	if _, err := sdkServer.Connect(ctx, streamable, nil); err != nil {
		return err
	}
	mux.Handle(defaultMCPPath, streamable)
	mux.Handle(defaultSSEPath, streamable)

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

	log.Printf("%s | preview + MCP HTTP server listening at %s", brand.BrandLine, DefaultHTTPEndpoint)
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
	case len(parts) == 3 && parts[2] == "diagnostics":
		payload := a.buildSessionDiagnostics(session)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(payload)
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

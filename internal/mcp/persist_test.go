package mcp

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestFileStoreRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sessions.json")

	// Create a session in a fileStore.
	fs1, err := newFileStore(path)
	if err != nil {
		t.Fatalf("newFileStore: %v", err)
	}
	s := fs1.createSession("Persist Test", "stitch")
	if s.ID == "" {
		t.Fatal("expected non-empty session ID")
	}
	_, err = fs1.updateSession(s.ID, func(sess *sessionState) error {
		sess.Page.Title = "Updated Title"
		return nil
	})
	if err != nil {
		t.Fatalf("updateSession: %v", err)
	}

	// Reload from the same file — session must survive.
	fs2, err := newFileStore(path)
	if err != nil {
		t.Fatalf("newFileStore reload: %v", err)
	}
	got, err := fs2.getSession(s.ID)
	if err != nil {
		t.Fatalf("getSession after reload: %v", err)
	}
	if got.ID != s.ID {
		t.Fatalf("expected ID %q, got %q", s.ID, got.ID)
	}
	if got.Page.Title != "Updated Title" {
		t.Fatalf("expected title 'Updated Title', got %q", got.Page.Title)
	}
}

func TestFileStoreMissingFileGraceful(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nonexistent", "sessions.json")

	// File (and its parent dir) do not exist — should start empty without error.
	fs, err := newFileStore(path)
	if err != nil {
		t.Fatalf("expected graceful start for missing file, got: %v", err)
	}
	_, err = fs.getSession("anything")
	if err == nil {
		t.Fatal("expected error for unknown session in empty store")
	}
}

func TestFileStoreDeletePersists(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sessions.json")

	fs1, err := newFileStore(path)
	if err != nil {
		t.Fatalf("newFileStore: %v", err)
	}
	s := fs1.createSession("To Delete", "stitch")

	deleted, err := fs1.deleteSession(s.ID)
	if err != nil || !deleted {
		t.Fatalf("deleteSession: deleted=%v err=%v", deleted, err)
	}

	// Reload: deleted session must not reappear.
	fs2, err := newFileStore(path)
	if err != nil {
		t.Fatalf("newFileStore reload: %v", err)
	}
	_, err = fs2.getSession(s.ID)
	if err == nil {
		t.Fatal("expected deleted session to be absent after reload")
	}
}

func TestFileStoreCorruptFileReturnsError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sessions.json")
	if err := os.WriteFile(path, []byte("not valid json{{{"), 0600); err != nil {
		t.Fatalf("write corrupt file: %v", err)
	}
	_, err := newFileStore(path)
	if err == nil {
		t.Fatal("expected error for corrupt sessions file")
	}
}

func TestFileStoreAppRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sessions.json")

	fs1, err := newFileStore(path)
	if err != nil {
		t.Fatalf("newFileStore: %v", err)
	}
	session := fs1.createSession("Persist App Session", "stitch")
	created := fs1.createApp("Persist App")
	_, err = fs1.updateApp(created.ID, func(app *appState) error {
		app.Shell = &appShellState{SessionID: session.ID, Block: "main"}
		app.Routes = append(app.Routes, appRoute{ID: "route_1", Path: "/", SessionID: session.ID, Block: "main"})
		return nil
	})
	if err != nil {
		t.Fatalf("updateApp: %v", err)
	}

	fs2, err := newFileStore(path)
	if err != nil {
		t.Fatalf("newFileStore reload: %v", err)
	}
	reloaded, err := fs2.getApp(created.ID)
	if err != nil {
		t.Fatalf("getApp after reload: %v", err)
	}
	if reloaded.Name != "Persist App" {
		t.Fatalf("expected app name to survive, got %q", reloaded.Name)
	}
	if reloaded.Shell == nil || reloaded.Shell.SessionID != session.ID {
		t.Fatalf("expected shell to survive reload, got %+v", reloaded.Shell)
	}
	if len(reloaded.Routes) != 1 || reloaded.Routes[0].Path != "/" {
		t.Fatalf("expected routes to survive reload, got %+v", reloaded.Routes)
	}
}

func TestFileStoreLoadsLegacySessionMap(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sessions.json")
	legacy := map[string]*sessionState{
		"sess_legacy": {
			ID:         "sess_legacy",
			Page:       pageState{Title: "Legacy", Lang: "en", Provider: "stitch"},
			Components: map[string]*componentNode{},
			Blocks:     map[string][]string{},
		},
	}
	data, err := json.Marshal(legacy)
	if err != nil {
		t.Fatalf("marshal legacy map: %v", err)
	}
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatalf("write legacy file: %v", err)
	}

	fs, err := newFileStore(path)
	if err != nil {
		t.Fatalf("newFileStore legacy load failed: %v", err)
	}
	loaded, err := fs.getSession("sess_legacy")
	if err != nil {
		t.Fatalf("expected legacy session to load: %v", err)
	}
	if loaded.Page.Title != "Legacy" {
		t.Fatalf("expected legacy title, got %q", loaded.Page.Title)
	}
	created := fs.createApp("fresh app")
	if created.ID == "" {
		t.Fatal("expected app creation to work after legacy load")
	}
}

func TestFileStoreBuildRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sessions.json")

	fs1, err := newFileStore(path)
	if err != nil {
		t.Fatalf("newFileStore: %v", err)
	}
	build := fs1.putBuild(&appBuildState{
		ID:           "build_1",
		AppID:        "app_1",
		Target:       "go-htmx",
		Name:         "Test App",
		Title:        "Test App",
		Lang:         "en",
		Provider:     "stitch",
		HeadSnippets: []string{"<meta charset=\"utf-8\">"},
		Routes:       []appBuildRoute{{RouteID: "route_1", Name: "overview", Path: "/", SessionID: "sess_1", Block: "main"}},
	})
	if build.ID != "build_1" {
		t.Fatalf("expected build to be stored, got %+v", build)
	}

	fs2, err := newFileStore(path)
	if err != nil {
		t.Fatalf("newFileStore reload: %v", err)
	}
	reloaded, err := fs2.getBuild("build_1")
	if err != nil {
		t.Fatalf("getBuild after reload: %v", err)
	}
	if reloaded.Target != "go-htmx" || len(reloaded.Routes) != 1 {
		t.Fatalf("expected build to survive reload, got %+v", reloaded)
	}
}

package mcp

import (
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

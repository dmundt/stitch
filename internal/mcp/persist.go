package mcp

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
)

// fileStore is a write-through sessionStore backed by a JSON file on disk.
// Reads are served from the embedded in-memory memStore; every mutation is
// flushed atomically to disk via a temp-file + rename so that sessions survive
// server restarts.
type fileStore struct {
	*memStore
	path   string
	saveMu sync.Mutex
}

// newFileStore constructs a fileStore rooted at path, pre-loading any
// sessions persisted from a previous run.  A missing file is not an error
// (first run).  A malformed file returns an error so the caller can fall back
// to a plain memStore instead of silently losing data.
func newFileStore(path string) (sessionStore, error) {
	fs := &fileStore{
		memStore: &memStore{sessions: map[string]*sessionState{}},
		path:     path,
	}
	if err := fs.loadFromDisk(); err != nil {
		return nil, err
	}
	return fs, nil
}

// loadFromDisk reads the persisted sessions file into the in-memory map.
// File-not-found is treated as an empty store (first run).
func (f *fileStore) loadFromDisk() error {
	data, err := os.ReadFile(f.path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("stitch: load sessions %s: %w", f.path, err)
	}
	if len(data) == 0 {
		return nil
	}
	var loaded map[string]*sessionState
	if err := json.Unmarshal(data, &loaded); err != nil {
		return fmt.Errorf("stitch: parse sessions %s: %w", f.path, err)
	}
	for id, s := range loaded {
		if s != nil {
			f.memStore.sessions[id] = s
		}
	}
	return nil
}

// persist snapshots the current in-memory sessions map to disk atomically.
// It serialises disk writes via saveMu so concurrent mutations never interleave
// partial writes.
func (f *fileStore) persist() {
	f.saveMu.Lock()
	defer f.saveMu.Unlock()

	f.memStore.mu.RLock()
	data, err := json.MarshalIndent(f.memStore.sessions, "", "  ")
	f.memStore.mu.RUnlock()

	if err != nil {
		log.Printf("stitch: persist marshal: %v", err)
		return
	}

	dir := filepath.Dir(f.path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		log.Printf("stitch: persist mkdir %s: %v", dir, err)
		return
	}

	tmp, err := os.CreateTemp(dir, "stitch-sessions-*.tmp")
	if err != nil {
		log.Printf("stitch: persist tempfile: %v", err)
		return
	}
	tmpPath := tmp.Name()

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		log.Printf("stitch: persist write: %v", err)
		return
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpPath)
		log.Printf("stitch: persist close: %v", err)
		return
	}
	if err := os.Rename(tmpPath, f.path); err != nil {
		os.Remove(tmpPath)
		log.Printf("stitch: persist rename: %v", err)
	}
}

// createSession delegates to memStore then persists.
func (f *fileStore) createSession(title, provider string) *sessionState {
	s := f.memStore.createSession(title, provider)
	f.persist()
	return s
}

// updateSession delegates to memStore then persists on success.
func (f *fileStore) updateSession(id string, fn func(*sessionState) error) (*sessionState, error) {
	s, err := f.memStore.updateSession(id, fn)
	if err == nil {
		f.persist()
	}
	return s, err
}

// deleteSession delegates to memStore then persists when a session was removed.
func (f *fileStore) deleteSession(id string) (bool, error) {
	ok, err := f.memStore.deleteSession(id)
	if err == nil && ok {
		f.persist()
	}
	return ok, err
}

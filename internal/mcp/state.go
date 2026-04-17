package mcp

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/dmundt/stitch/htmx"
	"github.com/dmundt/stitch/internal/brand"
	stitchtpl "github.com/dmundt/stitch/template"
)

type componentNode struct {
	ID       string            `json:"id"`
	Type     string            `json:"type"`
	Props    map[string]any    `json:"props"`
	Attrs    map[string]string `json:"attrs,omitempty"`
	Children []string          `json:"children"`
}

type pageState struct {
	Title        string   `json:"title"`
	Lang         string   `json:"lang"`
	Provider     string   `json:"provider"`
	HeadSnippets []string `json:"head_snippets"`
}

type sessionState struct {
	ID         string                    `json:"id"`
	Page       pageState                 `json:"page"`
	Components map[string]*componentNode `json:"components"`
	Blocks     map[string][]string       `json:"blocks"`
	CreatedAt  time.Time                 `json:"created_at"`
	UpdatedAt  time.Time                 `json:"updated_at"`
}

type appShellState struct {
	SessionID string `json:"session_id"`
	Block     string `json:"block"`
}

type appRoute struct {
	ID        string    `json:"id"`
	Name      string    `json:"name,omitempty"`
	Path      string    `json:"path"`
	SessionID string    `json:"session_id"`
	Block     string    `json:"block"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type appState struct {
	ID        string         `json:"id"`
	Name      string         `json:"name"`
	Shell     *appShellState `json:"shell,omitempty"`
	Routes    []appRoute     `json:"routes"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

type appBuildRoute struct {
	RouteID   string `json:"route_id"`
	Name      string `json:"name"`
	Path      string `json:"path"`
	SessionID string `json:"session_id"`
	Block     string `json:"block"`
}

type appBuildState struct {
	ID           string           `json:"id"`
	AppID        string           `json:"app_id"`
	Target       string           `json:"target"`
	Name         string           `json:"name"`
	Title        string           `json:"title"`
	Lang         string           `json:"lang"`
	Provider     string           `json:"provider"`
	HeadSnippets []string         `json:"head_snippets"`
	Shell        *appShellState   `json:"shell,omitempty"`
	Routes       []appBuildRoute  `json:"routes"`
	Warnings     []map[string]any `json:"warnings"`
	CreatedAt    time.Time        `json:"created_at"`
	UpdatedAt    time.Time        `json:"updated_at"`
}

// sessionStore is the interface for session lifecycle management.
type sessionStore interface {
	createSession(title, provider string) *sessionState
	getSession(id string) (*sessionState, error)
	updateSession(id string, fn func(*sessionState) error) (*sessionState, error)
	deleteSession(id string) (bool, error)
	createApp(name string) *appState
	getApp(id string) (*appState, error)
	updateApp(id string, fn func(*appState) error) (*appState, error)
	putBuild(build *appBuildState) *appBuildState
	getBuild(id string) (*appBuildState, error)
}

type memStore struct {
	mu       sync.RWMutex
	sessions map[string]*sessionState
	apps     map[string]*appState
	builds   map[string]*appBuildState
}

func newStore() sessionStore {
	return &memStore{sessions: map[string]*sessionState{}, apps: map[string]*appState{}, builds: map[string]*appBuildState{}}
}

func defaultHeadSnippets() []string {
	return []string{string(htmx.Head()), brand.FaviconLinkTag()}
}

func (s *memStore) createSession(title, provider string) *sessionState {
	s.mu.Lock()
	defer s.mu.Unlock()

	if strings.TrimSpace(title) == "" {
		title = "Stitch MCP UI"
	}
	if strings.TrimSpace(provider) == "" {
		provider = defaultProvider
	}
	id := randomID("sess")
	now := time.Now().UTC()
	session := &sessionState{
		ID: id,
		Page: pageState{
			Title:        title,
			Lang:         "en",
			Provider:     provider,
			HeadSnippets: defaultHeadSnippets(),
		},
		Components: map[string]*componentNode{},
		Blocks: map[string][]string{
			stitchtpl.BlockHeader: {},
			stitchtpl.BlockMain:   {},
			stitchtpl.BlockFooter: {},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
	s.sessions[id] = session
	return cloneSession(session)
}

func (s *memStore) getSession(id string) (*sessionState, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	session, ok := s.sessions[id]
	if !ok {
		return nil, fmt.Errorf("unknown session: %s", id)
	}
	return cloneSession(session), nil
}

func (s *memStore) updateSession(id string, fn func(*sessionState) error) (*sessionState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	session, ok := s.sessions[id]
	if !ok {
		return nil, fmt.Errorf("unknown session: %s", id)
	}
	if err := fn(session); err != nil {
		return nil, err
	}
	session.UpdatedAt = time.Now().UTC()
	return cloneSession(session), nil
}

func (s *memStore) deleteSession(id string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.sessions[id]; !ok {
		return false, nil
	}
	delete(s.sessions, id)
	return true, nil
}

func (s *memStore) createApp(name string) *appState {
	s.mu.Lock()
	defer s.mu.Unlock()

	if strings.TrimSpace(name) == "" {
		name = "Stitch App"
	}
	id := randomID("app")
	now := time.Now().UTC()
	app := &appState{
		ID:        id,
		Name:      name,
		Routes:    []appRoute{},
		CreatedAt: now,
		UpdatedAt: now,
	}
	s.apps[id] = app
	return cloneApp(app)
}

func (s *memStore) getApp(id string) (*appState, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	app, ok := s.apps[id]
	if !ok {
		return nil, fmt.Errorf("unknown app: %s", id)
	}
	return cloneApp(app), nil
}

func (s *memStore) updateApp(id string, fn func(*appState) error) (*appState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	app, ok := s.apps[id]
	if !ok {
		return nil, fmt.Errorf("unknown app: %s", id)
	}
	if err := fn(app); err != nil {
		return nil, err
	}
	app.UpdatedAt = time.Now().UTC()
	return cloneApp(app), nil
}

func (s *memStore) putBuild(build *appBuildState) *appBuildState {
	s.mu.Lock()
	defer s.mu.Unlock()
	copyBuild := cloneBuild(build)
	s.builds[copyBuild.ID] = copyBuild
	return cloneBuild(copyBuild)
}

func (s *memStore) getBuild(id string) (*appBuildState, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	build, ok := s.builds[id]
	if !ok {
		return nil, fmt.Errorf("unknown build: %s", id)
	}
	return cloneBuild(build), nil
}

func cloneSession(in *sessionState) *sessionState {
	out := &sessionState{
		ID:        in.ID,
		Page:      in.Page,
		Blocks:    map[string][]string{},
		CreatedAt: in.CreatedAt,
		UpdatedAt: in.UpdatedAt,
	}
	out.Page.HeadSnippets = append([]string{}, in.Page.HeadSnippets...)
	out.Components = map[string]*componentNode{}
	for id, node := range in.Components {
		copyNode := &componentNode{
			ID:       node.ID,
			Type:     node.Type,
			Props:    map[string]any{},
			Attrs:    map[string]string{},
			Children: append([]string{}, node.Children...),
		}
		for k, v := range node.Props {
			copyNode.Props[k] = deepCopyAny(v)
		}
		for k, v := range node.Attrs {
			copyNode.Attrs[k] = v
		}
		out.Components[id] = copyNode
	}
	for block, ids := range in.Blocks {
		out.Blocks[block] = append(out.Blocks[block], ids...)
	}
	return out
}

func cloneApp(in *appState) *appState {
	if in == nil {
		return nil
	}
	out := &appState{
		ID:        in.ID,
		Name:      in.Name,
		Routes:    make([]appRoute, len(in.Routes)),
		CreatedAt: in.CreatedAt,
		UpdatedAt: in.UpdatedAt,
	}
	if in.Shell != nil {
		out.Shell = &appShellState{
			SessionID: in.Shell.SessionID,
			Block:     in.Shell.Block,
		}
	}
	copy(out.Routes, in.Routes)
	return out
}

func cloneBuild(in *appBuildState) *appBuildState {
	if in == nil {
		return nil
	}
	out := &appBuildState{
		ID:           in.ID,
		AppID:        in.AppID,
		Target:       in.Target,
		Name:         in.Name,
		Title:        in.Title,
		Lang:         in.Lang,
		Provider:     in.Provider,
		HeadSnippets: append([]string{}, in.HeadSnippets...),
		Routes:       append([]appBuildRoute{}, in.Routes...),
		Warnings:     make([]map[string]any, 0, len(in.Warnings)),
		CreatedAt:    in.CreatedAt,
		UpdatedAt:    in.UpdatedAt,
	}
	if in.Shell != nil {
		out.Shell = &appShellState{SessionID: in.Shell.SessionID, Block: in.Shell.Block}
	}
	for _, warning := range in.Warnings {
		copyWarning := map[string]any{}
		for key, value := range warning {
			copyWarning[key] = deepCopyAny(value)
		}
		out.Warnings = append(out.Warnings, copyWarning)
	}
	return out
}

func deepCopyAny(v any) any {
	b, err := json.Marshal(v)
	if err != nil {
		return v
	}
	var out any
	if err := json.Unmarshal(b, &out); err != nil {
		return v
	}
	return out
}

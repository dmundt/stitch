package mcp

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/dmundt/stitch/htmx"
	stitchtpl "github.com/dmundt/stitch/template"
)

type componentNode struct {
	ID       string         `json:"id"`
	Type     string         `json:"type"`
	Props    map[string]any `json:"props"`
	Children []string       `json:"children"`
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

type stateStore struct {
	mu       sync.RWMutex
	sessions map[string]*sessionState
}

func newStore() *stateStore {
	return &stateStore{sessions: map[string]*sessionState{}}
}

func (s *stateStore) createSession(title, provider string) *sessionState {
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
			HeadSnippets: []string{string(htmx.Head())},
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

func (s *stateStore) getSession(id string) (*sessionState, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	session, ok := s.sessions[id]
	if !ok {
		return nil, fmt.Errorf("unknown session: %s", id)
	}
	return cloneSession(session), nil
}

func (s *stateStore) updateSession(id string, fn func(*sessionState) error) (*sessionState, error) {
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

func (s *stateStore) deleteSession(id string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.sessions[id]; !ok {
		return false, nil
	}
	delete(s.sessions, id)
	return true, nil
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
			Children: append([]string{}, node.Children...),
		}
		for k, v := range node.Props {
			copyNode.Props[k] = deepCopyAny(v)
		}
		out.Components[id] = copyNode
	}
	for block, ids := range in.Blocks {
		out.Blocks[block] = append(out.Blocks[block], ids...)
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

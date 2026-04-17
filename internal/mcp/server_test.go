package mcp

import (
	"strings"
	"testing"
)

func TestSessionCreateDefaults(t *testing.T) {
	a := newApp()

	res, err := a.safeDispatch("session.create", map[string]any{})
	if err != nil {
		t.Fatalf("session.create failed: %v", err)
	}

	s := res["session"].(*sessionState)
	if s.Page.Title != "Stitch MCP UI" {
		t.Fatalf("unexpected default title: %q", s.Page.Title)
	}
	if s.Page.Provider != "stitch" {
		t.Fatalf("unexpected default provider: %q", s.Page.Provider)
	}
	if s.Page.Lang != "en" {
		t.Fatalf("unexpected default language: %q", s.Page.Lang)
	}
	if len(s.Page.HeadSnippets) == 0 {
		t.Fatal("expected default htmx head snippet")
	}
}

func TestToolCallMissingRequiredArgReturnsError(t *testing.T) {
	a := newApp()

	_, err := a.safeDispatch("session.get", map[string]any{})
	if err == nil {
		t.Fatal("expected error for missing session_id")
	}
	if !strings.Contains(err.Error(), "missing required argument: session_id") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRenderBlockAndComponentFragments(t *testing.T) {
	a := newApp()

	sessionID := mustCreateSession(t, a)

	_, err := a.safeDispatch("ui.create_component", map[string]any{
		"session_id": sessionID,
		"type":       "heading",
		"props": map[string]any{
			"level": 1,
			"text":  "Header Title",
		},
		"block": "header",
	})
	if err != nil {
		t.Fatalf("create header failed: %v", err)
	}

	paragraphRes, err := a.safeDispatch("ui.create_component", map[string]any{
		"session_id": sessionID,
		"type":       "paragraph",
		"props": map[string]any{
			"text": "Hello from focused test",
		},
		"block": "main",
	})
	if err != nil {
		t.Fatalf("create paragraph failed: %v", err)
	}
	paragraphID := paragraphRes["component"].(*componentNode).ID

	blockRes, err := a.safeDispatch("render.block", map[string]any{
		"session_id": sessionID,
		"block":      "main",
	})
	if err != nil {
		t.Fatalf("render.block failed: %v", err)
	}
	mainHTML := blockRes["html"].(string)
	if !strings.Contains(mainHTML, "Hello from focused test") {
		t.Fatalf("main block missing paragraph text: %s", mainHTML)
	}

	componentRes, err := a.safeDispatch("render.component", map[string]any{
		"session_id":   sessionID,
		"component_id": paragraphID,
	})
	if err != nil {
		t.Fatalf("render.component failed: %v", err)
	}
	componentHTML := componentRes["html"].(string)
	if !strings.Contains(componentHTML, "Hello from focused test") {
		t.Fatalf("component fragment missing paragraph text: %s", componentHTML)
	}
	if strings.Contains(componentHTML, "Header Title") {
		t.Fatalf("component fragment should not contain sibling/header content: %s", componentHTML)
	}
}

func TestMoveComponentCycleRejected(t *testing.T) {
	a := newApp()
	sessionID := mustCreateSession(t, a)

	parentRes, err := a.safeDispatch("ui.create_component", map[string]any{
		"session_id": sessionID,
		"type":       "section",
		"props": map[string]any{
			"title": "Parent",
		},
		"block": "main",
	})
	if err != nil {
		t.Fatalf("create parent failed: %v", err)
	}
	parentID := parentRes["component"].(*componentNode).ID

	childRes, err := a.safeDispatch("ui.create_component", map[string]any{
		"session_id": sessionID,
		"type":       "paragraph",
		"props": map[string]any{
			"text": "Child",
		},
		"parent_id": parentID,
	})
	if err != nil {
		t.Fatalf("create child failed: %v", err)
	}
	childID := childRes["component"].(*componentNode).ID

	_, err = a.safeDispatch("ui.move_component", map[string]any{
		"session_id":    sessionID,
		"component_id":  parentID,
		"new_parent_id": childID,
	})
	if err == nil {
		t.Fatal("expected cycle detection error")
	}
	if !strings.Contains(err.Error(), "cycle") {
		t.Fatalf("unexpected move error: %v", err)
	}
}

func TestDeleteComponentRemovesSubtree(t *testing.T) {
	a := newApp()
	sessionID := mustCreateSession(t, a)

	parentRes, err := a.safeDispatch("ui.create_component", map[string]any{
		"session_id": sessionID,
		"type":       "section",
		"props":      map[string]any{"title": "Parent"},
		"block":      "main",
	})
	if err != nil {
		t.Fatalf("create parent failed: %v", err)
	}
	parentID := parentRes["component"].(*componentNode).ID

	childRes, err := a.safeDispatch("ui.create_component", map[string]any{
		"session_id": sessionID,
		"type":       "paragraph",
		"props":      map[string]any{"text": "Child"},
		"parent_id":  parentID,
	})
	if err != nil {
		t.Fatalf("create child failed: %v", err)
	}
	childID := childRes["component"].(*componentNode).ID

	_, err = a.safeDispatch("ui.delete_component", map[string]any{
		"session_id":   sessionID,
		"component_id": parentID,
	})
	if err != nil {
		t.Fatalf("delete component failed: %v", err)
	}

	_, err = a.safeDispatch("ui.get_component", map[string]any{
		"session_id":   sessionID,
		"component_id": parentID,
	})
	if err == nil {
		t.Fatal("expected parent to be deleted")
	}

	_, err = a.safeDispatch("ui.get_component", map[string]any{
		"session_id":   sessionID,
		"component_id": childID,
	})
	if err == nil {
		t.Fatal("expected child subtree to be deleted")
	}
}

func TestMoveComponentToMainPosition(t *testing.T) {
	a := newApp()
	sessionID := mustCreateSession(t, a)

	firstRes, err := a.safeDispatch("ui.create_component", map[string]any{
		"session_id": sessionID,
		"type":       "paragraph",
		"props":      map[string]any{"text": "first"},
		"block":      "main",
	})
	if err != nil {
		t.Fatalf("create first failed: %v", err)
	}
	firstID := firstRes["component"].(*componentNode).ID

	secondRes, err := a.safeDispatch("ui.create_component", map[string]any{
		"session_id": sessionID,
		"type":       "paragraph",
		"props":      map[string]any{"text": "second"},
		"block":      "main",
	})
	if err != nil {
		t.Fatalf("create second failed: %v", err)
	}
	secondID := secondRes["component"].(*componentNode).ID

	_, err = a.safeDispatch("ui.move_component", map[string]any{
		"session_id":   sessionID,
		"component_id": secondID,
		"new_block":    "main",
		"position":     0,
	})
	if err != nil {
		t.Fatalf("move failed: %v", err)
	}

	listRes, err := a.safeDispatch("ui.list_components", map[string]any{"session_id": sessionID})
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	blocks := listRes["blocks"].(map[string][]string)
	if len(blocks["main"]) != 2 {
		t.Fatalf("unexpected main block length: %d", len(blocks["main"]))
	}
	if blocks["main"][0] != secondID || blocks["main"][1] != firstID {
		t.Fatalf("unexpected order after move: %v", blocks["main"])
	}
}

func TestHeadSnippetValidationAndProviderValidation(t *testing.T) {
	a := newApp()
	sessionID := mustCreateSession(t, a)

	_, err := a.safeDispatch("page.set_head_snippet", map[string]any{
		"session_id": sessionID,
		"snippet":    "<script>alert(1)</script>",
	})
	if err == nil {
		t.Fatal("expected script head snippet to be rejected")
	}

	_, err = a.safeDispatch("page.set_head_snippet", map[string]any{
		"session_id": sessionID,
		"snippet":    "<meta name=\"x\" content=\"y\">",
	})
	if err != nil {
		t.Fatalf("meta snippet should be allowed: %v", err)
	}

	_, err = a.safeDispatch("page.set_css_provider", map[string]any{
		"session_id": sessionID,
		"provider":   "does-not-exist",
	})
	if err == nil {
		t.Fatal("expected invalid provider error")
	}
}

func TestRenderFullContainsComposedContent(t *testing.T) {
	a := newApp()
	sessionID := mustCreateSession(t, a)

	_, err := a.safeDispatch("ui.create_component", map[string]any{
		"session_id": sessionID,
		"type":       "heading",
		"props":      map[string]any{"level": 1, "text": "Top"},
		"block":      "header",
	})
	if err != nil {
		t.Fatalf("create header failed: %v", err)
	}
	_, err = a.safeDispatch("ui.create_component", map[string]any{
		"session_id": sessionID,
		"type":       "paragraph",
		"props":      map[string]any{"text": "Body"},
		"block":      "main",
	})
	if err != nil {
		t.Fatalf("create main failed: %v", err)
	}
	_, err = a.safeDispatch("ui.create_component", map[string]any{
		"session_id": sessionID,
		"type":       "paragraph",
		"props":      map[string]any{"text": "Foot"},
		"block":      "footer",
	})
	if err != nil {
		t.Fatalf("create footer failed: %v", err)
	}

	res, err := a.safeDispatch("render.full", map[string]any{"session_id": sessionID})
	if err != nil {
		t.Fatalf("render.full failed: %v", err)
	}
	html := res["html"].(string)
	for _, marker := range []string{"<html", "<header", "Top", "Body", "Foot"} {
		if !strings.Contains(html, marker) {
			t.Fatalf("full render missing marker %q", marker)
		}
	}
}

func TestRenderBlockInvalidBlock(t *testing.T) {
	a := newApp()
	sessionID := mustCreateSession(t, a)

	_, err := a.safeDispatch("render.block", map[string]any{
		"session_id": sessionID,
		"block":      "sidebar",
	})
	if err == nil {
		t.Fatal("expected invalid block error")
	}
}

func mustCreateSession(t *testing.T, a *app) string {
	t.Helper()
	res, err := a.safeDispatch("session.create", map[string]any{})
	if err != nil {
		t.Fatalf("session.create failed: %v", err)
	}
	s := res["session"].(*sessionState)
	if s.ID == "" {
		t.Fatal("session id is empty")
	}
	return s.ID
}

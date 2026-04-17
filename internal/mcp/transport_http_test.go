package mcp

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestAppWithSession(t *testing.T) (*app, string) {
	t.Helper()
	a := newApp()
	sessionID := mustCreateSession(t, a)
	return a, sessionID
}

func TestHandleSessionHTTPPage(t *testing.T) {
	a, sessionID := newTestAppWithSession(t)

	req := httptest.NewRequest(http.MethodGet, "/sessions/"+sessionID+"/page", nil)
	w := httptest.NewRecorder()
	a.handleSessionHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
	}
	body := w.Body.String()
	if !strings.Contains(body, "<html") {
		t.Fatalf("expected HTML response, got: %s", body)
	}
}

func TestHandleSessionHTTPBlock(t *testing.T) {
	a, sessionID := newTestAppWithSession(t)

	_, err := a.safeDispatch("ui.create_component", map[string]any{
		"session_id": sessionID,
		"type":       "paragraph",
		"props":      map[string]any{"text": "block-test"},
		"block":      "main",
	})
	if err != nil {
		t.Fatalf("create component failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/sessions/"+sessionID+"/blocks/main", nil)
	w := httptest.NewRecorder()
	a.handleSessionHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
	}
	if !strings.Contains(w.Body.String(), "block-test") {
		t.Fatalf("expected block content in response: %s", w.Body.String())
	}
}

func TestHandleSessionHTTPComponent(t *testing.T) {
	a, sessionID := newTestAppWithSession(t)

	createRes, err := a.safeDispatch("ui.create_component", map[string]any{
		"session_id": sessionID,
		"type":       "paragraph",
		"props":      map[string]any{"text": "component-fragment"},
		"block":      "main",
	})
	if err != nil {
		t.Fatalf("create component failed: %v", err)
	}
	componentID := createRes["component"].(*componentNode).ID

	req := httptest.NewRequest(http.MethodGet, "/sessions/"+sessionID+"/components/"+componentID, nil)
	w := httptest.NewRecorder()
	a.handleSessionHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
	}
	if !strings.Contains(w.Body.String(), "component-fragment") {
		t.Fatalf("expected component content in response: %s", w.Body.String())
	}
}

func TestHandleSessionHTTPMethodNotAllowed(t *testing.T) {
	a, sessionID := newTestAppWithSession(t)

	req := httptest.NewRequest(http.MethodPost, "/sessions/"+sessionID+"/page", nil)
	w := httptest.NewRecorder()
	a.handleSessionHTTP(w, req)

	if w.Result().StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Result().StatusCode)
	}
}

func TestHandleSessionHTTPNotFound(t *testing.T) {
	a := newApp()

	req := httptest.NewRequest(http.MethodGet, "/sessions/nonexistent/page", nil)
	w := httptest.NewRecorder()
	a.handleSessionHTTP(w, req)

	if w.Result().StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Result().StatusCode)
	}
}

func TestHandleSessionHTTPShortPath(t *testing.T) {
	a := newApp()

	req := httptest.NewRequest(http.MethodGet, "/sessions/only-two", nil)
	w := httptest.NewRecorder()
	a.handleSessionHTTP(w, req)

	if w.Result().StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404 for short path, got %d", w.Result().StatusCode)
	}
}

func TestHandleSessionHTTPUnknownPath(t *testing.T) {
	a, sessionID := newTestAppWithSession(t)

	req := httptest.NewRequest(http.MethodGet, "/sessions/"+sessionID+"/unknown-sub", nil)
	w := httptest.NewRecorder()
	a.handleSessionHTTP(w, req)

	if w.Result().StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404 for unknown sub-path, got %d", w.Result().StatusCode)
	}
}

func TestHandleSessionHTTPInvalidBlock(t *testing.T) {
	a, sessionID := newTestAppWithSession(t)

	req := httptest.NewRequest(http.MethodGet, "/sessions/"+sessionID+"/blocks/invalid-block", nil)
	w := httptest.NewRecorder()
	a.handleSessionHTTP(w, req)

	if w.Result().StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Result().StatusCode)
	}
}

func TestHandleSessionHTTPUnknownComponent(t *testing.T) {
	a, sessionID := newTestAppWithSession(t)

	req := httptest.NewRequest(http.MethodGet, "/sessions/"+sessionID+"/components/ghost-id", nil)
	w := httptest.NewRecorder()
	a.handleSessionHTTP(w, req)

	if w.Result().StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Result().StatusCode)
	}
}

func TestAsStringJsonNumber(t *testing.T) {
	num := json.Number("42")
	result := asString(num)
	if result != "42" {
		t.Fatalf("expected '42', got %q", result)
	}
}

func TestAsStringDefault(t *testing.T) {
	m := map[string]any{"key": "value", "empty": ""}
	if asStringDefault(m, "key", "default") != "value" {
		t.Fatal("expected 'value' for existing key")
	}
	if asStringDefault(m, "empty", "fallback") != "fallback" {
		t.Fatal("expected fallback for empty value")
	}
	if asStringDefault(m, "missing", "fallback") != "fallback" {
		t.Fatal("expected fallback for missing key")
	}
}

func TestAsIntAllTypes(t *testing.T) {
	if asInt(int32(5), 0) != 5 {
		t.Fatal("expected asInt(int32) == 5")
	}
	if asInt(int64(7), 0) != 7 {
		t.Fatal("expected asInt(int64) == 7")
	}
	if asInt(float64(3.0), 0) != 3 {
		t.Fatal("expected asInt(float64) == 3")
	}
	if asInt(json.Number("9"), 0) != 9 {
		t.Fatal("expected asInt(json.Number) == 9")
	}
	if asInt(json.Number("not-a-number"), 99) != 99 {
		t.Fatal("expected default for invalid json.Number")
	}
	if asInt("string", 42) != 42 {
		t.Fatal("expected default for string type")
	}
	if asInt(nil, 10) != 10 {
		t.Fatal("expected default for nil")
	}
}

func TestAsIntDefault(t *testing.T) {
	m := map[string]any{"level": 3}
	if asIntDefault(m, "level", 1) != 3 {
		t.Fatal("expected 3 for existing key")
	}
	if asIntDefault(m, "missing", 99) != 99 {
		t.Fatal("expected default for missing key")
	}
}

func TestValidateHeadSnippetStyleAllowed(t *testing.T) {
	err := validateHeadSnippet(`<style>body { color: red; }</style>`)
	if err != nil {
		t.Fatalf("expected style snippet to be allowed: %v", err)
	}
}

func TestValidateHeadSnippetEmpty(t *testing.T) {
	err := validateHeadSnippet("   ")
	if err == nil {
		t.Fatal("expected error for empty snippet")
	}
}

func TestValidateHeadSnippetInvalidTag(t *testing.T) {
	err := validateHeadSnippet(`<div class="x">content</div>`)
	if err == nil {
		t.Fatal("expected error for non-meta/link/style snippet")
	}
}

func TestBuildNodeRendersAllTypes(t *testing.T) {
	a := newApp()
	sessionID := mustCreateSession(t, a)

	renderableTypes := []struct {
		typeName string
		props    map[string]any
	}{
		{"badge", map[string]any{"text": "New", "tone": "success"}},
		{"blockquote", map[string]any{"text": "Quote", "cite": "Author"}},
		{"breadcrumbs", map[string]any{"items": []any{map[string]any{"label": "Home", "href": "/"}}}},
		{"card", map[string]any{"title": "Card", "body": "Body"}},
		{"checkbox", map[string]any{"name": "c", "value": "1", "label": "Check", "checked": true}},
		{"codeblock", map[string]any{"code": "code"}},
		{"descriptionlist", map[string]any{"items": []any{map[string]any{"term": "K", "definition": "V"}}}},
		{"heading", map[string]any{"level": 2, "text": "Title"}},
		{"horizontal_rule", map[string]any{}},
		{"image", map[string]any{"src": "/x.png", "alt": "Alt"}},
		{"input", map[string]any{"label": "L", "name": "n", "placeholder": "p"}},
		{"interactive_action", map[string]any{"text": "Go", "kind": "primary", "interaction": map[string]any{"get": "/data"}}},
		{"interactive_menu", map[string]any{"links": []any{map[string]any{"label": "X", "href": "/"}}}},
		{"list", map[string]any{"items": []any{"a", "b"}}},
		{"nav", map[string]any{"links": []any{map[string]any{"label": "H", "href": "/"}}}},
		{"ordered_list", map[string]any{"items": []any{"x"}}},
		{"pagination", map[string]any{"items": []any{map[string]any{"label": "1", "current": true}}}},
		{"radio", map[string]any{"name": "r", "value": "v", "label": "R", "checked": false}},
		{"section", map[string]any{"title": "S"}},
		{"select", map[string]any{"label": "L", "name": "s", "options": []any{map[string]any{"value": "a", "label": "A"}}}},
		{"table", map[string]any{"headers": []any{"A"}, "rows": []any{[]any{"v"}}}},
		{"textarea", map[string]any{"label": "Bio", "name": "bio", "placeholder": "ph"}},
		{"theme_toggle", map[string]any{}},
	}

	for _, tc := range renderableTypes {
		createRes, err := a.safeDispatch("ui.create_component", map[string]any{
			"session_id": sessionID,
			"type":       tc.typeName,
			"props":      tc.props,
			"block":      "main",
		})
		if err != nil {
			t.Fatalf("create_component type=%q failed: %v", tc.typeName, err)
		}

		componentID := createRes["component"].(*componentNode).ID
		_, err = a.safeDispatch("render.component", map[string]any{
			"session_id":   sessionID,
			"component_id": componentID,
		})
		if err != nil {
			t.Errorf("render.component type=%q failed: %v", tc.typeName, err)
		}
	}
}

func TestBuildNodeContainerTypes(t *testing.T) {
	a := newApp()
	sessionID := mustCreateSession(t, a)

	containerTypes := []string{"container", "container_fluid", "row", "fragment"}

	for _, typeName := range containerTypes {
		createRes, err := a.safeDispatch("ui.create_component", map[string]any{
			"session_id": sessionID,
			"type":       typeName,
			"props":      map[string]any{},
			"block":      "main",
		})
		if err != nil {
			t.Fatalf("create %s failed: %v", typeName, err)
		}
		componentID := createRes["component"].(*componentNode).ID

		// Add a child
		_, err = a.safeDispatch("ui.create_component", map[string]any{
			"session_id": sessionID,
			"type":       "paragraph",
			"props":      map[string]any{"text": "child"},
			"parent_id":  componentID,
		})
		if err != nil {
			t.Fatalf("create child for %s failed: %v", typeName, err)
		}

		_, err = a.safeDispatch("render.component", map[string]any{
			"session_id":   sessionID,
			"component_id": componentID,
		})
		if err != nil {
			t.Errorf("render %s failed: %v", typeName, err)
		}
	}
}

func TestBuildNodeGridWithItems(t *testing.T) {
	a := newApp()
	sessionID := mustCreateSession(t, a)

	gridRes, err := a.safeDispatch("ui.create_component", map[string]any{
		"session_id": sessionID,
		"type":       "grid",
		"props":      map[string]any{"columns_class": "grid-3"},
		"block":      "main",
	})
	if err != nil {
		t.Fatalf("create grid failed: %v", err)
	}
	gridID := gridRes["component"].(*componentNode).ID

	_, err = a.safeDispatch("ui.create_component", map[string]any{
		"session_id": sessionID,
		"type":       "grid_item",
		"props":      map[string]any{"span_class": "span-2"},
		"parent_id":  gridID,
	})
	if err != nil {
		t.Fatalf("create grid_item failed: %v", err)
	}

	_, err = a.safeDispatch("render.component", map[string]any{
		"session_id":   sessionID,
		"component_id": gridID,
	})
	if err != nil {
		t.Errorf("render grid failed: %v", err)
	}
}

func TestBuildNodeStackAndClusterAndColumn(t *testing.T) {
	a := newApp()
	sessionID := mustCreateSession(t, a)

	for _, typeName := range []string{"stack", "cluster", "column"} {
		createRes, err := a.safeDispatch("ui.create_component", map[string]any{
			"session_id": sessionID,
			"type":       typeName,
			"props":      map[string]any{"extra_class": "gap-sm", "size_class": "col-6"},
			"block":      "main",
		})
		if err != nil {
			t.Fatalf("create %s failed: %v", typeName, err)
		}
		componentID := createRes["component"].(*componentNode).ID

		_, err = a.safeDispatch("render.component", map[string]any{
			"session_id":   sessionID,
			"component_id": componentID,
		})
		if err != nil {
			t.Errorf("render %s failed: %v", typeName, err)
		}
	}
}

func TestBuildNodeHeroWithActions(t *testing.T) {
	a := newApp()
	sessionID := mustCreateSession(t, a)

	heroRes, err := a.safeDispatch("ui.create_component", map[string]any{
		"session_id": sessionID,
		"type":       "hero",
		"props":      map[string]any{"title": "Welcome", "subtitle": "Start here"},
		"block":      "main",
	})
	if err != nil {
		t.Fatalf("create hero failed: %v", err)
	}
	heroID := heroRes["component"].(*componentNode).ID

	_, err = a.safeDispatch("ui.create_component", map[string]any{
		"session_id": sessionID,
		"type":       "button",
		"props":      map[string]any{"text": "Start", "kind": "primary"},
		"parent_id":  heroID,
	})
	if err != nil {
		t.Fatalf("create hero button failed: %v", err)
	}

	renderRes, err := a.safeDispatch("render.component", map[string]any{
		"session_id":   sessionID,
		"component_id": heroID,
	})
	if err != nil {
		t.Fatalf("render hero failed: %v", err)
	}
	if !strings.Contains(renderRes["html"].(string), "Welcome") {
		t.Fatalf("expected 'Welcome' in hero output: %s", renderRes["html"])
	}
}

func TestBuildNodeArticleAndFieldsetAndDetails(t *testing.T) {
	a := newApp()
	sessionID := mustCreateSession(t, a)

	for _, typeName := range []string{"article", "fieldset", "details"} {
		createRes, err := a.safeDispatch("ui.create_component", map[string]any{
			"session_id": sessionID,
			"type":       typeName,
			"props":      map[string]any{"title": "T", "legend": "L", "summary": "S"},
			"block":      "main",
		})
		if err != nil {
			t.Fatalf("create %s failed: %v", typeName, err)
		}
		componentID := createRes["component"].(*componentNode).ID

		_, err = a.safeDispatch("ui.create_component", map[string]any{
			"session_id": sessionID,
			"type":       "paragraph",
			"props":      map[string]any{"text": "child text"},
			"parent_id":  componentID,
		})
		if err != nil {
			t.Fatalf("create child for %s failed: %v", typeName, err)
		}

		_, err = a.safeDispatch("render.component", map[string]any{
			"session_id":   sessionID,
			"component_id": componentID,
		})
		if err != nil {
			t.Errorf("render %s failed: %v", typeName, err)
		}
	}
}

func TestBuildNodeFormWithChildren(t *testing.T) {
	a := newApp()
	sessionID := mustCreateSession(t, a)

	formRes, err := a.safeDispatch("ui.create_component", map[string]any{
		"session_id": sessionID,
		"type":       "form",
		"props":      map[string]any{"action": "/save", "method": "post"},
		"block":      "main",
	})
	if err != nil {
		t.Fatalf("create form failed: %v", err)
	}
	formID := formRes["component"].(*componentNode).ID

	_, err = a.safeDispatch("ui.create_component", map[string]any{
		"session_id": sessionID,
		"type":       "input",
		"props":      map[string]any{"label": "Name", "name": "name", "placeholder": ""},
		"parent_id":  formID,
	})
	if err != nil {
		t.Fatalf("create input failed: %v", err)
	}

	renderRes, err := a.safeDispatch("render.component", map[string]any{
		"session_id":   sessionID,
		"component_id": formID,
	})
	if err != nil {
		t.Fatalf("render form failed: %v", err)
	}
	if !strings.Contains(renderRes["html"].(string), "<form") {
		t.Fatalf("expected form in output: %s", renderRes["html"])
	}
}

func TestBuildNodeUnsupportedType(t *testing.T) {
	a := newApp()
	sessionCreated := a.store.createSession("test", "stitch")

	_, err := a.store.updateSession(sessionCreated.ID, func(s *sessionState) error {
		s.Components["x"] = &componentNode{ID: "x", Type: "unsupported_xyz", Props: map[string]any{}, Children: []string{}}
		s.Blocks["main"] = append(s.Blocks["main"], "x")
		return nil
	})
	if err != nil {
		t.Fatalf("updateSession failed: %v", err)
	}

	session, err := a.store.getSession(sessionCreated.ID)
	if err != nil {
		t.Fatalf("get session failed: %v", err)
	}

	_, err = a.buildComponent(session, "x")
	if err == nil {
		t.Fatal("expected error for unsupported component type")
	}
	if !strings.Contains(err.Error(), "unsupported component type") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBuildChildrenMissingChild(t *testing.T) {
	a := newApp()
	sessionCreated := a.store.createSession("test", "stitch")

	_, err := a.store.updateSession(sessionCreated.ID, func(s *sessionState) error {
		s.Components["parent"] = &componentNode{
			ID:       "parent",
			Type:     "section",
			Props:    map[string]any{"title": "P"},
			Children: []string{"missing-child-id"},
		}
		s.Blocks["main"] = append(s.Blocks["main"], "parent")
		return nil
	})
	if err != nil {
		t.Fatalf("updateSession failed: %v", err)
	}

	session, err := a.store.getSession(sessionCreated.ID)
	if err != nil {
		t.Fatalf("get session failed: %v", err)
	}

	_, err = a.buildComponent(session, "parent")
	if err == nil {
		t.Fatal("expected error for missing child")
	}
	if !strings.Contains(err.Error(), "missing child component") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeepCopyAny(t *testing.T) {
	original := map[string]any{"key": "value", "num": 42.0}
	copied := deepCopyAny(original)
	m, ok := copied.(map[string]any)
	if !ok {
		t.Fatalf("expected map, got %T", copied)
	}
	if m["key"] != "value" {
		t.Fatalf("expected key='value', got %v", m["key"])
	}
}

func TestStateStoreCRUD(t *testing.T) {
	store := newStore()

	s := store.createSession("Test Title", "stitch")
	if s.ID == "" {
		t.Fatal("expected non-empty session ID")
	}
	if s.Page.Title != "Test Title" {
		t.Fatalf("expected title 'Test Title', got %q", s.Page.Title)
	}

	got, err := store.getSession(s.ID)
	if err != nil {
		t.Fatalf("getSession failed: %v", err)
	}
	if got.ID != s.ID {
		t.Fatalf("got unexpected ID: %s", got.ID)
	}

	updated, err := store.updateSession(s.ID, func(sess *sessionState) error {
		sess.Page.Title = "Updated"
		return nil
	})
	if err != nil {
		t.Fatalf("updateSession failed: %v", err)
	}
	if updated.Page.Title != "Updated" {
		t.Fatalf("expected 'Updated', got %q", updated.Page.Title)
	}

	deleted, err := store.deleteSession(s.ID)
	if err != nil {
		t.Fatalf("deleteSession failed: %v", err)
	}
	if !deleted {
		t.Fatal("expected deleted=true")
	}

	deleted2, err := store.deleteSession(s.ID)
	if err != nil {
		t.Fatalf("deleteSession for missing session failed: %v", err)
	}
	if deleted2 {
		t.Fatal("expected deleted=false for already-deleted session")
	}
}

func TestUpdateSessionError(t *testing.T) {
	store := newStore()
	_, err := store.updateSession("nonexistent", func(s *sessionState) error {
		return nil
	})
	if err == nil {
		t.Fatal("expected error for nonexistent session")
	}
}

func TestSessionCreateDefaultsWithEmptyStrings(t *testing.T) {
	store := newStore()
	s := store.createSession("  ", "  ")
	if s.Page.Title != "Stitch MCP UI" {
		t.Fatalf("expected default title for whitespace, got %q", s.Page.Title)
	}
	if s.Page.Provider != "stitch" {
		t.Fatalf("expected default provider for whitespace, got %q", s.Page.Provider)
	}
}

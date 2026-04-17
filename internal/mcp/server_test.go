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
		"snippet":    "<script src=\"https://cdn.example.com/app.js\"></script>",
	})
	if err != nil {
		t.Fatalf("expected external script head snippet to be allowed: %v", err)
	}

	_, err = a.safeDispatch("page.set_head_snippet", map[string]any{
		"session_id": sessionID,
		"snippet":    "<script>alert(1)</script>",
	})
	if err == nil {
		t.Fatal("expected inline script head snippet to be rejected")
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

func mustCreateApp(t *testing.T, a *app) string {
	t.Helper()
	res, err := a.safeDispatch("app.create", map[string]any{})
	if err != nil {
		t.Fatalf("app.create failed: %v", err)
	}
	created := res["app"].(*appState)
	if created.ID == "" {
		t.Fatal("app id is empty")
	}
	return created.ID
}

func TestAppCreateAndGet(t *testing.T) {
	a := newApp()
	appID := mustCreateApp(t, a)

	res, err := a.safeDispatch("app.get", map[string]any{"app_id": appID})
	if err != nil {
		t.Fatalf("app.get failed: %v", err)
	}
	created := res["app"].(*appState)
	if created.ID != appID {
		t.Fatalf("expected app ID %q, got %q", appID, created.ID)
	}
	if created.Name == "" {
		t.Fatal("expected non-empty app name")
	}
}

func TestAppRouteLifecycle(t *testing.T) {
	a := newApp()
	appID := mustCreateApp(t, a)
	sessionID := mustCreateSession(t, a)

	addRes, err := a.safeDispatch("app.add_route", map[string]any{
		"app_id":     appID,
		"path":       "/vehicles",
		"session_id": sessionID,
		"block":      "main",
	})
	if err != nil {
		t.Fatalf("app.add_route failed: %v", err)
	}
	route := addRes["route"].(appRoute)
	if route.Path != "/vehicles" || route.Block != "main" {
		t.Fatalf("unexpected route: %+v", route)
	}

	listRes, err := a.safeDispatch("app.list_routes", map[string]any{"app_id": appID})
	if err != nil {
		t.Fatalf("app.list_routes failed: %v", err)
	}
	routes := listRes["routes"].([]appRoute)
	if len(routes) != 1 || routes[0].ID != route.ID {
		t.Fatalf("unexpected routes after add: %+v", routes)
	}

	_, err = a.safeDispatch("app.update_route", map[string]any{
		"app_id":   appID,
		"route_id": route.ID,
		"name":     "fleet",
		"path":     "/fleet",
	})
	if err != nil {
		t.Fatalf("app.update_route failed: %v", err)
	}

	getRes, err := a.safeDispatch("app.get", map[string]any{"app_id": appID})
	if err != nil {
		t.Fatalf("app.get failed: %v", err)
	}
	updated := getRes["app"].(*appState)
	if len(updated.Routes) != 1 || updated.Routes[0].Path != "/fleet" || updated.Routes[0].Name != "fleet" {
		t.Fatalf("unexpected route after update: %+v", updated.Routes)
	}

	removeRes, err := a.safeDispatch("app.remove_route", map[string]any{"app_id": appID, "route_id": route.ID})
	if err != nil {
		t.Fatalf("app.remove_route failed: %v", err)
	}
	if removeRes["deleted"] != true {
		t.Fatalf("expected deleted=true, got %v", removeRes["deleted"])
	}
}

func TestAppSetShellAndValidate(t *testing.T) {
	a := newApp()
	appID := mustCreateApp(t, a)
	sessionID := mustCreateSession(t, a)

	_, err := a.safeDispatch("app.set_shell", map[string]any{"app_id": appID, "session_id": sessionID, "block": "main"})
	if err != nil {
		t.Fatalf("app.set_shell failed: %v", err)
	}
	_, err = a.safeDispatch("app.add_route", map[string]any{"app_id": appID, "path": "/", "session_id": sessionID})
	if err != nil {
		t.Fatalf("app.add_route failed: %v", err)
	}

	validRes, err := a.safeDispatch("app.validate", map[string]any{"app_id": appID})
	if err != nil {
		t.Fatalf("app.validate failed: %v", err)
	}
	if validRes["valid"] != true {
		t.Fatalf("expected valid app, got %v with warnings %v", validRes["valid"], validRes["warnings"])
	}

	_, err = a.safeDispatch("session.delete", map[string]any{"session_id": sessionID})
	if err != nil {
		t.Fatalf("session.delete failed: %v", err)
	}

	invalidRes, err := a.safeDispatch("app.validate", map[string]any{"app_id": appID})
	if err != nil {
		t.Fatalf("app.validate after delete failed: %v", err)
	}
	if invalidRes["valid"] != false {
		t.Fatalf("expected invalid app after deleting referenced session, got %v", invalidRes["valid"])
	}
	warnings := invalidRes["warnings"].([]map[string]any)
	if len(warnings) == 0 {
		t.Fatal("expected validation warnings after deleting referenced session")
	}
}

func TestAppRouteValidationRejectsBadInput(t *testing.T) {
	a := newApp()
	appID := mustCreateApp(t, a)
	sessionID := mustCreateSession(t, a)

	_, err := a.safeDispatch("app.add_route", map[string]any{"app_id": appID, "path": "vehicles", "session_id": sessionID})
	if err == nil {
		t.Fatal("expected invalid path error")
	}

	_, err = a.safeDispatch("app.add_route", map[string]any{"app_id": appID, "path": "/vehicles", "session_id": sessionID})
	if err != nil {
		t.Fatalf("app.add_route initial add failed: %v", err)
	}
	_, err = a.safeDispatch("app.add_route", map[string]any{"app_id": appID, "path": "/vehicles", "session_id": sessionID})
	if err == nil {
		t.Fatal("expected duplicate path error")
	}
	_, err = a.safeDispatch("app.set_shell", map[string]any{"app_id": appID, "session_id": sessionID, "block": "sidebar"})
	if err == nil {
		t.Fatal("expected invalid shell block error")
	}
}

func TestAppBuildAndGetBuild(t *testing.T) {
	a := newApp()
	appID := mustCreateApp(t, a)
	sessionID := mustCreateSession(t, a)

	_, err := a.safeDispatch("app.set_shell", map[string]any{"app_id": appID, "session_id": sessionID, "block": "main"})
	if err != nil {
		t.Fatalf("app.set_shell failed: %v", err)
	}
	_, err = a.safeDispatch("app.add_route", map[string]any{"app_id": appID, "path": "/overview", "session_id": sessionID, "name": "overview"})
	if err != nil {
		t.Fatalf("app.add_route failed: %v", err)
	}

	buildRes, err := a.safeDispatch("app.build", map[string]any{"app_id": appID, "target": "go-htmx"})
	if err != nil {
		t.Fatalf("app.build failed: %v", err)
	}
	build := buildRes["build"].(*appBuildState)
	if build.AppID != appID || build.Target != "go-htmx" {
		t.Fatalf("unexpected build metadata: %+v", build)
	}
	if build.Shell == nil || build.Shell.SessionID != sessionID {
		t.Fatalf("expected shell session in build, got %+v", build.Shell)
	}
	if len(build.Routes) != 1 || build.Routes[0].Path != "/overview" {
		t.Fatalf("unexpected build routes: %+v", build.Routes)
	}
	if len(build.HeadSnippets) == 0 {
		t.Fatal("expected build to inherit head snippets")
	}

	getBuildRes, err := a.safeDispatch("app.get_build", map[string]any{"build_id": build.ID})
	if err != nil {
		t.Fatalf("app.get_build failed: %v", err)
	}
	reloaded := getBuildRes["build"].(*appBuildState)
	if reloaded.ID != build.ID || len(reloaded.Routes) != 1 {
		t.Fatalf("unexpected reloaded build: %+v", reloaded)
	}

	_, err = a.safeDispatch("app.build", map[string]any{"app_id": appID, "target": "unknown-target"})
	if err == nil {
		t.Fatal("expected unsupported build target error")
	}
}

func TestAppExport(t *testing.T) {
	a := newApp()
	appID := mustCreateApp(t, a)
	sessionID := mustCreateSession(t, a)

	_, err := a.safeDispatch("app.add_route", map[string]any{"app_id": appID, "path": "/", "session_id": sessionID, "name": "home"})
	if err != nil {
		t.Fatalf("app.add_route failed: %v", err)
	}
	buildRes, err := a.safeDispatch("app.build", map[string]any{"app_id": appID, "target": "go-htmx"})
	if err != nil {
		t.Fatalf("app.build failed: %v", err)
	}
	build := buildRes["build"].(*appBuildState)

	exportRes, err := a.safeDispatch("app.export", map[string]any{"build_id": build.ID, "module_path": "example.com/generated/app"})
	if err != nil {
		t.Fatalf("app.export failed: %v", err)
	}
	project := exportRes["project"].(*exportedProject)
	if project.RunCommand != "go run ." {
		t.Fatalf("unexpected run command: %q", project.RunCommand)
	}
	if len(project.Files) == 0 {
		t.Fatal("expected exported files")
	}
	seen := map[string]string{}
	for _, file := range project.Files {
		seen[file.Path] = file.Content
	}
	for _, expected := range []string{"go.mod", "main.go", "internal/web/server.go", "stitch/routes.json", "stitch/manifest.json"} {
		if _, ok := seen[expected]; !ok {
			t.Fatalf("missing exported file %q", expected)
		}
	}
	if !strings.Contains(seen["go.mod"], "module example.com/generated/app") {
		t.Fatalf("unexpected go.mod content: %s", seen["go.mod"])
	}
	if !strings.Contains(seen["go.mod"], "require github.com/dmundt/stitch") {
		t.Fatalf("expected go.mod to require stitch module: %s", seen["go.mod"])
	}
	if !strings.Contains(seen["internal/web/server.go"], "HX-Request") {
		t.Fatalf("expected generated server to include HX branching: %s", seen["internal/web/server.go"])
	}
	if !strings.Contains(seen["internal/web/server.go"], "github.com/dmundt/stitch/css") || !strings.Contains(seen["internal/web/server.go"], "css.Assets") {
		t.Fatalf("expected generated server to use stitch css assets: %s", seen["internal/web/server.go"])
	}

	_, err = a.safeDispatch("app.export", map[string]any{"build_id": build.ID, "output_mode": "zip"})
	if err == nil {
		t.Fatal("expected unsupported output_mode error")
	}
}

func TestAppEmitProject(t *testing.T) {
	a := newApp()
	appID := mustCreateApp(t, a)
	sessionID := mustCreateSession(t, a)

	_, err := a.safeDispatch("app.add_route", map[string]any{"app_id": appID, "path": "/overview", "session_id": sessionID, "name": "overview"})
	if err != nil {
		t.Fatalf("app.add_route failed: %v", err)
	}
	res, err := a.safeDispatch("app.emit_project", map[string]any{"app_id": appID, "target": "go-htmx"})
	if err != nil {
		t.Fatalf("app.emit_project failed: %v", err)
	}
	if _, ok := res["build"].(*appBuildState); !ok {
		t.Fatalf("expected build in response, got %T", res["build"])
	}
	project, ok := res["project"].(*exportedProject)
	if !ok {
		t.Fatalf("expected project in response, got %T", res["project"])
	}
	if len(project.GeneratedFiles) == 0 {
		t.Fatal("expected generated file list")
	}
}

func TestSessionDelete(t *testing.T) {
	a := newApp()
	sessionID := mustCreateSession(t, a)

	res, err := a.safeDispatch("session.delete", map[string]any{"session_id": sessionID})
	if err != nil {
		t.Fatalf("session.delete failed: %v", err)
	}
	if res["deleted"] != true {
		t.Fatalf("expected deleted=true, got %v", res["deleted"])
	}

	_, err = a.safeDispatch("session.get", map[string]any{"session_id": sessionID})
	if err == nil {
		t.Fatal("expected error after deleting session")
	}
}

func TestSessionDeleteNonExistent(t *testing.T) {
	a := newApp()
	res, err := a.safeDispatch("session.delete", map[string]any{"session_id": "nonexistent"})
	if err != nil {
		t.Fatalf("session.delete for nonexistent session should not error: %v", err)
	}
	if res["deleted"] != false {
		t.Fatalf("expected deleted=false for nonexistent session, got %v", res["deleted"])
	}
}

func TestSessionDiagnostics(t *testing.T) {
	a := newApp()
	sessionID := mustCreateSession(t, a)

	_, err := a.safeDispatch("ui.create_component", map[string]any{
		"session_id": sessionID,
		"type":       "paragraph",
		"props": map[string]any{
			"text": "hello",
			"id":   "content",
			"attrs": map[string]any{
				"data-test": "main",
			},
		},
		"block": "main",
	})
	if err != nil {
		t.Fatalf("ui.create_component failed: %v", err)
	}

	res, err := a.safeDispatch("session.diagnostics", map[string]any{"session_id": sessionID})
	if err != nil {
		t.Fatalf("session.diagnostics failed: %v", err)
	}

	rendered, ok := res["rendered_attrs"].([]map[string]any)
	if !ok {
		t.Fatalf("expected rendered_attrs array, got %T", res["rendered_attrs"])
	}
	if len(rendered) == 0 {
		t.Fatal("expected rendered attrs to include at least one component")
	}

	headSummary, ok := res["head_summary"].(map[string]any)
	if !ok {
		t.Fatalf("expected head_summary map, got %T", res["head_summary"])
	}
	if _, ok := headSummary["scripts"]; !ok {
		t.Fatalf("expected scripts count in head_summary: %v", headSummary)
	}
}

func TestSessionReset(t *testing.T) {
	a := newApp()
	sessionID := mustCreateSession(t, a)

	_, err := a.safeDispatch("ui.create_component", map[string]any{
		"session_id": sessionID,
		"type":       "paragraph",
		"props":      map[string]any{"text": "before reset"},
		"block":      "main",
	})
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	_, err = a.safeDispatch("session.reset", map[string]any{"session_id": sessionID})
	if err != nil {
		t.Fatalf("session.reset failed: %v", err)
	}

	listRes, err := a.safeDispatch("ui.list_components", map[string]any{"session_id": sessionID})
	if err != nil {
		t.Fatalf("list after reset failed: %v", err)
	}
	components := listRes["components"].([]*componentNode)
	if len(components) != 0 {
		t.Fatalf("expected empty components after reset, got %d", len(components))
	}

	res, err := a.safeDispatch("session.get", map[string]any{"session_id": sessionID})
	if err != nil {
		t.Fatalf("session.get after reset failed: %v", err)
	}
	s := res["session"].(*sessionState)
	if len(s.Page.HeadSnippets) != len(defaultHeadSnippets()) {
		t.Fatalf("expected %d default head snippets after reset, got %d", len(defaultHeadSnippets()), len(s.Page.HeadSnippets))
	}
	for i, want := range defaultHeadSnippets() {
		if s.Page.HeadSnippets[i] != want {
			t.Fatalf("head snippet mismatch at %d: want %q, got %q", i, want, s.Page.HeadSnippets[i])
		}
	}
}

func TestRenderFullIncludesScriptSrcAndComponentID(t *testing.T) {
	a := newApp()
	sessionID := mustCreateSession(t, a)

	_, err := a.safeDispatch("page.set_head_snippet", map[string]any{
		"session_id": sessionID,
		"snippet":    "<script src=\"https://cdn.example.com/extra.js\"></script>",
	})
	if err != nil {
		t.Fatalf("page.set_head_snippet failed: %v", err)
	}

	_, err = a.safeDispatch("ui.create_component", map[string]any{
		"session_id": sessionID,
		"type":       "paragraph",
		"props":      map[string]any{"text": "Body", "id": "main-content"},
		"block":      "main",
	})
	if err != nil {
		t.Fatalf("ui.create_component failed: %v", err)
	}

	renderRes, err := a.safeDispatch("render.full", map[string]any{"session_id": sessionID})
	if err != nil {
		t.Fatalf("render.full failed: %v", err)
	}
	html := renderRes["html"].(string)
	if !strings.Contains(html, `<script src="https://cdn.example.com/extra.js"></script>`) {
		t.Fatalf("expected external script tag in rendered html: %s", html)
	}
	if !strings.Contains(html, `<p id="main-content">Body</p>`) {
		t.Fatalf("expected paragraph id in rendered html: %s", html)
	}
}

func TestRenderFullRendersNestedLinkItemMetadata(t *testing.T) {
	a := newApp()
	sessionID := mustCreateSession(t, a)

	_, err := a.safeDispatch("ui.create_component", map[string]any{
		"session_id": sessionID,
		"type":       "nav",
		"props": map[string]any{
			"links": []any{map[string]any{
				"label": "Overview",
				"href":  "/overview",
				"id":    "probe-link-overview",
				"class": "probe-link",
				"attrs": map[string]any{
					"data-route": "overview",
				},
			}},
		},
		"block": "main",
	})
	if err != nil {
		t.Fatalf("create nav failed: %v", err)
	}

	_, err = a.safeDispatch("ui.create_component", map[string]any{
		"session_id": sessionID,
		"type":       "breadcrumbs",
		"props": map[string]any{
			"items": []any{map[string]any{
				"label": "Home",
				"href":  "/",
				"id":    "crumb-home",
				"class": "crumb-link",
				"attrs": map[string]any{
					"data-crumb": "home",
				},
			}},
		},
		"block": "main",
	})
	if err != nil {
		t.Fatalf("create breadcrumbs failed: %v", err)
	}

	renderRes, err := a.safeDispatch("render.full", map[string]any{"session_id": sessionID})
	if err != nil {
		t.Fatalf("render.full failed: %v", err)
	}
	html := renderRes["html"].(string)
	for _, want := range []string{
		`id="probe-link-overview"`,
		`class="probe-link"`,
		`data-route="overview"`,
		`id="crumb-home"`,
		`class="crumb-link"`,
		`data-crumb="home"`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("expected rendered html to include %q: %s", want, html)
		}
	}
}

func TestSessionGet(t *testing.T) {
	a := newApp()
	sessionID := mustCreateSession(t, a)

	res, err := a.safeDispatch("session.get", map[string]any{"session_id": sessionID})
	if err != nil {
		t.Fatalf("session.get failed: %v", err)
	}
	s := res["session"].(*sessionState)
	if s.ID != sessionID {
		t.Fatalf("expected session ID %q, got %q", sessionID, s.ID)
	}
}

func TestSessionGetNonExistent(t *testing.T) {
	a := newApp()
	_, err := a.safeDispatch("session.get", map[string]any{"session_id": "ghost"})
	if err == nil {
		t.Fatal("expected error for nonexistent session")
	}
}

func TestProvidersListTool(t *testing.T) {
	a := newApp()
	res, err := a.safeDispatch("providers.list", map[string]any{})
	if err != nil {
		t.Fatalf("providers.list failed: %v", err)
	}
	providers, ok := res["providers"].([]string)
	if !ok {
		t.Fatalf("expected []string for providers, got %T", res["providers"])
	}
	if len(providers) == 0 {
		t.Fatal("expected at least one provider")
	}
}

func TestPageSetMeta(t *testing.T) {
	a := newApp()
	sessionID := mustCreateSession(t, a)

	res, err := a.safeDispatch("page.set_meta", map[string]any{
		"session_id": sessionID,
		"title":      "New Title",
		"lang":       "de",
	})
	if err != nil {
		t.Fatalf("page.set_meta failed: %v", err)
	}
	s := res["session"].(*sessionState)
	if s.Page.Title != "New Title" {
		t.Fatalf("expected title 'New Title', got %q", s.Page.Title)
	}
	if s.Page.Lang != "de" {
		t.Fatalf("expected lang 'de', got %q", s.Page.Lang)
	}
}

func TestPageSetMetaEmptyFieldsNotOverwrite(t *testing.T) {
	a := newApp()
	sessionID := mustCreateSession(t, a)

	_, err := a.safeDispatch("page.set_meta", map[string]any{
		"session_id": sessionID,
		"title":      "",
		"lang":       "",
	})
	if err != nil {
		t.Fatalf("page.set_meta failed: %v", err)
	}

	res, err := a.safeDispatch("session.get", map[string]any{"session_id": sessionID})
	if err != nil {
		t.Fatalf("session.get failed: %v", err)
	}
	s := res["session"].(*sessionState)
	if s.Page.Title == "" {
		t.Fatal("expected title to remain non-empty when set to empty string")
	}
}

func TestSchemaListComponentTypes(t *testing.T) {
	a := newApp()
	res, err := a.safeDispatch("schema.list_component_types", map[string]any{})
	if err != nil {
		t.Fatalf("schema.list_component_types failed: %v", err)
	}
	types, ok := res["types"].(map[string]map[string]any)
	if !ok {
		t.Fatalf("expected map for types, got %T", res["types"])
	}
	if _, ok := types["paragraph"]; !ok {
		t.Fatal("expected 'paragraph' in schema types")
	}
}

func TestSchemaGetComponentFields(t *testing.T) {
	a := newApp()
	res, err := a.safeDispatch("schema.get_component_fields", map[string]any{"type": "heading"})
	if err != nil {
		t.Fatalf("schema.get_component_fields failed: %v", err)
	}
	if res["type"] != "heading" {
		t.Fatalf("expected type 'heading', got %v", res["type"])
	}
	schema := res["schema"].(map[string]any)
	props := schema["props"].([]string)
	hasID := false
	hasAttrs := false
	for _, p := range props {
		if p == "id" {
			hasID = true
		}
		if p == "attrs" {
			hasAttrs = true
		}
	}
	if !hasID {
		t.Fatalf("expected heading schema props to include id, got %v", props)
	}
	if !hasAttrs {
		t.Fatalf("expected heading schema props to include attrs, got %v", props)
	}
}

func TestSchemaGetComponentFieldsIncludesIDForEmptyPropsType(t *testing.T) {
	a := newApp()
	res, err := a.safeDispatch("schema.get_component_fields", map[string]any{"type": "theme_toggle"})
	if err != nil {
		t.Fatalf("schema.get_component_fields failed: %v", err)
	}
	schema := res["schema"].(map[string]any)
	props := schema["props"].([]string)
	hasID := false
	hasAttrs := false
	for _, p := range props {
		if p == "id" {
			hasID = true
		}
		if p == "attrs" {
			hasAttrs = true
		}
	}
	if !hasID || !hasAttrs {
		t.Fatalf("expected theme_toggle schema props to include id and attrs, got %v", props)
	}
}

func TestToolCreateComponentSanitizesAttrsAndReturnsWarnings(t *testing.T) {
	a := newApp()
	sessionID := mustCreateSession(t, a)

	res, err := a.safeDispatch("ui.create_component", map[string]any{
		"session_id": sessionID,
		"type":       "paragraph",
		"props": map[string]any{
			"text": "Body",
			"attrs": map[string]any{
				"data-test": "ok",
				"hx-target": "#page-content",
				"onclick":   "alert(1)",
				"style":     "color:red",
			},
		},
		"block": "main",
	})
	if err != nil {
		t.Fatalf("ui.create_component failed: %v", err)
	}

	node := res["component"].(*componentNode)
	if got := node.Attrs["data-test"]; got != "ok" {
		t.Fatalf("expected data-test attr to survive, got %q", got)
	}
	if got := node.Attrs["hx-target"]; got != "#page-content" {
		t.Fatalf("expected hx-target attr to survive, got %q", got)
	}
	if _, exists := node.Attrs["onclick"]; exists {
		t.Fatalf("expected onclick attr to be dropped, attrs=%v", node.Attrs)
	}
	warnings, ok := res["warnings"].([]map[string]any)
	if !ok {
		t.Fatalf("expected warnings payload, got %T", res["warnings"])
	}
	if len(warnings) < 2 {
		t.Fatalf("expected at least 2 warnings, got %v", warnings)
	}

	renderRes, err := a.safeDispatch("render.full", map[string]any{"session_id": sessionID})
	if err != nil {
		t.Fatalf("render.full failed: %v", err)
	}
	html := renderRes["html"].(string)
	if !strings.Contains(html, `data-test="ok"`) {
		t.Fatalf("expected rendered html to include data-test attr: %s", html)
	}
	if !strings.Contains(html, `hx-target="#page-content"`) {
		t.Fatalf("expected rendered html to include hx-target attr: %s", html)
	}
	if strings.Contains(html, `onclick=`) {
		t.Fatalf("expected blocked onclick attr to be absent: %s", html)
	}
}

func TestSchemaGetComponentFieldsUnknownType(t *testing.T) {
	a := newApp()
	_, err := a.safeDispatch("schema.get_component_fields", map[string]any{"type": "nonexistent"})
	if err == nil {
		t.Fatal("expected error for unknown component type")
	}
}

func TestToolUpdateComponent(t *testing.T) {
	a := newApp()
	sessionID := mustCreateSession(t, a)

	createRes, err := a.safeDispatch("ui.create_component", map[string]any{
		"session_id": sessionID,
		"type":       "paragraph",
		"props":      map[string]any{"text": "original"},
		"block":      "main",
	})
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	componentID := createRes["component"].(*componentNode).ID

	_, err = a.safeDispatch("ui.update_component", map[string]any{
		"session_id":   sessionID,
		"component_id": componentID,
		"props":        map[string]any{"text": "updated"},
	})
	if err != nil {
		t.Fatalf("ui.update_component failed: %v", err)
	}

	getRes, err := a.safeDispatch("ui.get_component", map[string]any{
		"session_id":   sessionID,
		"component_id": componentID,
	})
	if err != nil {
		t.Fatalf("ui.get_component after update failed: %v", err)
	}
	node := getRes["component"].(*componentNode)
	if node.Props["text"] != "updated" {
		t.Fatalf("expected updated text, got %v", node.Props["text"])
	}
}

func TestToolUpdateComponentChangeType(t *testing.T) {
	a := newApp()
	sessionID := mustCreateSession(t, a)

	createRes, err := a.safeDispatch("ui.create_component", map[string]any{
		"session_id": sessionID,
		"type":       "paragraph",
		"props":      map[string]any{"text": "hi"},
		"block":      "main",
	})
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	componentID := createRes["component"].(*componentNode).ID

	_, err = a.safeDispatch("ui.update_component", map[string]any{
		"session_id":   sessionID,
		"component_id": componentID,
		"type":         "heading",
		"props":        map[string]any{"level": 2, "text": "hi"},
	})
	if err != nil {
		t.Fatalf("update type failed: %v", err)
	}
}

func TestToolUpdateComponentUnknownID(t *testing.T) {
	a := newApp()
	sessionID := mustCreateSession(t, a)

	_, err := a.safeDispatch("ui.update_component", map[string]any{
		"session_id":   sessionID,
		"component_id": "nonexistent-id",
		"props":        map[string]any{},
	})
	if err == nil {
		t.Fatal("expected error for unknown component_id")
	}
}

func TestToolUpdateComponentInvalidType(t *testing.T) {
	a := newApp()
	sessionID := mustCreateSession(t, a)

	createRes, err := a.safeDispatch("ui.create_component", map[string]any{
		"session_id": sessionID,
		"type":       "paragraph",
		"props":      map[string]any{"text": "hi"},
		"block":      "main",
	})
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	componentID := createRes["component"].(*componentNode).ID

	_, err = a.safeDispatch("ui.update_component", map[string]any{
		"session_id":   sessionID,
		"component_id": componentID,
		"type":         "invalid_type_xyz",
		"props":        map[string]any{},
	})
	if err == nil {
		t.Fatal("expected error for unsupported component type in update")
	}
}

func TestDispatchUnknownTool(t *testing.T) {
	a := newApp()
	_, err := a.safeDispatch("unknown.tool", map[string]any{})
	if err == nil {
		t.Fatal("expected error for unknown tool")
	}
	if !strings.Contains(err.Error(), "unknown tool") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCanonicalTypeNormalization(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"view", "section"},
		{"panel", "article"},
		{"text", "paragraph"},
		{"action", "button"},
		{"menu", "nav"},
		{"workspace", "appshell"},
		{"status", "alert"},
		{"toolbar", "cluster"},
		{"datagrid", "table"},
		{"heading", "heading"},
		{"Heading", "heading"},
		{"my-component", "my_component"},
		{"my component", "my_component"},
	}
	for _, tc := range cases {
		got := canonicalType(tc.input)
		if got != tc.want {
			t.Errorf("canonicalType(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestNormalizeToolName(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"session_create", "session.create"},
		{"app_create", "app.create"},
		{"app_get", "app.get"},
		{"app_list_routes", "app.list_routes"},
		{"app_add_route", "app.add_route"},
		{"app_update_route", "app.update_route"},
		{"app_remove_route", "app.remove_route"},
		{"app_set_shell", "app.set_shell"},
		{"app_validate", "app.validate"},
		{"app_build", "app.build"},
		{"app_get_build", "app.get_build"},
		{"app_export", "app.export"},
		{"app_emit_project", "app.emit_project"},
		{"session_get", "session.get"},
		{"session_reset", "session.reset"},
		{"session_delete", "session.delete"},
		{"session_diagnostics", "session.diagnostics"},
		{"providers_list", "providers.list"},
		{"page_set_meta", "page.set_meta"},
		{"page_set_css_provider", "page.set_css_provider"},
		{"page_set_head_snippet", "page.set_head_snippet"},
		{"schema_list_component_types", "schema.list_component_types"},
		{"schema_get_component_fields", "schema.get_component_fields"},
		{"ui_create_component", "ui.create_component"},
		{"ui_update_component", "ui.update_component"},
		{"ui_delete_component", "ui.delete_component"},
		{"ui_move_component", "ui.move_component"},
		{"ui_get_component", "ui.get_component"},
		{"ui_list_components", "ui.list_components"},
		{"render_full", "render.full"},
		{"render_block", "render.block"},
		{"render_component", "render.component"},
		{"already.dotted", "already.dotted"},
	}
	for _, tc := range cases {
		got := normalizeToolName(tc.input)
		if got != tc.want {
			t.Errorf("normalizeToolName(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestAsBool(t *testing.T) {
	if asBool(true) != true {
		t.Fatal("expected asBool(true) == true")
	}
	if asBool(false) != false {
		t.Fatal("expected asBool(false) == false")
	}
	if asBool("true") != false {
		t.Fatal("expected asBool(string) == false")
	}
	if asBool(nil) != false {
		t.Fatal("expected asBool(nil) == false")
	}
	if asBool(1) != false {
		t.Fatal("expected asBool(int) == false")
	}
}

func TestAsBoolDefault(t *testing.T) {
	m := map[string]any{"yes": true, "no": false}
	if !asBoolDefault(m, "yes", false) {
		t.Fatal("expected true for key 'yes'")
	}
	if asBoolDefault(m, "no", true) {
		t.Fatal("expected false for key 'no'")
	}
	if !asBoolDefault(m, "missing", true) {
		t.Fatal("expected default true for missing key")
	}
	if asBoolDefault(m, "missing", false) {
		t.Fatal("expected default false for missing key")
	}
}

func TestAsStringSlice(t *testing.T) {
	// nil input
	if s := asStringSlice(nil); len(s) != 0 {
		t.Fatalf("expected empty slice, got %v", s)
	}
	// []any input
	result := asStringSlice([]any{"a", "b", "c"})
	if len(result) != 3 || result[0] != "a" {
		t.Fatalf("unexpected result: %v", result)
	}
	// []string input
	result2 := asStringSlice([]string{"x", "y"})
	if len(result2) != 2 || result2[0] != "x" {
		t.Fatalf("unexpected result: %v", result2)
	}
	// filters empty strings
	result3 := asStringSlice([]any{"a", "", "b"})
	if len(result3) != 2 {
		t.Fatalf("expected filtered empty strings: %v", result3)
	}
}

func TestAsStringMatrix(t *testing.T) {
	// nil input
	if m := asStringMatrix(nil); len(m) != 0 {
		t.Fatalf("expected empty matrix, got %v", m)
	}
	// []any input
	result := asStringMatrix([]any{[]any{"a", "b"}, []any{"c", "d"}})
	if len(result) != 2 || len(result[0]) != 2 || result[0][0] != "a" {
		t.Fatalf("unexpected matrix result: %v", result)
	}
	// [][]string input
	result2 := asStringMatrix([][]string{{"x", "y"}, {"z"}})
	if len(result2) != 2 || result2[0][0] != "x" {
		t.Fatalf("unexpected matrix result: %v", result2)
	}
}

func TestAsMapParsesJSONString(t *testing.T) {
	m := asMap(`{"text":"hello","level":2}`)
	if got := asString(m["text"]); got != "hello" {
		t.Fatalf("expected parsed text to be hello, got %q", got)
	}
	if got := asInt(m["level"], 0); got != 2 {
		t.Fatalf("expected parsed level to be 2, got %d", got)
	}
}

func TestMCPToolsPropsAreObjectSchema(t *testing.T) {
	tools := mcpTools()
	findTool := func(name string) map[string]any {
		for _, tool := range tools {
			if toolName, _ := tool["name"].(string); toolName == name {
				return tool
			}
		}
		return nil
	}

	for _, toolName := range []string{"ui_create_component", "ui_update_component"} {
		tool := findTool(toolName)
		if tool == nil {
			t.Fatalf("missing tool definition for %s", toolName)
		}
		inputSchema, ok := tool["inputSchema"].(map[string]any)
		if !ok {
			t.Fatalf("tool %s missing input schema", toolName)
		}
		properties, ok := inputSchema["properties"].(map[string]map[string]any)
		if !ok {
			t.Fatalf("tool %s properties have unexpected type %T", toolName, inputSchema["properties"])
		}
		propsSchema, ok := properties["props"]
		if !ok {
			t.Fatalf("tool %s missing props schema", toolName)
		}
		if got := asString(propsSchema["type"]); got != "object" {
			t.Fatalf("tool %s props type must be object, got %q", toolName, got)
		}
	}
}

func TestAsNavLinks(t *testing.T) {
	// nil input
	if links := asNavLinks(nil); len(links) != 0 {
		t.Fatalf("expected empty nav links, got %v", links)
	}
	// valid input
	links := asNavLinks([]any{
		map[string]any{"label": "Home", "href": "/", "id": "home-link", "class": "nav-link", "attrs": map[string]any{"data-route": "home", "onclick": "x"}},
		map[string]any{"label": "About", "href": "/about"},
	})
	if len(links) != 2 || links[0].Label != "Home" || links[0].Href != "/" {
		t.Fatalf("unexpected nav links: %v", links)
	}
	if links[0].ID != "home-link" || links[0].Class != "nav-link" {
		t.Fatalf("expected nested id/class fields, got %+v", links[0])
	}
	if links[0].Attrs["data-route"] != "home" {
		t.Fatalf("expected nested data-route attr, got %+v", links[0].Attrs)
	}
	if _, ok := links[0].Attrs["onclick"]; ok {
		t.Fatalf("expected blocked nested attr to be removed, got %+v", links[0].Attrs)
	}
}

func TestAsBreadcrumbItems(t *testing.T) {
	if items := asBreadcrumbItems(nil); len(items) != 0 {
		t.Fatalf("expected empty breadcrumb items, got %v", items)
	}
	items := asBreadcrumbItems([]any{
		map[string]any{"label": "Home", "href": "/", "id": "crumb-home", "class": "crumb-link", "attrs": map[string]any{"data-crumb": "home"}, "current": false},
		map[string]any{"label": "Docs", "current": true},
	})
	if len(items) != 2 || items[0].Label != "Home" || !items[1].Current {
		t.Fatalf("unexpected breadcrumb items: %v", items)
	}
	if items[0].ID != "crumb-home" || items[0].Class != "crumb-link" || items[0].Attrs["data-crumb"] != "home" {
		t.Fatalf("expected nested breadcrumb metadata, got %+v", items[0])
	}
}

func TestAsDescriptionItems(t *testing.T) {
	if items := asDescriptionItems(nil); len(items) != 0 {
		t.Fatalf("expected empty description items, got %v", items)
	}
	items := asDescriptionItems([]any{
		map[string]any{"term": "Key", "definition": "Value"},
	})
	if len(items) != 1 || items[0].Term != "Key" || items[0].Definition != "Value" {
		t.Fatalf("unexpected description items: %v", items)
	}
}

func TestAsPageItems(t *testing.T) {
	if items := asPageItems(nil); len(items) != 0 {
		t.Fatalf("expected empty page items, got %v", items)
	}
	items := asPageItems([]any{
		map[string]any{"label": "Prev", "disabled": true},
		map[string]any{"label": "1", "current": true, "href": "/?page=1", "id": "page-1", "class": "page-link", "attrs": map[string]any{"data-page": "1"}},
	})
	if len(items) != 2 || !items[0].Disabled || !items[1].Current {
		t.Fatalf("unexpected page items: %v", items)
	}
	if items[1].ID != "page-1" || items[1].Class != "page-link" || items[1].Attrs["data-page"] != "1" {
		t.Fatalf("expected nested page metadata, got %+v", items[1])
	}
}

func TestAsSelectOptions(t *testing.T) {
	if opts := asSelectOptions(nil); len(opts) != 0 {
		t.Fatalf("expected empty select options, got %v", opts)
	}
	opts := asSelectOptions([]any{
		map[string]any{"value": "a", "label": "A", "selected": true},
		map[string]any{"value": "b", "label": "B"},
	})
	if len(opts) != 2 || opts[0].Value != "a" || !opts[0].Selected || opts[1].Selected {
		t.Fatalf("unexpected select options: %v", opts)
	}
}

func TestAsInteraction(t *testing.T) {
	ix := asInteraction(map[string]any{
		"get":      "/items",
		"target":   "#list",
		"swap":     "innerHTML",
		"trigger":  "click",
		"push_url": "/items",
		"boost":    true,
		"post":     "/save",
		"put":      "/update",
		"delete":   "/remove",
		"select":   "#result",
	})
	if ix.Get != "/items" {
		t.Fatalf("expected Get='/items', got %q", ix.Get)
	}
	if ix.Target != "#list" {
		t.Fatalf("expected Target='#list', got %q", ix.Target)
	}
	if !ix.Boost {
		t.Fatal("expected Boost=true")
	}
}

func TestAsInteractionNilInput(t *testing.T) {
	ix := asInteraction(nil)
	if ix.Get != "" || ix.Target != "" || ix.Boost {
		t.Fatalf("expected zero interaction for nil input: %+v", ix)
	}
}

func TestAsInteractiveMenuLinks(t *testing.T) {
	if links := asInteractiveMenuLinks(nil); len(links) != 0 {
		t.Fatalf("expected empty interactive menu links, got %v", links)
	}
	links := asInteractiveMenuLinks([]any{
		map[string]any{
			"label": "Home",
			"href":  "/",
			"id":    "menu-home",
			"class": "menu-link",
			"attrs": map[string]any{"data-route": "home"},
			"interaction": map[string]any{
				"get":    "/home",
				"target": "main",
			},
		},
	})
	if len(links) != 1 || links[0].Label != "Home" || links[0].Interaction.Get != "/home" {
		t.Fatalf("unexpected interactive menu links: %v", links)
	}
	if links[0].ID != "menu-home" || links[0].Class != "menu-link" || links[0].Attrs["data-route"] != "home" {
		t.Fatalf("expected nested interactive menu metadata, got %+v", links[0])
	}
}

func TestJsonText(t *testing.T) {
	text := jsonText(map[string]any{"key": "value"})
	if !strings.Contains(text, "key") || !strings.Contains(text, "value") {
		t.Fatalf("unexpected jsonText output: %s", text)
	}
}

func TestEscapeForText(t *testing.T) {
	result := escapeForText("<script>alert(1)</script>")
	if strings.Contains(result, "<script>") {
		t.Fatalf("expected escaped HTML, got: %s", result)
	}
	if !strings.Contains(result, "&lt;script&gt;") {
		t.Fatalf("expected &lt;script&gt; in result: %s", result)
	}
}

func TestToPrettyBytes(t *testing.T) {
	b := toPrettyBytes(map[string]any{"a": 1})
	if len(b) == 0 {
		t.Fatal("expected non-empty pretty bytes")
	}
	if !strings.Contains(string(b), "a") {
		t.Fatalf("unexpected pretty bytes: %s", b)
	}
}

func TestBuildNodeAllComponentTypes(t *testing.T) {
	a := newApp()
	sessionID := mustCreateSession(t, a)

	componentTypes := []struct {
		typeName string
		props    map[string]any
	}{
		{"action", map[string]any{"text": "Go", "kind": "primary"}},
		{"alert", map[string]any{"text": "Info", "tone": "info"}},
		{"badge", map[string]any{"text": "New", "tone": "success"}},
		{"blockquote", map[string]any{"text": "Quote", "cite": "Author"}},
		{"breadcrumbs", map[string]any{"items": []any{map[string]any{"label": "Home", "href": "/"}}}},
		{"card", map[string]any{"title": "Card", "body": "Content"}},
		{"checkbox", map[string]any{"name": "cb", "value": "1", "label": "Check", "checked": true}},
		{"codeblock", map[string]any{"code": "go test"}},
		{"container", map[string]any{}},
		{"container_fluid", map[string]any{}},
		{"descriptionlist", map[string]any{"items": []any{map[string]any{"term": "K", "definition": "V"}}}},
		{"details", map[string]any{"summary": "More"}},
		{"fieldset", map[string]any{"legend": "Group"}},
		{"form", map[string]any{"action": "/", "method": "post"}},
		{"fragment", map[string]any{}},
		{"grid", map[string]any{"columns_class": "grid-3"}},
		{"grid_item", map[string]any{"span_class": "span-2"}},
		{"heading", map[string]any{"level": 2, "text": "Title"}},
		{"hero", map[string]any{"title": "Hero", "subtitle": "Sub"}},
		{"horizontal_rule", map[string]any{}},
		{"image", map[string]any{"src": "/img.png", "alt": "Image"}},
		{"input", map[string]any{"label": "Name", "name": "name", "placeholder": "Enter name"}},
		{"interactive_action", map[string]any{"text": "Load", "kind": "primary", "interaction": map[string]any{"get": "/data", "target": "#out"}}},
		{"interactive_menu", map[string]any{"links": []any{map[string]any{"label": "Home", "href": "/", "interaction": map[string]any{"get": "/"}}}}},
		{"list", map[string]any{"items": []any{"a", "b"}}},
		{"nav", map[string]any{"links": []any{map[string]any{"label": "Home", "href": "/"}}}},
		{"ordered_list", map[string]any{"items": []any{"x", "y"}}},
		{"pagination", map[string]any{"items": []any{map[string]any{"label": "1", "current": true}}}},
		{"paragraph", map[string]any{"text": "Body text"}},
		{"radio", map[string]any{"name": "r", "value": "v", "label": "Radio", "checked": false}},
		{"row", map[string]any{}},
		{"section", map[string]any{"title": "Section"}},
		{"select", map[string]any{"label": "Pick", "name": "s", "options": []any{map[string]any{"value": "a", "label": "A"}}}},
		{"stack", map[string]any{"extra_class": "gap-md"}},
		{"table", map[string]any{"headers": []any{"A", "B"}, "rows": []any{[]any{"1", "2"}}}},
		{"textarea", map[string]any{"label": "Bio", "name": "bio", "placeholder": "Enter bio"}},
		{"theme_toggle", map[string]any{}},
		{"status", map[string]any{"text": "OK", "tone": "success"}},
		{"text", map[string]any{"text": "Text"}},
	}

	for _, tc := range componentTypes {
		_, err := a.safeDispatch("ui.create_component", map[string]any{
			"session_id": sessionID,
			"type":       tc.typeName,
			"props":      tc.props,
			"block":      "main",
		})
		if err != nil {
			t.Errorf("create_component type=%q failed: %v", tc.typeName, err)
		}
	}
}

func TestBuildNodeAppShellAndSidebarWithChildren(t *testing.T) {
	a := newApp()
	sessionID := mustCreateSession(t, a)

	shellRes, err := a.safeDispatch("ui.create_component", map[string]any{
		"session_id": sessionID,
		"type":       "appshell",
		"props":      map[string]any{},
		"block":      "main",
	})
	if err != nil {
		t.Fatalf("create appshell failed: %v", err)
	}
	shellID := shellRes["component"].(*componentNode).ID

	_, err = a.safeDispatch("ui.create_component", map[string]any{
		"session_id": sessionID,
		"type":       "nav",
		"props":      map[string]any{"links": []any{}},
		"parent_id":  shellID,
	})
	if err != nil {
		t.Fatalf("create sidebar nav failed: %v", err)
	}

	_, err = a.safeDispatch("ui.create_component", map[string]any{
		"session_id": sessionID,
		"type":       "paragraph",
		"props":      map[string]any{"text": "Content"},
		"parent_id":  shellID,
	})
	if err != nil {
		t.Fatalf("create content failed: %v", err)
	}

	res, err := a.safeDispatch("render.full", map[string]any{"session_id": sessionID})
	if err != nil {
		t.Fatalf("render.full with appshell failed: %v", err)
	}
	if !strings.Contains(res["html"].(string), "Content") {
		t.Fatalf("expected 'Content' in rendered output: %s", res["html"])
	}
}

func TestBuildNodeSidebarLayout(t *testing.T) {
	a := newApp()
	sessionID := mustCreateSession(t, a)

	layoutRes, err := a.safeDispatch("ui.create_component", map[string]any{
		"session_id": sessionID,
		"type":       "sidebar_layout",
		"props":      map[string]any{},
		"block":      "main",
	})
	if err != nil {
		t.Fatalf("create sidebar_layout failed: %v", err)
	}
	layoutID := layoutRes["component"].(*componentNode).ID

	for _, childType := range []string{"nav", "section"} {
		_, err = a.safeDispatch("ui.create_component", map[string]any{
			"session_id": sessionID,
			"type":       childType,
			"props":      map[string]any{"links": []any{}, "title": "S"},
			"parent_id":  layoutID,
		})
		if err != nil {
			t.Fatalf("create %s child failed: %v", childType, err)
		}
	}

	res, err := a.safeDispatch("render.full", map[string]any{"session_id": sessionID})
	if err != nil {
		t.Fatalf("render.full with sidebar_layout failed: %v", err)
	}
	if !strings.Contains(res["html"].(string), "layout-sidebar") {
		t.Fatalf("expected layout-sidebar in output: %s", res["html"])
	}
}

func TestBuildNodeSplitWithChildren(t *testing.T) {
	a := newApp()
	sessionID := mustCreateSession(t, a)

	splitRes, err := a.safeDispatch("ui.create_component", map[string]any{
		"session_id": sessionID,
		"type":       "split",
		"props":      map[string]any{},
		"block":      "main",
	})
	if err != nil {
		t.Fatalf("create split failed: %v", err)
	}
	splitID := splitRes["component"].(*componentNode).ID

	for _, text := range []string{"Primary", "Secondary"} {
		_, err = a.safeDispatch("ui.create_component", map[string]any{
			"session_id": sessionID,
			"type":       "paragraph",
			"props":      map[string]any{"text": text},
			"parent_id":  splitID,
		})
		if err != nil {
			t.Fatalf("create split child failed: %v", err)
		}
	}

	res, err := a.safeDispatch("render.full", map[string]any{"session_id": sessionID})
	if err != nil {
		t.Fatalf("render.full with split failed: %v", err)
	}
	if !strings.Contains(res["html"].(string), "Primary") {
		t.Fatalf("expected 'Primary' in output: %s", res["html"])
	}
}

func TestBuildNodeToolbarAlias(t *testing.T) {
	a := newApp()
	sessionID := mustCreateSession(t, a)

	_, err := a.safeDispatch("ui.create_component", map[string]any{
		"session_id": sessionID,
		"type":       "toolbar",
		"props":      map[string]any{},
		"block":      "main",
	})
	if err != nil {
		t.Fatalf("create toolbar failed: %v", err)
	}
}

func TestCreateComponentInvalidBlock(t *testing.T) {
	a := newApp()
	sessionID := mustCreateSession(t, a)

	_, err := a.safeDispatch("ui.create_component", map[string]any{
		"session_id": sessionID,
		"type":       "paragraph",
		"props":      map[string]any{"text": "x"},
		"block":      "invalid-block",
	})
	if err == nil {
		t.Fatal("expected error for invalid block")
	}
}

func TestCreateComponentUnsupportedType(t *testing.T) {
	a := newApp()
	sessionID := mustCreateSession(t, a)

	_, err := a.safeDispatch("ui.create_component", map[string]any{
		"session_id": sessionID,
		"type":       "nonexistent_component_type",
		"props":      map[string]any{},
		"block":      "main",
	})
	if err == nil {
		t.Fatal("expected error for unsupported component type")
	}
}

func TestCreateComponentUnknownParentID(t *testing.T) {
	a := newApp()
	sessionID := mustCreateSession(t, a)

	_, err := a.safeDispatch("ui.create_component", map[string]any{
		"session_id": sessionID,
		"type":       "paragraph",
		"props":      map[string]any{"text": "x"},
		"parent_id":  "nonexistent-parent",
	})
	if err == nil {
		t.Fatal("expected error for unknown parent_id")
	}
}

func TestMoveComponentToParent(t *testing.T) {
	a := newApp()
	sessionID := mustCreateSession(t, a)

	sectionRes, err := a.safeDispatch("ui.create_component", map[string]any{
		"session_id": sessionID,
		"type":       "section",
		"props":      map[string]any{"title": "Container"},
		"block":      "main",
	})
	if err != nil {
		t.Fatalf("create section failed: %v", err)
	}
	sectionID := sectionRes["component"].(*componentNode).ID

	paraRes, err := a.safeDispatch("ui.create_component", map[string]any{
		"session_id": sessionID,
		"type":       "paragraph",
		"props":      map[string]any{"text": "hello"},
		"block":      "main",
	})
	if err != nil {
		t.Fatalf("create para failed: %v", err)
	}
	paraID := paraRes["component"].(*componentNode).ID

	_, err = a.safeDispatch("ui.move_component", map[string]any{
		"session_id":    sessionID,
		"component_id":  paraID,
		"new_parent_id": sectionID,
	})
	if err != nil {
		t.Fatalf("move to parent failed: %v", err)
	}

	getRes, err := a.safeDispatch("ui.get_component", map[string]any{
		"session_id":   sessionID,
		"component_id": sectionID,
	})
	if err != nil {
		t.Fatalf("get section failed: %v", err)
	}
	section := getRes["component"].(*componentNode)
	found := false
	for _, childID := range section.Children {
		if childID == paraID {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected paragraph to be child of section after move")
	}
}

func TestMoveComponentUnknownComponent(t *testing.T) {
	a := newApp()
	sessionID := mustCreateSession(t, a)

	_, err := a.safeDispatch("ui.move_component", map[string]any{
		"session_id":   sessionID,
		"component_id": "ghost-id",
		"new_block":    "main",
	})
	if err == nil {
		t.Fatal("expected error for unknown component_id in move")
	}
}

func TestMoveComponentUnknownNewParent(t *testing.T) {
	a := newApp()
	sessionID := mustCreateSession(t, a)

	paraRes, err := a.safeDispatch("ui.create_component", map[string]any{
		"session_id": sessionID,
		"type":       "paragraph",
		"props":      map[string]any{"text": "x"},
		"block":      "main",
	})
	if err != nil {
		t.Fatalf("create para failed: %v", err)
	}
	paraID := paraRes["component"].(*componentNode).ID

	_, err = a.safeDispatch("ui.move_component", map[string]any{
		"session_id":    sessionID,
		"component_id":  paraID,
		"new_parent_id": "ghost-parent",
	})
	if err == nil {
		t.Fatal("expected error for unknown new_parent_id")
	}
}

func TestMoveComponentInvalidBlock(t *testing.T) {
	a := newApp()
	sessionID := mustCreateSession(t, a)

	paraRes, err := a.safeDispatch("ui.create_component", map[string]any{
		"session_id": sessionID,
		"type":       "paragraph",
		"props":      map[string]any{"text": "x"},
		"block":      "main",
	})
	if err != nil {
		t.Fatalf("create para failed: %v", err)
	}
	paraID := paraRes["component"].(*componentNode).ID

	_, err = a.safeDispatch("ui.move_component", map[string]any{
		"session_id":   sessionID,
		"component_id": paraID,
		"new_block":    "invalid-block",
	})
	if err == nil {
		t.Fatal("expected error for invalid new_block in move")
	}
}

func TestDeleteComponentUnknownID(t *testing.T) {
	a := newApp()
	sessionID := mustCreateSession(t, a)

	_, err := a.safeDispatch("ui.delete_component", map[string]any{
		"session_id":   sessionID,
		"component_id": "ghost-id",
	})
	if err == nil {
		t.Fatal("expected error for unknown component_id in delete")
	}
}

func TestRenderComponentUnknownID(t *testing.T) {
	a := newApp()
	sessionID := mustCreateSession(t, a)

	_, err := a.safeDispatch("render.component", map[string]any{
		"session_id":   sessionID,
		"component_id": "ghost-id",
	})
	if err == nil {
		t.Fatal("expected error for unknown component_id in render.component")
	}
}

func TestRenderFullUnknownProvider(t *testing.T) {
	a := newApp()
	res, err := a.safeDispatch("session.create", map[string]any{"provider": "custom-unknown"})
	if err != nil {
		t.Fatalf("session.create failed: %v", err)
	}
	sessionID := res["session"].(*sessionState).ID

	_, err = a.safeDispatch("render.full", map[string]any{"session_id": sessionID})
	if err != nil {
		t.Fatalf("render.full with unknown provider should fall back to stitch: %v", err)
	}
}

func TestInsertAtEdgeCases(t *testing.T) {
	list := []string{"a", "b", "c"}
	insertAt(&list, "x", 0)
	if list[0] != "x" {
		t.Fatalf("expected 'x' at position 0, got %v", list)
	}

	list2 := []string{"a", "b"}
	insertAt(&list2, "z", 10)
	if list2[len(list2)-1] != "z" {
		t.Fatalf("expected 'z' appended for out-of-range position, got %v", list2)
	}

	list3 := []string{"a", "b"}
	insertAt(&list3, "m", -1)
	if list3[len(list3)-1] != "m" {
		t.Fatalf("expected 'm' appended for negative position, got %v", list3)
	}
}

func TestCreateComponentDefaultsToMain(t *testing.T) {
	a := newApp()
	sessionID := mustCreateSession(t, a)

	_, err := a.safeDispatch("ui.create_component", map[string]any{
		"session_id": sessionID,
		"type":       "paragraph",
		"props":      map[string]any{"text": "default block"},
	})
	if err != nil {
		t.Fatalf("create_component without block/parent failed: %v", err)
	}

	listRes, err := a.safeDispatch("ui.list_components", map[string]any{"session_id": sessionID})
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	blocks := listRes["blocks"].(map[string][]string)
	if len(blocks["main"]) == 0 {
		t.Fatal("expected component to be added to main block by default")
	}
}

func TestCreateComponentWithPositionZero(t *testing.T) {
	a := newApp()
	sessionID := mustCreateSession(t, a)

	_, err := a.safeDispatch("ui.create_component", map[string]any{
		"session_id": sessionID,
		"type":       "paragraph",
		"props":      map[string]any{"text": "first"},
		"block":      "main",
	})
	if err != nil {
		t.Fatalf("create first failed: %v", err)
	}

	secondRes, err := a.safeDispatch("ui.create_component", map[string]any{
		"session_id": sessionID,
		"type":       "paragraph",
		"props":      map[string]any{"text": "insert at 0"},
		"block":      "main",
		"position":   0,
	})
	if err != nil {
		t.Fatalf("create at position 0 failed: %v", err)
	}
	insertedID := secondRes["component"].(*componentNode).ID

	listRes, err := a.safeDispatch("ui.list_components", map[string]any{"session_id": sessionID})
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	blocks := listRes["blocks"].(map[string][]string)
	if blocks["main"][0] != insertedID {
		t.Fatalf("expected inserted component at index 0, got %v", blocks["main"])
	}
}

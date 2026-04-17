package mcp

import (
	"testing"
)

func TestRegisteredToolsIncludesCoreCalls(t *testing.T) {
	tools, err := registeredTools()
	if err != nil {
		t.Fatalf("registeredTools failed: %v", err)
	}
	if len(tools) == 0 {
		t.Fatal("expected non-empty tool registry")
	}

	seen := map[string]bool{}
	for _, tool := range tools {
		if tool.Name == "" {
			t.Fatal("tool name must not be empty")
		}
		if tool.InputSchema == nil {
			t.Fatalf("tool %s missing input schema", tool.Name)
		}
		seen[tool.Name] = true
	}

	for _, expected := range []string{"session_create", "session_diagnostics", "ui_create_component", "render_full"} {
		if !seen[expected] {
			t.Fatalf("missing expected tool registration: %s", expected)
		}
	}
}

func TestNewSDKServerBuilds(t *testing.T) {
	a := newApp()
	server, err := a.newSDKServer()
	if err != nil {
		t.Fatalf("newSDKServer failed: %v", err)
	}
	if server == nil {
		t.Fatal("expected non-nil server")
	}
}

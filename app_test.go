package stitch_test

import (
	"html/template"
	"strings"
	"testing"

	stitch "github.com/dmundt/stitch"
	"github.com/dmundt/stitch/render"
)

func TestNewWithValidProvider(t *testing.T) {
	app, err := stitch.New("stitch")
	if err != nil {
		t.Fatalf("New with valid provider failed: %v", err)
	}
	if app == nil {
		t.Fatal("expected non-nil app")
	}
	if app.Provider == nil {
		t.Fatal("expected non-nil provider")
	}
	if app.Provider.Name() != "stitch" {
		t.Fatalf("expected provider name 'stitch', got %q", app.Provider.Name())
	}
}

func TestNewWithInvalidProvider(t *testing.T) {
	_, err := stitch.New("does-not-exist")
	if err == nil {
		t.Fatal("expected error for unknown provider")
	}
	if !strings.Contains(err.Error(), "unknown provider") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestNewWithNoneProvider(t *testing.T) {
	app, err := stitch.New("none")
	if err != nil {
		t.Fatalf("New with 'none' provider failed: %v", err)
	}
	if app.Provider.Name() != "none" {
		t.Fatalf("expected provider name 'none', got %q", app.Provider.Name())
	}
}

func TestAppNewPage(t *testing.T) {
	app, _ := stitch.New("stitch")
	page := app.NewPage("My Title")
	if page == nil {
		t.Fatal("expected non-nil page")
	}
	if page.Title != "My Title" {
		t.Fatalf("expected title 'My Title', got %q", page.Title)
	}
}

func TestAppRender(t *testing.T) {
	app, _ := stitch.New("stitch")
	page := app.NewPage("Render Test")
	page.Main(template.HTML("<p>hello</p>"))

	html, err := app.Render(page)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}
	if !strings.Contains(html, "Render Test") {
		t.Fatalf("missing title in rendered output: %s", html)
	}
	if !strings.Contains(html, "<p>hello</p>") {
		t.Fatalf("missing content in rendered output: %s", html)
	}
}

func TestAppRenderNilPage(t *testing.T) {
	app, _ := stitch.New("stitch")
	_, err := app.Render(nil)
	if err == nil {
		t.Fatal("expected error for nil page")
	}
}

func TestAddWithNilPage(t *testing.T) {
	err := stitch.Add(nil, "main", template.HTML("<p>x</p>"))
	if err != nil {
		t.Fatalf("Add with nil page should not error on valid block: %v", err)
	}
}

func TestAddWithNilPageInvalidBlock(t *testing.T) {
	err := stitch.Add(nil, "bogus", template.HTML("<p>x</p>"))
	if err == nil {
		t.Fatal("expected error for invalid block with nil page")
	}
}

func TestAddWithNilComposer(t *testing.T) {
	page := &render.Page{Title: "no composer"}
	err := stitch.Add(page, "main", template.HTML("<p>x</p>"))
	if err != nil {
		t.Fatalf("Add with nil composer should not error on valid block: %v", err)
	}
}

func TestAddWithNilComposerInvalidBlock(t *testing.T) {
	page := &render.Page{Title: "no composer"}
	err := stitch.Add(page, "invalid-block", template.HTML("<p>x</p>"))
	if err == nil {
		t.Fatal("expected error for invalid block with nil composer")
	}
}

func TestAddAppendsFragment(t *testing.T) {
	app, _ := stitch.New("stitch")
	page := app.NewPage("Test")
	err := stitch.Add(page, "main", template.HTML("<p>added</p>"))
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}
	html, err := app.Render(page)
	if err != nil {
		t.Fatalf("render failed: %v", err)
	}
	if !strings.Contains(html, "<p>added</p>") {
		t.Fatalf("expected added fragment in output: %s", html)
	}
}

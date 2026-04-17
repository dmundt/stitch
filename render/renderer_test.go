package render

import (
	"html/template"
	"strings"
	"testing"

	"github.com/dmundt/stitch/css"
	stitchtpl "github.com/dmundt/stitch/template"
)

type testComponent struct{ html string }

func TestDocumentTemplateSourceEmbedded(t *testing.T) {
	if !strings.Contains(documentTemplateSource, "<!doctype html>") {
		t.Fatalf("expected embedded document template source, got %q", documentTemplateSource)
	}
}

func TestGuiAliasesForComponents(t *testing.T) {
	p := NewWindow("Component Workspace")
	p.TopBarComponent(testComponent{html: `<nav>top component</nav>`})
	p.ContentComponent(testComponent{html: `<section>content component</section>`})
	p.StatusBarComponent(testComponent{html: `<small>status component</small>`})

	html, err := p.Render(css.Stitch())
	if err != nil {
		t.Fatalf("render failed: %v", err)
	}

	for _, want := range []string{"top component", "content component", "status component"} {
		if !strings.Contains(html, want) {
			t.Fatalf("expected %q in output", want)
		}
	}
}

func TestGuiAliasesForPageComposition(t *testing.T) {
	p := NewWindow("Workbench")
	p.TopBar(template.HTML("<div>top</div>"))
	p.Content(template.HTML("<div>content</div>"))
	p.StatusBar(template.HTML("<div>status</div>"))

	html, err := p.Render(css.Stitch())
	if err != nil {
		t.Fatalf("render failed: %v", err)
	}

	for _, want := range []string{"Workbench", "top", "content", "status"} {
		if !strings.Contains(html, want) {
			t.Fatalf("expected %q in output", want)
		}
	}
}

func TestPageComponentComposition(t *testing.T) {
	provider := css.NewStaticProvider("test", template.HTML(`<link rel="stylesheet" href="https://example.com/component.css">`))
	p := NewPage("Component Chain")
	p.WithHeadRaw(`<meta name="component" content="true">`)
	p.HeaderComponent(testComponent{html: `<nav>header component</nav>`})
	p.MainComponent(testComponent{html: `<section>main component</section>`})
	p.FooterComponent(testComponent{html: `<small>footer component</small>`})

	html, err := p.Render(provider)

	if err != nil {
		t.Fatalf("component chain render failed: %v", err)
	}
	for _, want := range []string{
		"Component Chain",
		`name="component"`,
		"header component",
		"main component",
		"footer component",
		"component.css",
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("expected %q in output", want)
		}
	}
}

func TestPageComposition(t *testing.T) {
	provider := css.NewStaticProvider("test", template.HTML(`<link rel="stylesheet" href="https://example.com/chain.css">`))
	p := NewPage("Chain Test")
	p.WithHead(template.HTML(`<meta name="x" content="y">`))
	p.Header(template.HTML("<nav>top</nav>"))
	p.Main(template.HTML("<p>body content</p>"))
	p.Footer(template.HTML("<p>footer text</p>"))

	html, err := p.Render(provider)

	if err != nil {
		t.Fatalf("fluent render failed: %v", err)
	}
	for _, want := range []string{
		"<title>Chain Test</title>",
		`name="x"`,
		"chain.css",
		"top",
		"body content",
		"footer text",
		`class="stitch-page"`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("page composition missing %q in output", want)
		}
	}
}

func TestRenderNilPage(t *testing.T) {
	if _, err := Render(nil, css.None()); err == nil {
		t.Fatal("expected nil page error")
	}
}

func TestRenderNilComposer(t *testing.T) {
	p := &Page{Title: "no composer"}
	_, err := Render(p, css.None())
	if err == nil {
		t.Fatal("expected nil composer error")
	}
	if !strings.Contains(err.Error(), "composer is nil") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestRenderNilProvider(t *testing.T) {
	p := NewPage("nil provider test")
	p.Main(template.HTML("<p>hello</p>"))
	html, err := Render(p, nil)
	if err != nil {
		t.Fatalf("render with nil provider should use None(): %v", err)
	}
	if !strings.Contains(html, "nil provider test") {
		t.Fatalf("missing title in output: %s", html)
	}
}

func TestAddHead(t *testing.T) {
	p := NewPage("head test")
	p.AddHead(template.HTML(`<meta name="x" content="y">`))
	html, err := Render(p, css.None())
	if err != nil {
		t.Fatalf("render failed: %v", err)
	}
	if !strings.Contains(html, `name="x"`) {
		t.Fatalf("expected head fragment in output: %s", html)
	}
}

func TestAddHeadRaw(t *testing.T) {
	p := NewPage("head raw test")
	p.AddHeadRaw(`<meta name="raw" content="yes">`)
	html, err := Render(p, css.None())
	if err != nil {
		t.Fatalf("render failed: %v", err)
	}
	if !strings.Contains(html, `name="raw"`) {
		t.Fatalf("expected raw head fragment in output: %s", html)
	}
}

func TestHeaderComponentNilNoOp(t *testing.T) {
	p := NewPage("nil component test")
	p.HeaderComponent(nil)
	html, err := Render(p, css.None())
	if err != nil {
		t.Fatalf("render failed: %v", err)
	}
	if !strings.Contains(html, "nil component test") {
		t.Fatalf("missing title: %s", html)
	}
}

func TestMainComponentNilNoOp(t *testing.T) {
	p := NewPage("nil main test")
	p.MainComponent(nil)
	_, err := Render(p, css.None())
	if err != nil {
		t.Fatalf("render with nil MainComponent failed: %v", err)
	}
}

func TestFooterComponentNilNoOp(t *testing.T) {
	p := NewPage("nil footer test")
	p.FooterComponent(nil)
	_, err := Render(p, css.None())
	if err != nil {
		t.Fatalf("render with nil FooterComponent failed: %v", err)
	}
}

func TestContentComponentNilNoOp(t *testing.T) {
	p := NewPage("nil content test")
	p.ContentComponent(nil)
	_, err := Render(p, css.None())
	if err != nil {
		t.Fatalf("render with nil ContentComponent failed: %v", err)
	}
}

func TestStatusBarComponentNilNoOp(t *testing.T) {
	p := NewPage("nil statusbar test")
	p.StatusBarComponent(nil)
	_, err := Render(p, css.None())
	if err != nil {
		t.Fatalf("render with nil StatusBarComponent failed: %v", err)
	}
}

func TestTopBarComponentNilNoOp(t *testing.T) {
	p := NewPage("nil topbar test")
	p.TopBarComponent(nil)
	_, err := Render(p, css.None())
	if err != nil {
		t.Fatalf("render with nil TopBarComponent failed: %v", err)
	}
}

func TestRenderWithProviderAndBlocks(t *testing.T) {
	p := NewPage("My Page")
	_ = p.Composer.AddString(stitchtpl.BlockHeader, "<h1>X</h1>")
	_ = p.Composer.AddString(stitchtpl.BlockMain, "hello")
	_ = p.Composer.AddString(stitchtpl.BlockFooter, "bye")

	html, err := Render(p, css.NewStaticProvider("test", template.HTML(`<link rel="stylesheet" href="https://example.com/test.css">`)))
	if err != nil {
		t.Fatalf("render failed: %v", err)
	}

	if !strings.Contains(html, "<title>My Page</title>") {
		t.Fatalf("missing title: %s", html)
	}
	if !strings.Contains(html, "example.com/test.css") {
		t.Fatalf("missing provider link: %s", html)
	}
	if !(strings.Index(html, "<header>") < strings.Index(html, "<main>") && strings.Index(html, "<main>") < strings.Index(html, "<footer>")) {
		t.Fatalf("unexpected section order: %s", html)
	}
	if !strings.Contains(html, "&lt;h1&gt;X&lt;/h1&gt;") {
		t.Fatalf("expected escaped block content: %s", html)
	}
}

func (c testComponent) HTML() string { return c.html }

package ui

import (
	"strings"
	"testing"
)

type fixedIDRootComponent struct{}

func (fixedIDRootComponent) HTML() string {
	return `<div id="existing">x</div>`
}

func assertContains(t *testing.T, html, want string) {
	t.Helper()
	if !strings.Contains(html, want) {
		t.Fatalf("expected %q in:\n%s", want, html)
	}
}

func TestContentComponents(t *testing.T) {
	assertContains(t, NewHeading(1, "H1").HTML(), "<h1>H1</h1>")
	assertContains(t, NewHeading(2, "H2").HTML(), "<h2>H2</h2>")
	assertContains(t, NewHeading(3, "H3").HTML(), "<h3>H3</h3>")
	assertContains(t, NewHeading(6, "H6").HTML(), "<h6>H6</h6>")
	assertContains(t, NewHeading(0, "lo").HTML(), "<h1>lo</h1>")
	assertContains(t, NewHeading(9, "hi").HTML(), "<h6>hi</h6>")
	assertContains(t, NewParagraph("text").HTML(), "<p>text</p>")
	assertContains(t, NewCodeBlock("<tag>").HTML(), "&lt;tag&gt;")
	assertContains(t, NewBlockquote("quote", "src").HTML(), "<blockquote>")
	assertContains(t, NewBlockquote("quote", "src").HTML(), "<cite>src</cite>")
	assertContains(t, NewBlockquote("q", "").HTML(), "<p>q</p>")
	assertContains(t, NewImage("/a.png", "alt").HTML(), `src="/a.png"`)
	assertContains(t, NewHorizontalRule().HTML(), "<hr>")
}

func TestEscaping(t *testing.T) {
	assertContains(t, NewHeading(2, "Hello <world>").HTML(), "<h2>Hello &lt;world&gt;</h2>")
	assertContains(t, NewCodeBlock("<tag>").HTML(), "&lt;tag&gt;")
	assertContains(t, NewAlert("Heads up", "warning").HTML(), "alert-warning")
}

func TestFeedbackComponents(t *testing.T) {
	assertContains(t, NewAlert("msg", "info").HTML(), `class="alert alert-info"`)
	assertContains(t, NewAlert("msg", "").HTML(), `class="alert alert-info"`)
	assertContains(t, NewBadge("new", "success").HTML(), `class="badge badge-success"`)
	assertContains(t, NewBadge("x", "").HTML(), `class="badge badge-default"`)
	assertContains(t, NewButton("Go", "primary").HTML(), `btn-primary`)
	assertContains(t, NewButton("x", "").HTML(), `btn-default`)
	assertContains(t, NewCard("T", "B").HTML(), `class="card"`)
}

func TestFormComponents(t *testing.T) {
	form := NewForm("/submit", "post", []Component{
		NewFieldset("Profile", []Component{
			NewInput("Name", "name", "Jane"),
			NewButton("Save", "primary"),
		}),
	})
	html := form.HTML()
	assertContains(t, html, `action="/submit"`)
	assertContains(t, html, "<legend>Profile</legend>")
	assertContains(t, html, "btn-primary")

	assertContains(t, NewCheckbox("ch", "1", "Check", true).HTML(), `checked`)
	assertContains(t, NewCheckbox("ch", "1", "Check", false).HTML(), `type="checkbox"`)
	assertContains(t, NewRadio("r", "v", "Radio", true).HTML(), `checked`)
	assertContains(t, NewTextArea("L", "ta", "ph").HTML(), `<textarea`)
	sel := NewSelect("Pick", "s", []SelectOption{{Value: "a", Label: "A", Selected: true}})
	assertContains(t, sel.HTML(), `<select`)
	assertContains(t, sel.HTML(), `selected`)
}

func TestGridComponents(t *testing.T) {
	grid := NewGrid("grid-3", []Component{
		NewGridItem("span-2", []Component{NewParagraph("wide")}),
		NewGridItem("", []Component{NewParagraph("narrow")}),
	})
	html := grid.HTML()
	assertContains(t, html, `class="grid grid-3"`)
	assertContains(t, html, `class="grid-item span-2"`)
	assertContains(t, html, `class="grid-item"`)
}

func TestGridAdd(t *testing.T) {
	item := NewGridItem("", []Component{NewParagraph("a")})
	item.Add([]Component{NewParagraph("b")})
	g := NewGrid("", []Component{})
	g.Add([]Component{item})
	html := g.HTML()
	assertContains(t, html, `<p>a</p>`)
	assertContains(t, html, `<p>b</p>`)
}

func TestRowAndColumnComponents(t *testing.T) {
	left := NewColumn("col-8", []Component{NewParagraph("left")})
	right := NewColumn("col-4", []Component{NewParagraph("right")})
	row := NewRow([]Component{left, right})
	html := row.HTML()
	assertContains(t, html, `class="row"`)
	assertContains(t, html, `class="column col-8"`)
	assertContains(t, html, `class="column col-4"`)
}

func TestRowAndColumnAdd(t *testing.T) {
	col := NewColumn("", []Component{NewParagraph("a")})
	col.Add([]Component{NewParagraph("b")})
	row := NewRow([]Component{})
	row.Add([]Component{col})
	html := row.HTML()
	assertContains(t, html, `<p>a</p>`)
	assertContains(t, html, `<p>b</p>`)
}

func TestLayoutComponents(t *testing.T) {
	hero := NewHero("Title", "Sub", []Component{NewButton("Go", "primary")})
	assertContains(t, hero.HTML(), `class="hero"`)
	assertContains(t, hero.HTML(), "<h1>Title</h1>")
	assertContains(t, hero.HTML(), "<p>Sub</p>")

	stack := NewStack("gap-md", []Component{NewParagraph("One"), NewParagraph("Two")})
	assertContains(t, stack.HTML(), `class="stack gap-md"`)

	cluster := NewCluster("justify-between", []Component{NewBadge("new", "success")})
	assertContains(t, cluster.HTML(), `class="cluster justify-between"`)

	split := NewSplit(NewCard("L", "left"), NewCard("R", "right"))
	assertContains(t, split.HTML(), `class="split"`)

	sidebar := NewSidebarLayout(
		NewNav([]NavLink{{Label: "Docs", Href: "/docs"}}),
		NewArticle("Main", []Component{NewParagraph("content")}),
	)
	assertContains(t, sidebar.HTML(), `class="layout-sidebar-main"`)

	container := NewContainer([]Component{NewParagraph("inside")})
	assertContains(t, container.HTML(), `class="container"`)

	fluid := NewContainerFluid([]Component{NewParagraph("wide")})
	assertContains(t, fluid.HTML(), `class="container-fluid"`)

	assertContains(t, NewThemeToggle().HTML(), "theme-toggle")
}

func TestAppShellWithSidebar(t *testing.T) {
	shell := NewAppShell(NewNav([]NavLink{{Label: "Home", Href: "/"}}), NewCard("WS", "content"))
	html := shell.HTML()
	assertContains(t, html, `class="app-shell"`)
	assertContains(t, html, `class="app-shell-sidebar"`)
	assertContains(t, html, `class="app-shell-content"`)
}

func TestAppShellWithoutSidebar(t *testing.T) {
	shell := NewWorkspace(nil, NewPanel("Main", []Component{NewText("content")}))
	html := shell.HTML()
	assertContains(t, html, `class="app-shell app-shell-single"`)
	assertContains(t, html, `class="app-shell-content"`)
}

func TestSectionAndArticle(t *testing.T) {
	section := NewSection("Title", []Component{
		NewParagraph("one"),
		NewArticle("Nested", []Component{NewParagraph("two")}),
	})
	html := section.HTML()
	assertContains(t, html, "<section>")
	assertContains(t, html, "<h2>Title</h2>")
	assertContains(t, html, "<p>one</p>")
	assertContains(t, html, "<h3>Nested</h3>")
}

func TestSectionAdd(t *testing.T) {
	s := NewSection("S", []Component{NewParagraph("a")})
	s.Add([]Component{NewParagraph("b")})
	html := s.HTML()
	assertContains(t, html, "<p>a</p>")
	assertContains(t, html, "<p>b</p>")
}

func TestListComponents(t *testing.T) {
	assertContains(t, NewList([]string{"a", "b"}).HTML(), "<ul>")
	assertContains(t, NewList([]string{"a"}).HTML(), "<li>a</li>")
	assertContains(t, NewOrderedList([]string{"x"}).HTML(), "<ol>")
	dl := NewDescriptionList([]DescriptionItem{{Term: "K", Definition: "V"}})
	assertContains(t, dl.HTML(), "<dt>K</dt>")
	assertContains(t, dl.HTML(), "<dd>V</dd>")
}

func TestNavigationComponents(t *testing.T) {
	nav := NewNav([]NavLink{{Label: "Home", Href: "/"}})
	assertContains(t, nav.HTML(), "<nav>")
	assertContains(t, nav.HTML(), `href="/"`)

	bc := NewBreadcrumbs([]BreadcrumbItem{
		{Label: "Home", Href: "/"},
		{Label: "Demo", Current: true},
	})
	assertContains(t, bc.HTML(), `aria-label="Breadcrumb"`)
	assertContains(t, bc.HTML(), `aria-current="page"`)

	pg := NewPagination([]PageItem{
		{Label: "Prev", Disabled: true},
		{Label: "1", Current: true},
		{Label: "2", Href: "/?page=2"},
	})
	assertContains(t, pg.HTML(), `aria-label="Pagination"`)
	assertContains(t, pg.HTML(), `aria-disabled="true"`)
	assertContains(t, pg.HTML(), `aria-current="page"`)
}

func TestTableComponents(t *testing.T) {
	assertContains(t, NewTable([]string{"A"}, [][]string{{"B"}}).HTML(), "<table>")
	assertContains(t, NewTableWithClass("ms-table", []string{"A"}, [][]string{{"B"}}).HTML(), `class="ms-table"`)
}

func TestDetailsComponent(t *testing.T) {
	d := NewDetails("Summary", []Component{NewParagraph("body")})
	assertContains(t, d.HTML(), "<details>")
	assertContains(t, d.HTML(), "<summary>Summary</summary>")
	assertContains(t, d.HTML(), "<p>body</p>")
}

func TestFragmentComponent(t *testing.T) {
	f := NewFragment([]Component{NewParagraph("a"), NewParagraph("b")})
	f.Add([]Component{NewParagraph("c")})
	html := f.HTML()
	assertContains(t, html, "<p>a</p>")
	assertContains(t, html, "<p>b</p>")
	assertContains(t, html, "<p>c</p>")
}

func TestGUIAliases(t *testing.T) {
	assertContains(t, NewView("D", []Component{NewText("x")}).HTML(), "<section>")
	assertContains(t, NewPanel("P", []Component{NewText("x")}).HTML(), "<article")
	assertContains(t, NewAction("Run", "primary").HTML(), "btn-primary")
	assertContains(t, NewMenu([]NavLink{{Label: "X", Href: "/"}}).HTML(), "<nav>")
	assertContains(t, NewStatus("ok", "info").HTML(), "alert-info")
	assertContains(t, NewDataGridWithClass("table-x", []string{"A"}, [][]string{{"B"}}).HTML(), `class="table-x"`)
	assertContains(t, NewToolbar([]Component{NewButton("B", "default")}).HTML(), `class="cluster toolbar"`)
}

func TestAddMethods(t *testing.T) {
	a := NewArticle("T", []Component{NewParagraph("a")})
	a.Add([]Component{NewParagraph("b")})
	assertContains(t, a.HTML(), "<p>a</p>")
	assertContains(t, a.HTML(), "<p>b</p>")

	cl := NewCluster("", []Component{NewParagraph("a")})
	cl.Add([]Component{NewParagraph("b")})
	assertContains(t, cl.HTML(), "<p>b</p>")

	co := NewContainer([]Component{NewParagraph("a")})
	co.Add([]Component{NewParagraph("b")})
	assertContains(t, co.HTML(), "<p>b</p>")

	cf := NewContainerFluid([]Component{NewParagraph("a")})
	cf.Add([]Component{NewParagraph("b")})
	assertContains(t, cf.HTML(), "<p>b</p>")

	d := NewDetails("S", []Component{NewParagraph("a")})
	d.Add([]Component{NewParagraph("b")})
	assertContains(t, d.HTML(), "<p>b</p>")

	fs := NewFieldset("L", []Component{NewParagraph("a")})
	fs.Add([]Component{NewParagraph("b")})
	assertContains(t, fs.HTML(), "<p>b</p>")

	fm := NewForm("/", "post", []Component{NewParagraph("a")})
	fm.Add([]Component{NewParagraph("b")})
	assertContains(t, fm.HTML(), "<p>b</p>")

	h := NewHero("T", "S", []Component{NewButton("a", "primary")})
	h.Add([]Component{NewButton("b", "default")})
	assertContains(t, h.HTML(), "btn-default")

	st := NewStack("", []Component{NewParagraph("a")})
	st.Add([]Component{NewParagraph("b")})
	assertContains(t, st.HTML(), "<p>b</p>")
}

func TestDataGridAlias(t *testing.T) {
	dg := NewDataGrid([]string{"Col"}, [][]string{{"val"}})
	assertContains(t, dg.HTML(), "<table>")
	assertContains(t, dg.HTML(), "<th>Col</th>")
}

func TestNilChildren(t *testing.T) {
	s := NewSection("T", nil)
	assertContains(t, s.HTML(), "<section>")
	f := NewFragment(nil)
	if f.HTML() != "" {
		t.Fatalf("expected empty fragment for nil children, got %q", f.HTML())
	}
}

func TestInteractiveWrappers(t *testing.T) {
	ix := Interaction{
		Get:     "/provider/stitch",
		Target:  "#provider-status",
		Swap:    "innerHTML",
		Trigger: "click",
		PushURL: "true",
	}
	action := NewInteractiveAction("Switch", "primary", ix)
	aHTML := action.HTML()
	assertContains(t, aHTML, `hx-get="/provider/stitch"`)
	assertContains(t, aHTML, `hx-target="#provider-status"`)
	assertContains(t, aHTML, `hx-swap="innerHTML"`)
	assertContains(t, aHTML, `hx-trigger="click"`)
	assertContains(t, aHTML, `hx-push-url="true"`)

	menu := NewInteractiveMenu([]InteractiveMenuLink{{Label: "R", Href: "#", Interaction: ix}})
	mHTML := menu.HTML()
	assertContains(t, mHTML, `hx-get="/provider/stitch"`)
	assertContains(t, mHTML, `hx-target="#provider-status"`)
}

func TestWithIDWrapsTemplateComponent(t *testing.T) {
	h := WithID("hero-main", NewHero("Title", "Sub", nil)).HTML()
	assertContains(t, h, `<section`)
	assertContains(t, h, `id="hero-main"`)
	assertContains(t, h, `class="hero"`)
}

func TestWithIDWrapsHTMXComponent(t *testing.T) {
	ix := Interaction{Get: "/provider/stitch", Target: "#provider-status"}
	h := WithID("menu-root", NewInteractiveMenu([]InteractiveMenuLink{{Label: "R", Href: "#", Interaction: ix}})).HTML()
	assertContains(t, h, `<nav id="menu-root">`)
}

func TestWithIDDoesNotDuplicateExistingID(t *testing.T) {
	h := WithID("outer-id", fixedIDRootComponent{}).HTML()
	if strings.Count(h, `id="existing"`) != 1 {
		t.Fatalf("expected existing id to be preserved once, got: %s", h)
	}
	if strings.Contains(h, `id="outer-id"`) {
		t.Fatalf("expected wrapper id to be skipped when root already has id, got: %s", h)
	}
}

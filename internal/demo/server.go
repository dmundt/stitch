package demo

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/dmundt/stitch/css"
	"github.com/dmundt/stitch/htmx"
	"github.com/dmundt/stitch/render"
	"github.com/dmundt/stitch/ui"
)

// BuildDemoPage renders the sample showcase page for providerName.
func BuildDemoPage(providerName string) (string, error) {
	provider, err := css.Get(providerName)
	if err != nil {
		return "", err
	}

	tableClass := ""
	if providerName == "minstyle" {
		tableClass = "ms-table ms-striped"
	}

	layoutSection := ui.NewView("Layout & Structure", []ui.Component{
		ui.NewBreadcrumbs([]ui.BreadcrumbItem{
			{Label: "Home", Href: "/"},
			{Label: "Showcase", Current: true},
		}),
		ui.NewText("Row and column primitives map to a simple 12-column CSS grid in the default stitch theme."),
		ui.NewRow([]ui.Component{
			ui.NewColumn("col-8", []ui.Component{
				ui.NewPanel("Primary", []ui.Component{ui.NewText("Stable provider-agnostic layout region.")}),
			}),
			ui.NewColumn("col-4", []ui.Component{
				ui.NewPanel("Secondary", []ui.Component{ui.NewText("Another semantic region for side-by-side concepts.")}),
			}),
		}),
		ui.NewRow([]ui.Component{
			ui.NewColumn("col-12", []ui.Component{
				ui.NewArticle("Workspace", []ui.Component{ui.NewParagraph("Layout structures stay semantic while row/column define grid placement.")}),
			}),
		}),
		ui.NewText("Grid and grid-item constructors offer compact equal-column and span-based layouts."),
		ui.NewGrid("grid-3", []ui.Component{
			ui.NewGridItem("span-2", []ui.Component{ui.NewPanel("Grid Area A", []ui.Component{ui.NewText("Spans two columns with semantic content.")})}),
			ui.NewGridItem("", []ui.Component{ui.NewPanel("Grid Area B", []ui.Component{ui.NewText("Standard one-column grid item.")})}),
		}),
	})

	layoutSystemsSection := ui.NewView("Layout Systems", []ui.Component{
		ui.NewStack("", []ui.Component{
			ui.NewText("Stack and cluster primitives help build predictable spacing and action groups."),
			ui.NewCluster("", []ui.Component{
				ui.NewBadge("grid-first", "success"),
				ui.NewBadge("semantic", "info"),
				ui.NewBadge("composable", "default"),
			}),
		}),
		ui.NewSplit(
			ui.NewArticle("Split Primary", []ui.Component{ui.NewParagraph("Use split panes for editorial content and supplemental context.")}),
			ui.NewCard("Split Secondary", "Compact side-panel content."),
		),
	})

	contentSection := ui.NewView("Content Elements", []ui.Component{
		ui.NewStatus("Server-side HTML only", "info"),
		ui.NewBadge("beta", "success"),
		ui.NewBlockquote("Composable HTML feels better than giant templates.", "Stitch"),
		ui.NewDetails("Why Stitch?", []ui.Component{ui.NewText("Predefined blocks plus composable elements keep SSR maintainable.")}),
		ui.NewHorizontalRule(),
		ui.NewList([]string{"Fast", "Composable", "SSR"}),
		ui.NewOrderedList([]string{"Compose", "Render", "Ship"}),
		ui.NewDescriptionList([]ui.DescriptionItem{
			{Term: "SSR", Definition: "Server-side rendering"},
			{Term: "UI", Definition: "Reusable HTML builders"},
		}),
		ui.NewCodeBlock("go run ./cmd/demo"),
		ui.NewImage("https://placehold.co/320x120", "placeholder"),
	})

	formSection := ui.NewView("Form Elements", []ui.Component{
		ui.NewForm("/submit", "post", []ui.Component{
			ui.NewFieldset("User Profile", []ui.Component{
				ui.NewInput("Your name", "name", "Jane Doe"),
				ui.NewSelect("Role", "role", []ui.SelectOption{
					{Value: "dev", Label: "Developer", Selected: true},
					{Value: "pm", Label: "Product Manager"},
				}),
				ui.NewTextArea("About", "about", "Tell us about your project"),
				ui.NewCheckbox("newsletter", "yes", "Subscribe to updates", true),
				ui.NewRadio("plan", "free", "Free", true),
				ui.NewRadio("plan", "pro", "Pro", false),
				ui.NewButton("Submit", "primary"),
			}),
		}),
	})

	tableSection := ui.NewView("Tables & Navigation", []ui.Component{
		ui.NewDataGridWithClass(
			tableClass,
			[]string{"Component", "Status"},
			[][]string{
				{"Buttons", "Implemented"},
				{"Forms", "Implemented"},
				{"Tables", "Implemented"},
				{"Nested Layout", "Implemented"},
			},
		),
		ui.NewNav([]ui.NavLink{{Label: "Home", Href: "/"}, {Label: "Milligram", Href: "/provider/milligram"}}),
	})

	navigationControlsSection := ui.NewView("Navigation Controls", []ui.Component{
		ui.NewText("Common CSS frameworks typically support breadcrumbs and pagination. Stitch provides both as semantic HTML builders."),
		ui.NewInteractiveAction("Reload Current Provider", "default", ui.Interaction{
			Get:     "/provider/" + providerName,
			Target:  "main",
			Swap:    "outerHTML",
			PushURL: "/provider/" + providerName,
		}),
		ui.NewInteractiveMenu([]ui.InteractiveMenuLink{
			{Label: "Stitch", Href: "/provider/stitch", Interaction: ui.Interaction{Get: "/provider/stitch", Target: "main", Swap: "outerHTML", PushURL: "/provider/stitch"}},
			{Label: "None", Href: "/provider/none", Interaction: ui.Interaction{Get: "/provider/none", Target: "main", Swap: "outerHTML", PushURL: "/provider/none"}},
			{Label: "Milligram", Href: "/provider/milligram", Interaction: ui.Interaction{Get: "/provider/milligram", Target: "main", Swap: "outerHTML", PushURL: "/provider/milligram"}},
		}),
		ui.NewBreadcrumbs([]ui.BreadcrumbItem{
			{Label: "Dashboard", Href: "#"},
			{Label: "Catalog", Href: "#"},
			{Label: "Components", Current: true},
		}),
		ui.NewPagination([]ui.PageItem{
			{Label: "Previous", Disabled: true},
			{Label: "1", Href: "#"},
			{Label: "2", Current: true},
			{Label: "3", Href: "#"},
			{Label: "Next", Href: "#"},
		}),
		ui.NewPagination([]ui.PageItem{
			{Label: "Previous", Href: "#"},
			{Label: "7", Href: "#"},
			{Label: "8", Current: true},
			{Label: "9", Href: "#"},
			{Label: "Next", Href: "#"},
		}),
		ui.NewPagination([]ui.PageItem{
			{Label: "Previous", Disabled: true},
			{Label: "1", Current: true},
			{Label: "2", Href: "#"},
			{Label: "Next", Href: "#"},
		}),
	})

	workspace := ui.NewStack("workspace-stack", []ui.Component{
		ui.NewHero("Stitch Workspace", "A compact, grid-first Go GUI-style framework with complete object composition.", []ui.Component{
			ui.NewAction("Explore Components", "primary"),
			ui.NewAction("View Source", "default"),
		}),
		ui.NewContainer([]ui.Component{ui.NewText("This page demonstrates complete object-style composition with provider-agnostic structure.")}),
		layoutSection,
		layoutSystemsSection,
		contentSection,
		formSection,
		tableSection,
		navigationControlsSection,
	})

	showcase := ui.NewWorkspace(nil, workspace)

	page := render.NewWindow("Stitch Demo")
	page.WithHeadRaw(`<meta name="description" content="Stitch demo app">`)
	page.WithHead(htmx.Head())
	page.WithHead(layoutNormalizationStyles())
	page.TopBarComponent(ui.NewNav([]ui.NavLink{
		{Label: "Stitch (Default)", Href: "/provider/stitch"},
		{Label: "minstyle.io", Href: "/provider/minstyle"},
		{Label: "Milligram", Href: "/provider/milligram"},
		{Label: "None", Href: "/provider/none"},
	}))
	page.TopBarComponent(ui.NewThemeToggle())
	page.TopBarComponent(ui.NewHeading(1, "Stitch Component Showcase"))
	page.TopBarComponent(ui.NewParagraph("Current provider: " + providerName))
	page.TopBarComponent(ui.NewParagraph("Provider badge: " + strings.ToUpper(providerName)))
	page.ContentComponent(showcase)
	page.StatusBarComponent(ui.NewParagraph(fmt.Sprintf("Stitch demo footer (%s)", providerName)))

	return page.Render(provider)
}

// NewHandler returns the demo HTTP handler.
func NewHandler() http.Handler {
	mux := http.NewServeMux()
	assets, err := css.Assets()
	if err != nil {
		panic("failed to open core css assets: " + err.Error())
	}
	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.FS(assets))))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		routeProvider(w, "stitch")
	})
	mux.HandleFunc("/provider/", func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/provider/")
		routeProvider(w, name)
	})
	return mux
}

func layoutNormalizationStyles() template.HTML {
	return template.HTML(`<style>
/* ── canvas ──────────────────────────────────────────── */
body {
	margin: 0;
	padding: 0.75rem 1rem;
	min-height: 100vh;
	box-sizing: border-box;
}

/* ── document card ───────────────────────────────────── */
.stitch-page {
	max-width: 960px;
	margin: 0 auto;
	padding: 1rem 1.1rem;
}

/* ── header / footer chrome ──────────────────────────── */
header {
	padding: 0.65rem 0 0.45rem;
	position: relative;
}

footer {
	padding: 0.45rem 0;
	color: #4b5563;
}

/* ── main content area ───────────────────────────────── */
main {
	padding: 0.55rem 0;
}

/* ── sections ────────────────────────────────────────── */
section {
	margin-top: 0.4rem;
	margin-bottom: 0.4rem;
	padding: 0.25rem 0;
}

section > h2 {
	margin-top: 0;
}

/* ── theme toggle ────────────────────────────────────── */
.theme-toggle {
	position: absolute;
	top: 1rem;
	right: 1.15rem;
	cursor: pointer;
	user-select: none;
	font-size: 0.95rem;
	opacity: 0.7;
}

.theme-toggle:hover {
	opacity: 1;
}

.theme-toggle input {
	position: absolute;
	opacity: 0;
	width: 0;
	height: 0;
}

/* ── nav ─────────────────────────────────────────────── */
nav ul {
	list-style: none;
	display: flex;
	flex-wrap: wrap;
	gap: 0.55rem;
	margin: 0;
	padding: 0;
}

nav a {
	display: inline-block;
	padding: 0.2rem 0.35rem;
}

/* ── containers ──────────────────────────────────────── */
.container {
	max-width: 960px;
	margin: 0 auto;
}

.container-fluid {
	width: 100%;
}

/* ── app shell ───────────────────────────────────────── */
.app-shell,
.app-shell.app-shell-single {
	display: block;
}

.app-shell-content {
	min-width: 0;
}

.app-shell-sidebar {
	display: none;
}

/* ── row / column grid ───────────────────────────────── */
.row {
	display: grid;
	grid-template-columns: repeat(12, minmax(0, 1fr));
	gap: 0.8rem;
	align-items: start;
}

.row > .column {
	min-width: 0;
	grid-column: span 12;
}

.column.col-1  { grid-column: span 1; }
.column.col-2  { grid-column: span 2; }
.column.col-3  { grid-column: span 3; }
.column.col-4  { grid-column: span 4; }
.column.col-5  { grid-column: span 5; }
.column.col-6  { grid-column: span 6; }
.column.col-7  { grid-column: span 7; }
.column.col-8  { grid-column: span 8; }
.column.col-9  { grid-column: span 9; }
.column.col-10 { grid-column: span 10; }
.column.col-11 { grid-column: span 11; }
.column.col-12 { grid-column: span 12; }

/* ── stack / cluster / toolbar ───────────────────────── */
.stack {
	display: grid;
	gap: 0.7rem;
}

.cluster {
	display: flex;
	flex-wrap: wrap;
	align-items: center;
	gap: 0.5rem;
}

.toolbar {
	justify-content: flex-start;
}

/* ── hero ────────────────────────────────────────────── */
.hero {
	padding: 1rem;
}

.hero .hero-actions {
	margin-top: 0.45rem;
}

/* ── split ───────────────────────────────────────────── */
.split {
	display: grid;
	grid-template-columns: 2fr 1fr;
	gap: 0.8rem;
}

/* ── sidebar layout ──────────────────────────────────── */
.layout-sidebar-main {
	display: grid;
	grid-template-columns: 220px minmax(0, 1fr);
	gap: 0.8rem;
}

.layout-sidebar,
.layout-main {
	padding: 0.65rem;
}

/* ── grid ────────────────────────────────────────────── */
.grid {
	display: grid;
	grid-template-columns: repeat(1, minmax(0, 1fr));
	gap: 0.8rem;
	align-items: start;
}

.grid.grid-2 { grid-template-columns: repeat(2, minmax(0, 1fr)); }
.grid.grid-3 { grid-template-columns: repeat(3, minmax(0, 1fr)); }
.grid.grid-4 { grid-template-columns: repeat(4, minmax(0, 1fr)); }

.grid > .grid-item { min-width: 0; }
.grid > .grid-item.span-2 { grid-column: span 2; }
.grid > .grid-item.span-3 { grid-column: span 3; }
.grid > .grid-item.span-4 { grid-column: span 4; }

/* ── badge ───────────────────────────────────────────── */
.badge {
	display: inline-block;
	padding: 0.15rem 0.35rem;
	font-size: 0.82rem;
}

/* ── alert ───────────────────────────────────────────── */
.alert {
	padding: 0.55rem 0.65rem 0.55rem 0.85rem;
}

/* ── table ───────────────────────────────────────────── */
table {
	margin-top: 0.3rem;
	margin-bottom: 0.3rem;
}

/* ── images ──────────────────────────────────────────── */
img {
	max-width: 100%;
	height: auto;
}

/* ── dark mode (pure CSS, no JS) ─────────────────────── */
html:has(#theme-toggle:checked) body {
	background: #1a1b1e;
	color: #adbac7;
}

html:has(#theme-toggle:checked) .stitch-page {
	background: #1a1b1e;
}

html:has(#theme-toggle:checked) .hero,
html:has(#theme-toggle:checked) .card,
html:has(#theme-toggle:checked) article,
html:has(#theme-toggle:checked) .layout-sidebar,
html:has(#theme-toggle:checked) .layout-main {
	background: #1b212a;
}

html:has(#theme-toggle:checked) header,
html:has(#theme-toggle:checked) footer {
	color: #adbac7;
}

html:has(#theme-toggle:checked) th,
html:has(#theme-toggle:checked) td {
	color: #cdd9e5;
}

/* ── responsive ──────────────────────────────────────── */
@media (max-width: 820px) {
	.row > .column {
		grid-column: span 12;
	}

	.split,
	.layout-sidebar-main,
	.grid.grid-2,
	.grid.grid-3,
	.grid.grid-4 {
		grid-template-columns: 1fr;
	}

	.grid > .grid-item.span-2,
	.grid > .grid-item.span-3,
	.grid > .grid-item.span-4 {
		grid-column: span 1;
	}
}

@media (max-width: 700px) {
	body {
		padding: 0;
	}

	.stitch-page {
		margin: 0;
	}

	header, main, footer {
		padding-left: 1rem;
		padding-right: 1rem;
	}

	section {
		padding: 0.2rem 0;
	}
}
</style>`)
}

func routeProvider(w http.ResponseWriter, providerName string) {
	html, err := BuildDemoPage(providerName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(html))
}

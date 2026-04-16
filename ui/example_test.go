package ui

import "fmt"

// ExampleNewParagraph shows the primary prose text component.
// Use it for body copy in composed layouts.
func ExampleNewParagraph() {
fmt.Println(NewParagraph("Hello, Stitch.").HTML())
// Output: <p>Hello, Stitch.</p>
}

// ExampleNewHeading shows a semantic heading element.
// Level is clamped to 1..6 at render time.
func ExampleNewHeading() {
fmt.Println(NewHeading(1, "Page Title").HTML())
// Output: <h1>Page Title</h1>
}

// ExampleNewBlockquote shows attributed quoted content.
// Cite is optional; omit it for unattributed quotes.
func ExampleNewBlockquote() {
fmt.Println(NewBlockquote("Design for simplicity.", "Go Proverbs").HTML())
// Output: <blockquote><p>Design for simplicity.</p><cite>Go Proverbs</cite></blockquote>
}

// ExampleNewCodeBlock shows preformatted code content.
// Content is HTML-escaped, so angle brackets and special characters are safe.
func ExampleNewCodeBlock() {
fmt.Println(NewCodeBlock("go test ./...").HTML())
// Output: <pre><code>go test ./...</code></pre>
}

// ExampleNewImage shows an image element with required alternative text.
// Alt is required for accessibility; always provide a meaningful description.
func ExampleNewImage() {
fmt.Println(NewImage("/logo.png", "Company logo").HTML())
// Output: <img src="/logo.png" alt="Company logo">
}

// ExampleNewHorizontalRule shows a thematic content boundary.
func ExampleNewHorizontalRule() {
fmt.Println(NewHorizontalRule().HTML())
// Output: <hr>
}

// ExampleNewAlert shows an inline status message.
// Tone selects the visual variant; valid values depend on the active CSS provider.
func ExampleNewAlert() {
fmt.Println(NewAlert("File saved successfully.", "success").HTML())
// Output: <aside class="alert alert-success" role="status">File saved successfully.</aside>
}

// ExampleNewBadge shows a compact metadata label.
// Tone sets the visual variant; valid values depend on the active CSS provider.
func ExampleNewBadge() {
fmt.Println(NewBadge("New", "success").HTML())
// Output: <span class="badge badge-success">New</span>
}

// ExampleNewButton shows a styled action button.
// Kind sets the visual variant; valid values depend on the active CSS provider.
func ExampleNewButton() {
fmt.Println(NewButton("Save", "primary").HTML())
// Output: <button class="btn btn-primary">Save</button>
}

// ExampleNewCard shows a compact titled content container.
// Use it for dashboard tiles and summary panels.
func ExampleNewCard() {
fmt.Println(NewCard("Summary", "A compact content block.").HTML())
// Output: <article class="card"><header><h3>Summary</h3></header><p>A compact content block.</p></article>
}

// ExampleNewInput shows a labeled text input field.
func ExampleNewInput() {
fmt.Println(NewInput("Name", "name", "Jane Doe").HTML())
// Output: <label>Name<input name="name" placeholder="Jane Doe"></label>
}

// ExampleNewCheckbox shows a checkbox form control.
// Checked controls the initial state in rendered markup.
func ExampleNewCheckbox() {
fmt.Println(NewCheckbox("agree", "yes", "I agree", true).HTML())
// Output: <label><input type="checkbox" name="agree" value="yes" checked>I agree</label>
}

// ExampleNewRadio shows a radio-button form control.
// All RadioComponents with the same Name share a selection group.
func ExampleNewRadio() {
fmt.Println(NewRadio("plan", "free", "Free plan", true).HTML())
// Output: <label><input type="radio" name="plan" value="free" checked>Free plan</label>
}

// ExampleNewTextArea shows a labeled multiline text input.
func ExampleNewTextArea() {
fmt.Println(NewTextArea("Bio", "bio", "About yourself").HTML())
// Output: <label>Bio<textarea name="bio" placeholder="About yourself"></textarea></label>
}

// ExampleNewSelect shows a labeled select input with options.
// Set SelectOption.Selected to mark the initial selection.
func ExampleNewSelect() {
s := NewSelect("Role", "role", []SelectOption{
{Value: "dev", Label: "Developer", Selected: true},
{Value: "pm", Label: "Manager"},
})
fmt.Println(s.HTML())
// Output: <label>Role<select name="role"><option value="dev" selected>Developer</option><option value="pm">Manager</option></select></label>
}

// ExampleNewFieldset shows grouped form controls under a shared legend.
// Wrap related inputs in a fieldset to add semantic grouping.
func ExampleNewFieldset() {
fs := NewFieldset("Profile", []Component{
NewInput("Name", "name", ""),
})
fmt.Println(fs.HTML())
// Output: <fieldset><legend>Profile</legend><label>Name<input name="name" placeholder=""></label></fieldset>
}

// ExampleNewForm shows an HTML form with controls.
// Action and Method are set directly on the form element.
func ExampleNewForm() {
form := NewForm("/save", "post", []Component{
NewInput("Name", "name", ""),
})
fmt.Println(form.HTML())
// Output: <form action="/save" method="post"><label>Name<input name="name" placeholder=""></label></form>
}

// ExampleNewSection shows a titled section of related content.
// Use it to group multiple content components under a shared heading.
func ExampleNewSection() {
s := NewSection("Introduction", []Component{
NewParagraph("Welcome to Stitch."),
})
fmt.Println(s.HTML())
// Output: <section><h2>Introduction</h2><p>Welcome to Stitch.</p></section>
}

// ExampleNewArticle shows a titled self-contained content region.
// Use it for cards, panels, or distinct content groups within a section.
func ExampleNewArticle() {
a := NewArticle("Background", []Component{
NewParagraph("Context for this page."),
})
fmt.Println(a.HTML())
// Output: <article><header><h3>Background</h3></header><p>Context for this page.</p></article>
}

// ExampleNewDetails shows expandable disclosure content.
// Summary is always visible; children are shown when the user expands.
func ExampleNewDetails() {
d := NewDetails("Advanced settings", []Component{
NewParagraph("Use with care."),
})
fmt.Println(d.HTML())
// Output: <details><summary>Advanced settings</summary><p>Use with care.</p></details>
}

// ExampleNewFragment shows children grouped without a wrapper element.
// Use it when adding a wrapper element would break semantic structure.
func ExampleNewFragment() {
f := NewFragment([]Component{
NewParagraph("First."),
NewParagraph("Second."),
})
fmt.Println(f.HTML())
// Output: <p>First.</p><p>Second.</p>
}

// ExampleNewContainer shows a centered constrained-width wrapper.
// Use it as a page content boundary to limit line length and center content.
func ExampleNewContainer() {
c := NewContainer([]Component{NewParagraph("Centered content.")})
fmt.Println(c.HTML())
// Output: <div class="container"><p>Centered content.</p></div>
}

// ExampleNewContainerFluid shows a full-width layout wrapper.
// Use it for edge-to-edge sections that should not be width-constrained.
func ExampleNewContainerFluid() {
c := NewContainerFluid([]Component{NewParagraph("Wide content.")})
fmt.Println(c.HTML())
// Output: <div class="container-fluid"><p>Wide content.</p></div>
}

// ExampleNewStack shows children arranged in vertical order.
// ExtraClass can add spacing or alignment modifiers from the active CSS provider.
func ExampleNewStack() {
s := NewStack("", []Component{
NewParagraph("Item one."),
NewParagraph("Item two."),
})
fmt.Println(s.HTML())
// Output: <div class="stack"><p>Item one.</p><p>Item two.</p></div>
}

// ExampleNewCluster shows children arranged in a wrapping horizontal group.
// Use it for action groups, tag lists, or badge rows.
func ExampleNewCluster() {
c := NewCluster("", []Component{
NewButton("One", "default"),
NewButton("Two", "default"),
})
fmt.Println(c.HTML())
// Output: <div class="cluster"><button class="btn btn-default">One</button><button class="btn btn-default">Two</button></div>
}

// ExampleNewHero shows a prominent introductory section.
// Use it for page headers and call-to-action areas.
func ExampleNewHero() {
h := NewHero("Welcome", "Build UIs with Go.", []Component{
NewButton("Get started", "primary"),
})
fmt.Println(h.HTML())
// Output: <section class="hero"><h1>Welcome</h1><p>Build UIs with Go.</p><div class="cluster hero-actions"><button class="btn btn-primary">Get started</button></div></section>
}

// ExampleNewSplit shows a two-pane layout with primary and secondary regions.
// Use it for detail panes and editorial layouts with supplemental content.
func ExampleNewSplit() {
s := NewSplit(NewParagraph("Primary."), NewParagraph("Secondary."))
fmt.Println(s.HTML())
// Output: <section class="split"><div class="split-primary"><p>Primary.</p></div><div class="split-secondary"><p>Secondary.</p></div></section>
}

// ExampleNewSidebarLayout shows a sidebar and main content layout.
// Use it for documentation pages and dashboard-style structures.
func ExampleNewSidebarLayout() {
s := NewSidebarLayout(NewParagraph("Nav."), NewParagraph("Main content."))
fmt.Println(s.HTML())
// Output: <div class="layout-sidebar-main"><aside class="layout-sidebar"><p>Nav.</p></aside><main class="layout-main"><p>Main content.</p></main></div>
}

// ExampleNewAppShell_withSidebar shows a page shell with sidebar and content.
// Use NewAppShell when both regions are present for a full workspace layout.
func ExampleNewAppShell_withSidebar() {
shell := NewAppShell(NewParagraph("Navigation."), NewParagraph("Content."))
fmt.Println(shell.HTML())
// Output: <section class="app-shell"><aside class="app-shell-sidebar"><p>Navigation.</p></aside><section class="app-shell-content"><p>Content.</p></section></section>
}

// ExampleNewAppShell_withoutSidebar shows a page shell with content only.
// Pass nil as sidebar to render a full-width single-pane layout.
func ExampleNewAppShell_withoutSidebar() {
shell := NewAppShell(nil, NewParagraph("Main content."))
fmt.Println(shell.HTML())
// Output: <section class="app-shell app-shell-single"><section class="app-shell-content"><p>Main content.</p></section></section>
}

// ExampleNewRow shows a 12-column grid row.
// Pair with NewColumn and a provider span class to control column widths.
func ExampleNewRow() {
r := NewRow([]Component{
NewColumn("col-8", []Component{NewParagraph("Main.")}),
NewColumn("col-4", []Component{NewParagraph("Side.")}),
})
fmt.Println(r.HTML())
// Output: <div class="row"><div class="column col-8"><p>Main.</p></div><div class="column col-4"><p>Side.</p></div></div>
}

// ExampleNewGrid shows a generic grid layout.
// ColumnsClass selects the column template; GridItem.SpanClass controls spanning.
func ExampleNewGrid() {
g := NewGrid("grid-3", []Component{
NewGridItem("span-2", []Component{NewParagraph("Wide.")}),
NewGridItem("", []Component{NewParagraph("Narrow.")}),
})
fmt.Println(g.HTML())
// Output: <div class="grid grid-3"><div class="grid-item span-2"><p>Wide.</p></div><div class="grid-item"><p>Narrow.</p></div></div>
}

// ExampleNewNav shows a semantic navigation region.
// Use it for site menus, footers, and secondary navigation.
func ExampleNewNav() {
nav := NewNav([]NavLink{
{Label: "Home", Href: "/"},
{Label: "Docs", Href: "/docs"},
})
fmt.Println(nav.HTML())
// Output: <nav><ul><li><a href="/">Home</a></li><li><a href="/docs">Docs</a></li></ul></nav>
}

// ExampleNewBreadcrumbs shows hierarchical location navigation.
// Set Current: true on the last item to mark the active location.
func ExampleNewBreadcrumbs() {
bc := NewBreadcrumbs([]BreadcrumbItem{
{Label: "Home", Href: "/"},
{Label: "Docs", Current: true},
})
fmt.Println(bc.HTML())
// Output: <nav aria-label="Breadcrumb"><ol class="breadcrumbs"><li><a href="/">Home</a></li><li><span aria-current="page">Docs</span></li></ol></nav>
}

// ExampleNewPagination shows page-to-page navigation controls.
// Use Current for the active page, Disabled for inactive controls.
func ExampleNewPagination() {
pg := NewPagination([]PageItem{
{Label: "Prev", Disabled: true},
{Label: "1", Current: true},
{Label: "2", Href: "/?page=2"},
})
fmt.Println(pg.HTML())
// Output: <nav aria-label="Pagination"><ul class="pagination"><li><span aria-disabled="true">Prev</span></li><li><span aria-current="page">1</span></li><li><a href="/?page=2">2</a></li></ul></nav>
}

// ExampleNewList shows an unordered bullet list.
func ExampleNewList() {
fmt.Println(NewList([]string{"Alpha", "Beta", "Gamma"}).HTML())
// Output: <ul><li>Alpha</li><li>Beta</li><li>Gamma</li></ul>
}

// ExampleNewOrderedList shows a numbered list.
// Use it for steps, instructions, and ranked content.
func ExampleNewOrderedList() {
fmt.Println(NewOrderedList([]string{"Compose", "Render", "Ship"}).HTML())
// Output: <ol><li>Compose</li><li>Render</li><li>Ship</li></ol>
}

// ExampleNewDescriptionList shows semantic term-definition content.
// Use it for glossaries, metadata tables, and key-value summaries.
func ExampleNewDescriptionList() {
dl := NewDescriptionList([]DescriptionItem{
{Term: "SSR", Definition: "Server-side rendering"},
})
fmt.Println(dl.HTML())
// Output: <dl><dt>SSR</dt><dd>Server-side rendering</dd></dl>
}

// ExampleNewTable shows a data table with headers and rows.
func ExampleNewTable() {
t := NewTable(
[]string{"Component", "Status"},
[][]string{{"Button", "Done"}, {"Form", "Done"}},
)
fmt.Println(t.HTML())
// Output: <table><thead><tr><th>Component</th><th>Status</th></tr></thead><tbody><tr><td>Button</td><td>Done</td></tr><tr><td>Form</td><td>Done</td></tr></tbody></table>
}

// ExampleNewTableWithClass shows a table with a CSS provider class applied.
// Use it when your CSS provider requires a class for table styling.
func ExampleNewTableWithClass() {
t := NewTableWithClass("ms-table", []string{"Name"}, [][]string{{"Alpha"}})
fmt.Println(t.HTML())
// Output: <table class="ms-table"><thead><tr><th>Name</th></tr></thead><tbody><tr><td>Alpha</td></tr></tbody></table>
}

// ExampleNewThemeToggle shows the theme toggle control.
// Include it in the page header to give users a theme switch.
func ExampleNewThemeToggle() {
fmt.Println(NewThemeToggle().HTML())
// Output: <label class="theme-toggle" for="theme-toggle" title="Toggle dark / light mode"><input type="checkbox" id="theme-toggle"><span aria-hidden="true">☀︎ ╱ ☾</span></label>
}

// ExampleNewInteractiveAction shows a button that triggers an HTMX request.
// Use one request method (Get, Post, Put, or Delete) per Interaction.
func ExampleNewInteractiveAction() {
a := NewInteractiveAction("Load items", "primary", Interaction{
Get:    "/items",
Target: "#list",
Swap:   "outerHTML",
})
fmt.Println(a.HTML())
// Output: <button class="btn btn-primary" hx-get="/items" hx-swap="outerHTML" hx-target="#list">Load items</button>
}

// ExampleNewInteractiveMenu shows a nav menu whose links trigger HTMX updates.
// Each link can carry its own Interaction to update independent page regions.
func ExampleNewInteractiveMenu() {
m := NewInteractiveMenu([]InteractiveMenuLink{
{
Label: "Home",
Href:  "/",
Interaction: Interaction{
Get:    "/",
Target: "main",
Swap:   "outerHTML",
},
},
})
fmt.Println(m.HTML())
// Output: <nav><ul><li><a href="/" hx-get="/" hx-swap="outerHTML" hx-target="main">Home</a></li></ul></nav>
}

// ExampleNewView shows the GUI alias for NewSection.
func ExampleNewView() {
fmt.Println(NewView("Dashboard", []Component{NewParagraph("Overview.")}).HTML())
// Output: <section><h2>Dashboard</h2><p>Overview.</p></section>
}

// ExampleNewPanel shows the GUI alias for NewArticle.
func ExampleNewPanel() {
fmt.Println(NewPanel("Widget", []Component{NewParagraph("Content.")}).HTML())
// Output: <article><header><h3>Widget</h3></header><p>Content.</p></article>
}

// ExampleNewText shows the GUI alias for NewParagraph.
func ExampleNewText() {
fmt.Println(NewText("Readable prose.").HTML())
// Output: <p>Readable prose.</p>
}

// ExampleNewAction shows the GUI alias for NewButton.
func ExampleNewAction() {
fmt.Println(NewAction("Confirm", "primary").HTML())
// Output: <button class="btn btn-primary">Confirm</button>
}

// ExampleNewStatus shows the GUI alias for NewAlert.
func ExampleNewStatus() {
fmt.Println(NewStatus("Operation complete.", "success").HTML())
// Output: <aside class="alert alert-success" role="status">Operation complete.</aside>
}

// ExampleNewMenu shows the GUI alias for NewNav.
func ExampleNewMenu() {
fmt.Println(NewMenu([]NavLink{{Label: "Home", Href: "/"}}).HTML())
// Output: <nav><ul><li><a href="/">Home</a></li></ul></nav>
}

// ExampleNewWorkspace shows the GUI alias for NewAppShell.
func ExampleNewWorkspace() {
w := NewWorkspace(nil, NewParagraph("App content."))
fmt.Println(w.HTML())
// Output: <section class="app-shell app-shell-single"><section class="app-shell-content"><p>App content.</p></section></section>
}

// ExampleNewDataGrid shows the GUI alias for NewTable.
func ExampleNewDataGrid() {
dg := NewDataGrid([]string{"Key", "Value"}, [][]string{{"lang", "Go"}})
fmt.Println(dg.HTML())
// Output: <table><thead><tr><th>Key</th><th>Value</th></tr></thead><tbody><tr><td>lang</td><td>Go</td></tr></tbody></table>
}

// ExampleNewToolbar shows the GUI alias for NewCluster with the toolbar class.
// Use it to group action buttons in a horizontal toolbar strip.
func ExampleNewToolbar() {
tb := NewToolbar([]Component{
NewButton("Save", "primary"),
NewButton("Cancel", "default"),
})
fmt.Println(tb.HTML())
// Output: <div class="cluster toolbar"><button class="btn btn-primary">Save</button><button class="btn btn-default">Cancel</button></div>
}

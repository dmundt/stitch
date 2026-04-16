package htmx

import (
	"fmt"
	"html"
	"html/template"
	"strings"
)

// Attrs holds the most-used htmx request attributes.
// Fields with zero values are omitted from the rendered HTML output.
type Attrs struct {
	Boost   bool   // hx-boost
	Delete  string // hx-delete
	Get     string // hx-get
	Post    string // hx-post
	PushURL string // hx-push-url
	Put     string // hx-put
	Select  string // hx-select
	Swap    string // hx-swap    (e.g. "outerHTML", "innerHTML")
	Target  string // hx-target  (CSS selector, e.g. "#content")
	Trigger string // hx-trigger (e.g. "click", "submit", "change")
}

// NavLink is a navigation link optionally enhanced with htmx attributes.
type NavLink struct {
	Href  string
	HX    Attrs
	Label string
}

func (a Attrs) toAttrs() string {
	var b strings.Builder
	write := func(k, v string) {
		if v != "" {
			fmt.Fprintf(&b, ` %s="%s"`, k, html.EscapeString(v))
		}
	}
	if a.Boost {
		b.WriteString(` hx-boost="true"`)
	}
	write("hx-delete", a.Delete)
	write("hx-get", a.Get)
	write("hx-post", a.Post)
	write("hx-push-url", a.PushURL)
	write("hx-put", a.Put)
	write("hx-select", a.Select)
	write("hx-swap", a.Swap)
	write("hx-target", a.Target)
	write("hx-trigger", a.Trigger)
	return b.String()
}

// BoostNav renders a <nav hx-boost="true"> so every link inside is
// intercepted by htmx and navigated via AJAX with history pushState —
// without writing any JavaScript.
func BoostNav(links []NavLink) string {
	var b strings.Builder
	b.WriteString(`<nav hx-boost="true"><ul>`)
	for _, l := range links {
		fmt.Fprintf(&b, `<li><a href="%s">%s</a></li>`,
			html.EscapeString(l.Href), html.EscapeString(l.Label))
	}
	b.WriteString("</ul></nav>")
	return b.String()
}

// Button renders a <button> enhanced with htmx attributes.
func Button(text, kind string, hx Attrs) string {
	if kind == "" {
		kind = "default"
	}
	return fmt.Sprintf(`<button class="btn btn-%s"%s>%s</button>`,
		html.EscapeString(kind), hx.toAttrs(), html.EscapeString(text))
}

// Form renders a <form> enhanced with htmx attributes.
// The action and method attributes are retained for non-htmx fallback.
func Form(action, method string, hx Attrs, children ...string) string {
	return fmt.Sprintf(`<form action="%s" method="%s"%s>%s</form>`,
		html.EscapeString(action), html.EscapeString(method),
		hx.toAttrs(), strings.Join(children, ""))
}

// Head returns the htmx CDN script tag for inclusion in the page <head>.
// Including it enables all htmx attributes (hx-get, hx-post, etc.) without
// any hand-written JavaScript.
func Head() template.HTML {
	return template.HTML(`<script src="https://unpkg.com/htmx.org@2/dist/htmx.min.js"></script>`)
}

// Nav renders a <nav> whose links carry htmx attributes.
func Nav(links []NavLink) string {
	var b strings.Builder
	b.WriteString("<nav><ul>")
	for _, l := range links {
		fmt.Fprintf(&b, `<li><a href="%s"%s>%s</a></li>`,
			html.EscapeString(l.Href), l.HX.toAttrs(), html.EscapeString(l.Label))
	}
	b.WriteString("</ul></nav>")
	return b.String()
}

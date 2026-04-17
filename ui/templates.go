package ui

import (
	"bytes"
	"embed"
	"html"
	"html/template"
	"sort"
	"strings"
	"sync"
)

//go:embed templates/*.gohtml
var templateFS embed.FS

var tmpl *template.Template

var bufPool = sync.Pool{
	New: func() any { return new(bytes.Buffer) },
}

func init() {
	tmpl = template.Must(template.New("ui").Funcs(template.FuncMap{
		"renderAttrs": renderAttrs,
	}).ParseFS(templateFS, "templates/*.gohtml"))
}

func renderAttrs(attrs map[string]string) template.HTMLAttr {
	if len(attrs) == 0 {
		return ""
	}
	keys := make([]string, 0, len(attrs))
	for key := range attrs {
		if strings.TrimSpace(key) == "" {
			continue
		}
		keys = append(keys, key)
	}
	if len(keys) == 0 {
		return ""
	}
	sort.Strings(keys)
	var b strings.Builder
	for _, key := range keys {
		value := strings.TrimSpace(attrs[key])
		if value == "" {
			continue
		}
		b.WriteString(` `)
		b.WriteString(html.EscapeString(key))
		b.WriteString(`="`)
		b.WriteString(html.EscapeString(value))
		b.WriteString(`"`)
	}
	return template.HTMLAttr(b.String())
}

// execute runs the named template with data and returns the resulting HTML string.
// Panics if the template name is unknown — this is a programmer error.
func execute(name string, data any) string {
	buf := bufPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufPool.Put(buf)
	if err := tmpl.ExecuteTemplate(buf, name, data); err != nil {
		panic("ui: template " + name + ": " + err.Error())
	}
	return buf.String()
}

// renderChildrenHTML renders each child component to HTML and returns a slice
// of template.HTML values safe for use inside html/template templates.
func renderChildrenHTML(children []Component) []template.HTML {
	out := make([]template.HTML, len(children))
	for i, c := range children {
		out[i] = template.HTML(c.HTML())
	}
	return out
}

// renderChildren returns the HTML of each child as a plain string slice.
// Used by FragmentComponent which joins without a wrapper element.
func renderChildren(children []Component) []string {
	out := make([]string, len(children))
	for i, c := range children {
		out[i] = c.HTML()
	}
	return out
}

// cloneComponents returns a shallow copy of the children slice.
func cloneComponents(children []Component) []Component {
	if children == nil {
		return nil
	}
	out := make([]Component, len(children))
	copy(out, children)
	return out
}

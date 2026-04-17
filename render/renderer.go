package render

import (
	"bytes"
	_ "embed"
	"errors"
	"html/template"
	"sync"

	"github.com/dmundt/stitch/css"
	stitchtpl "github.com/dmundt/stitch/template"
)

//go:embed document.gohtml
var documentTemplateSource string

func mustLoadDocumentTemplate() *template.Template {
	if documentTemplateSource == "" {
		panic("render: embedded document template is empty")
	}

	return template.Must(template.New("doc").Parse(documentTemplateSource))
}

var (
	docTemplate   = mustLoadDocumentTemplate()
	renderBufPool = sync.Pool{New: func() any { return new(bytes.Buffer) }}
)

// HTMLComponent is the minimal rendering contract for object-style UI builders.
// It allows render.Page to compose with struct-based components directly.
type HTMLComponent interface {
	HTML() string
}

// Page represents a composable HTML document model.
type Page struct {
	Title    string
	Lang     string
	Composer *stitchtpl.Composer
	Head     []template.HTML
}

type documentData struct {
	Lang         string
	Title        string
	ProviderHead template.HTML
	Head         []template.HTML
	Header       []template.HTML
	Main         []template.HTML
	Footer       []template.HTML
}

// NewPage returns a Page with a default language and composer.
func NewPage(title string) *Page {
	return &Page{
		Title:    title,
		Lang:     "en",
		Composer: stitchtpl.NewComposer(),
		Head:     []template.HTML{},
	}
}

// NewWindow is a GUI-style alias for NewPage.
func NewWindow(title string) *Page {
	return NewPage(title)
}

// AddHead appends trusted HTML to the document head.
func (p *Page) AddHead(html template.HTML) {
	p.Head = append(p.Head, html)
}

// AddHeadRaw appends raw HTML to <head> and does not escape content.
// Use for trusted static fragments such as meta tags and inline style blocks.
func (p *Page) AddHeadRaw(raw string) {
	p.Head = append(p.Head, template.HTML(raw))
}

// Content is a GUI-style alias for Main.
func (p *Page) Content(h template.HTML) {
	p.Main(h)
}

// ContentComponent is a GUI-style alias for MainComponent.
func (p *Page) ContentComponent(c HTMLComponent) {
	p.MainComponent(c)
}

// Footer adds a fragment to the <footer> block.
func (p *Page) Footer(h template.HTML) {
	_ = p.Composer.AddHTML(stitchtpl.BlockFooter, h)
}

// FooterComponent adds a component to <footer>.
func (p *Page) FooterComponent(c HTMLComponent) {
	if c == nil {
		return
	}
	p.Footer(template.HTML(c.HTML()))
}

// Header adds a fragment to the <header> block.
func (p *Page) Header(h template.HTML) {
	_ = p.Composer.AddHTML(stitchtpl.BlockHeader, h)
}

// HeaderComponent adds a component to <header>.
func (p *Page) HeaderComponent(c HTMLComponent) {
	if c == nil {
		return
	}
	p.Header(template.HTML(c.HTML()))
}

// Main adds a fragment to the <main> block.
func (p *Page) Main(h template.HTML) {
	_ = p.Composer.AddHTML(stitchtpl.BlockMain, h)
}

// MainComponent adds a component to <main>.
func (p *Page) MainComponent(c HTMLComponent) {
	if c == nil {
		return
	}
	p.Main(template.HTML(c.HTML()))
}

// Render renders the page using the given provider.
func (p *Page) Render(provider css.Provider) (string, error) {
	return Render(p, provider)
}

// StatusBar is a GUI-style alias for Footer.
func (p *Page) StatusBar(h template.HTML) {
	p.Footer(h)
}

// StatusBarComponent is a GUI-style alias for FooterComponent.
func (p *Page) StatusBarComponent(c HTMLComponent) {
	p.FooterComponent(c)
}

// TopBar is a GUI-style alias for Header.
func (p *Page) TopBar(h template.HTML) {
	p.Header(h)
}

// TopBarComponent is a GUI-style alias for HeaderComponent.
func (p *Page) TopBarComponent(c HTMLComponent) {
	p.HeaderComponent(c)
}

// WithHead adds a <head> fragment.
func (p *Page) WithHead(h template.HTML) {
	p.Head = append(p.Head, h)
}

// WithHeadRaw adds a trusted raw HTML fragment to <head>.
func (p *Page) WithHeadRaw(raw string) {
	p.Head = append(p.Head, template.HTML(raw))
}

// Render renders page to a full HTML document string.
func Render(page *Page, provider css.Provider) (string, error) {
	if page == nil {
		return "", errors.New("page is nil")
	}
	if page.Composer == nil {
		return "", errors.New("composer is nil")
	}
	if provider == nil {
		provider = css.None()
	}

	header, _ := page.Composer.Fragments(stitchtpl.BlockHeader)
	main, _ := page.Composer.Fragments(stitchtpl.BlockMain)
	footer, _ := page.Composer.Fragments(stitchtpl.BlockFooter)

	data := documentData{
		Lang:         page.Lang,
		Title:        page.Title,
		ProviderHead: provider.HeadHTML(),
		Head:         page.Head,
		Header:       header,
		Main:         main,
		Footer:       footer,
	}

	b := renderBufPool.Get().(*bytes.Buffer)
	b.Reset()
	defer renderBufPool.Put(b)

	if err := docTemplate.Execute(b, data); err != nil {
		return "", err
	}
	return b.String(), nil
}

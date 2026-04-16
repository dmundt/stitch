package stitch

import (
	"html/template"

	"github.com/dmundt/stitch/css"
	"github.com/dmundt/stitch/render"
	stitchtpl "github.com/dmundt/stitch/template"
)

// App ties rendering operations to a selected CSS provider.
type App struct {
	Provider css.Provider
}

// Add appends fragment to block on page.
//
// If page or its Composer is nil, Add validates block and returns that
// validation result without appending content.
func Add(page *render.Page, block string, fragment template.HTML) error {
	if page == nil || page.Composer == nil {
		return stitchtpl.ValidateBlocks([]string{block})
	}
	return page.Composer.AddHTML(block, fragment)
}

// NewPage creates a new render.Page with title.
func (a *App) NewPage(title string) *render.Page {
	return render.NewPage(title)
}

// Render renders page with the App provider.
func (a *App) Render(page *render.Page) (string, error) {
	return render.Render(page, a.Provider)
}

// New creates an App using a registered CSS provider name.
func New(providerName string) (*App, error) {
	provider, err := css.Get(providerName)
	if err != nil {
		return nil, err
	}
	return &App{Provider: provider}, nil
}

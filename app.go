package stitch

import (
	"html/template"

	"github.com/dmundt/stitch/css"
	"github.com/dmundt/stitch/render"
	stitchtpl "github.com/dmundt/stitch/template"
)

type App struct {
	Provider css.Provider
}

func Add(page *render.Page, block string, fragment template.HTML) error {
	if page == nil || page.Composer == nil {
		return stitchtpl.ValidateBlocks([]string{block})
	}
	return page.Composer.AddHTML(block, fragment)
}

func (a *App) NewPage(title string) *render.Page {
	return render.NewPage(title)
}

func (a *App) Render(page *render.Page) (string, error) {
	return render.Render(page, a.Provider)
}

func New(providerName string) (*App, error) {
	provider, err := css.Get(providerName)
	if err != nil {
		return nil, err
	}
	return &App{Provider: provider}, nil
}

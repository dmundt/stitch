# Stitch

[![CI](https://github.com/dmundt/stitch/actions/workflows/ci.yml/badge.svg)](https://github.com/dmundt/stitch/actions/workflows/ci.yml) [![Go Reference](https://pkg.go.dev/badge/github.com/dmundt/stitch.svg)](https://pkg.go.dev/github.com/dmundt/stitch)

![Stitch logo](internal/brand/stitch-logo.svg)

Stitch is a public, opinionated, HTML-only server-side rendering framework for Go.

It composes predefined page template blocks using object-style constructors, keeps generated markup CSS-framework agnostic, and lets you inject styles from external CSS frameworks such as minstyle.io or Milligram.

Brand line: **Stitch - compose UIs, agent-first**

## Purpose

Stitch exists to make server-rendered UI development in Go predictable, testable, and portable.

Many SSR codebases drift into one of two extremes: scattered string concatenation that is hard to maintain, or large template files that become difficult to evolve safely. Stitch provides a middle path: explicit page structure with composable, reusable object components.

The framework is intentionally HTML-first. It focuses on semantic markup and reliable rendering, while leaving visual identity to whichever CSS framework you prefer now or later.

## Core Idea

The central idea behind Stitch is: fixed page skeleton, flexible composition.

- Fixed skeleton: pages are composed from predefined structural blocks (`header`, `main`, `footer`) rendered in deterministic order.
- Flexible composition: each block is filled with constructor-based components, including nested layout primitives and content components.
- CSS decoupling: components emit framework-agnostic HTML, and styling is injected through pluggable providers.

This gives you consistency at the document level and freedom at the component/layout level, without coupling your app logic to a specific CSS system.

## Goals

- HTML-only SSR output with semantic markup.
- Code-composed template blocks (no mandatory template files).
- CSS-framework agnostic components.
- Pluggable stylesheet providers.
- Encapsulated, nested element composition for larger UI sections.
- Strong, table-driven and integration test coverage.

## Installation

```bash
go get github.com/dmundt/stitch
```

## Quick Start

```go
package main

import (
    "fmt"
    "html/template"

    "github.com/dmundt/stitch/css"
    "github.com/dmundt/stitch/render"
    "github.com/dmundt/stitch/ui"
)

func main() {
    card := ui.NewArticle("Welcome", []ui.Component{
        ui.NewParagraph("SSR by object-composed blocks."),
        ui.NewButton("Start", "primary"),
    })

    layout := ui.NewSection("Main", []ui.Component{card})

    provider := css.NewStaticProvider("custom", template.HTML(`<link rel="stylesheet" href="https://example.com/app.css">`))
    page := render.NewWindow("Hello Stitch")
    page.TopBarComponent(ui.NewHeading(1, "Hello"))
    page.ContentComponent(layout)

    html, err := page.Render(provider)
    if err != nil {
        panic(err)
    }

    fmt.Println(html)
}
```

## Architecture

Stitch is split into small packages:

- `template`: block constants and composition primitives.
- `render`: page model and deterministic document renderer.
- `ui`: opinionated, semantic HTML components (structs + constructors).
- `css`: provider interface and generic registry/registration utilities.

API policy: `ui` exposes constructor-based object components as the public API. Functional HTML helpers are internal implementation details and are not part of the public surface.

Render GUI layer: `render` provides GUI-style aliases `render.NewWindow`, `TopBar`, `Content`, and `StatusBar` (plus component variants) that map to the page header/main/footer structure.

GUI naming layer: Stitch also exposes GUI-style constructor aliases such as `ui.NewView`, `ui.NewPanel`, `ui.NewText`, `ui.NewAction`, `ui.NewMenu`, `ui.NewWorkspace`, and `ui.NewDataGrid` so application code can read more like desktop/web app UI frameworks.

### Block Composition Model

A page is assembled from predefined blocks:

- header
- main
- footer

You append component output to blocks in code and Stitch renders them in fixed order. This creates consistent structure while still allowing expressive composition.

### Encapsulated Element Composition

Stitch supports nested, object-oriented composition for common layout and content structures.

- `ui.NewSection`, `ui.NewArticle`, `ui.NewDetails`
- `ui.NewRow`, `ui.NewColumn`, `ui.NewGrid`, `ui.NewGridItem`
- `ui.NewContainer`, `ui.NewContainerFluid`, `ui.NewStack`, `ui.NewCluster`, `ui.NewSplit`, `ui.NewSidebarLayout`, `ui.NewAppShell`, `ui.NewHero`
- `ui.NewForm`, `ui.NewFieldset`, `ui.NewSelect`, `ui.NewTextArea`, `ui.NewCheckbox`, `ui.NewRadio`
- `ui.NewTable`, `ui.NewList`, `ui.NewOrderedList`, `ui.NewDescriptionList`, `ui.NewBreadcrumbs`, `ui.NewPagination`
- `ui.NewBlockquote`, `ui.NewBadge`, `ui.NewCodeBlock`, `ui.NewImage`, `ui.NewHeading`, `ui.NewParagraph`
- `ui.NewCard`, `ui.NewThemeToggle`, `ui.NewButton`, `ui.NewNav`

GUI aliases:

- `ui.NewView` -> section
- `ui.NewPanel` -> article/card-like content panel
- `ui.NewText` -> paragraph
- `ui.NewAction` -> button
- `ui.NewMenu` -> nav
- `ui.NewWorkspace` -> app shell
- `ui.NewDataGrid` / `ui.NewDataGridWithClass` -> table
- `ui.NewStatus` -> alert
- `ui.NewToolbar` -> clustered actions

HTMX interaction wrappers:

- `ui.NewInteractiveAction` and `ui.NewInteractiveMenu` accept an `ui.Interaction` value and apply HTMX attributes internally so app code can stay GUI-semantic.

Use these constructors to compose reusable groups of components and then inject them into page blocks with `HeaderComponent`, `MainComponent`, and `FooterComponent`.

Layout note: Stitch treats `row` and `column` as CSS Grid primitives. In the built-in `stitch` provider, `row` is a 12-column grid container and `column` accepts span classes such as `col-4`, `col-6`, and `col-12`.

Object constructors for these primitives are available as `ui.NewRow`, `ui.NewColumn`, `ui.NewGrid`, and `ui.NewGridItem`.
Higher-order layout constructors such as `ui.NewStack`, `ui.NewCluster`, `ui.NewSplit`, `ui.NewSidebarLayout`, `ui.NewAppShell`, and `ui.NewHero` provide complete OO coverage for app-shell and dashboard composition.

### Breadcrumbs And Pagination Controls

Stitch includes dedicated builders for two navigation controls commonly found in CSS frameworks.

```go
controls := ui.NewSection("Navigation Controls", []ui.Component{
    ui.NewBreadcrumbs([]ui.BreadcrumbItem{
        {Label: "Home", Href: "/"},
        {Label: "Library", Href: "/library"},
        {Label: "Components", Current: true},
    }),
    ui.NewPagination([]ui.PageItem{
        {Label: "Previous", Href: "?page=1"},
        {Label: "2", Current: true},
        {Label: "3", Href: "?page=3"},
        {Label: "Next", Href: "?page=3"},
    }),
})
```

The demo app includes a dedicated "Navigation Controls" showcase panel for these elements.

## Common Components Included

Stitch currently implements constructor-based component types for:

- Layout: container, section, article, row, column
- Layout systems: fluid container, stack/cluster, split panes, workspace shell, grid + grid items, hero sections
- Typography/content: headings, paragraph, blockquote, details/summary, code block, image, horizontal rule
- Navigation/data: nav, breadcrumbs, pagination, tables, unordered/ordered/description lists
- Forms: input, select, textarea, checkbox, radio, fieldset, form
- Feedback/actions: alert, badge, button, card

## CSS Provider Injection

Stitch core never hardcodes external CSS frameworks. It ships generic provider infrastructure and includes two built-in providers: `none` (baseline) and `stitch` (built-in default stitch style — flat geometry, restrained borders, readability-first typography).

Framework-specific external providers are defined by applications (the demo app registers minstyle.io and Milligram in [internal/demo/providers.go](internal/demo/providers.go)).

Register custom providers in your own app using `css.NewStaticProvider` and `css.Register`.

## Sample App

A runnable demo server is included and showcases a clean workspace shell, layout systems, and provider switching:

```bash
go run ./cmd/demo
```

Routes:

- `/`
- `/provider/stitch`
- `/provider/none`
- `/provider/minstyle`
- `/provider/milligram`

## Testing

Run everything:

```bash
go test ./...
```

Recommended validation:

```bash
go test -race ./...
go test -cover ./...
go test -bench . ./render -run ^$
```

Test coverage includes:

- template block composition and validation
- render ordering and head injection
- renderer benchmarks for small, large, deep, app-shell, and table-heavy views
- core provider registry behavior and demo provider injection behavior
- every implemented UI element
- sample server integration routes

## Contributing

- Keep API additions in public package boundaries.
- Maintain HTML-only output.
- Every new construct or UI element must include tests.
- Run `go test ./...` before submitting changes.

## Branding

Stitch includes a simple technical identity suitable for README, docs, CLI output, app icons, favicons, and MCP/agent tooling surfaces.

Assets:

- Full logo: [internal/brand/stitch-logo.svg](internal/brand/stitch-logo.svg)
- Monochrome variant: [internal/brand/stitch-logo-mono.svg](internal/brand/stitch-logo-mono.svg)
- Symbol-only mark: [internal/brand/stitch-mark.svg](internal/brand/stitch-mark.svg)

HTTP routes (demo + MCP servers):

- `/branding/stitch.svg`
- `/branding/stitch-mono.svg`
- `/branding/stitch-mark.svg`
- `/favicon.svg`

Export-friendly SVG direction:

- Keep vector source in SVG with fixed `viewBox` (no raster dependency).
- Preserve strokes and rounded joins for terminal-scale legibility.
- Prefer `currentColor` for monochrome assets to adapt to dark/light surfaces.
- Keep symbol mark readable at 16px, 24px, and 32px.

## License

MIT. See LICENSE.


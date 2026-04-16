package render

import (
	"fmt"
	"testing"

	"github.com/dmundt/stitch/css"
	"github.com/dmundt/stitch/ui"
)

func BenchmarkRenderAppShell(b *testing.B) {
	sidebar := ui.NewStack("shell-sidebar-menu", []ui.Component{
		ui.NewNav([]ui.NavLink{{Label: "Overview", Href: "#"}, {Label: "Data", Href: "#"}, {Label: "Settings", Href: "#"}}),
		ui.NewBadge("benchmark", "info"),
	})

	content := ui.NewStack("workspace-stack", []ui.Component{
		ui.NewHero("Dashboard", "Runtime benchmark shell", []ui.Component{ui.NewButton("Action", "primary")}),
		ui.NewSection("Metrics", []ui.Component{
			ui.NewTable([]string{"Name", "Value"}, [][]string{{"CPU", "14%"}, {"Memory", "1.2GB"}, {"Latency", "13ms"}}),
		}),
	})

	page := NewPage("app-shell")
	page.MainComponent(ui.NewAppShell(sidebar, content))
	provider := css.Stitch()

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := page.Render(provider); err != nil {
			b.Fatalf("render failed: %v", err)
		}
	}
}

func BenchmarkRenderDeepTree(b *testing.B) {
	current := ui.Component(ui.NewParagraph("leaf"))
	for i := 0; i < 40; i++ {
		current = ui.NewArticle(fmt.Sprintf("Level %d", i), []ui.Component{current})
	}

	page := NewPage("deep")
	page.MainComponent(current)
	provider := css.Stitch()

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := page.Render(provider); err != nil {
			b.Fatalf("render failed: %v", err)
		}
	}
}

func BenchmarkRenderLargeTree(b *testing.B) {
	page := NewPage("large")
	sections := make([]ui.Component, 0, 24)
	for i := 0; i < 24; i++ {
		sections = append(sections, ui.NewSection(fmt.Sprintf("Section %d", i+1), []ui.Component{
			ui.NewGrid("grid-3", []ui.Component{
				ui.NewGridItem("", []ui.Component{ui.NewCard("Card A", "alpha")}),
				ui.NewGridItem("", []ui.Component{ui.NewCard("Card B", "beta")}),
				ui.NewGridItem("", []ui.Component{ui.NewCard("Card C", "gamma")}),
			}),
		}))
	}
	page.MainComponent(ui.NewStack("", sections))
	provider := css.Stitch()

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := page.Render(provider); err != nil {
			b.Fatalf("render failed: %v", err)
		}
	}
}

func BenchmarkRenderSmallTree(b *testing.B) {
	page := NewPage("small")
	page.MainComponent(ui.NewSection("Overview", []ui.Component{
		ui.NewParagraph("Small render tree."),
		ui.NewButton("Save", "primary"),
	}))
	provider := css.Stitch()

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := page.Render(provider); err != nil {
			b.Fatalf("render failed: %v", err)
		}
	}
}

func BenchmarkRenderTableHeavyView(b *testing.B) {
	rows := make([][]string, 0, 500)
	for i := 0; i < 500; i++ {
		rows = append(rows, []string{fmt.Sprintf("id-%03d", i), fmt.Sprintf("Row %d", i), "active"})
	}

	page := NewPage("table-heavy")
	page.MainComponent(ui.NewSection("Data", []ui.Component{
		ui.NewTable([]string{"ID", "Name", "Status"}, rows),
	}))
	provider := css.Stitch()

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := page.Render(provider); err != nil {
			b.Fatalf("render failed: %v", err)
		}
	}
}


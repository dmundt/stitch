package htmx_test

import (
	"strings"
	"testing"

	"github.com/dmundt/stitch/htmx"
)

func TestBoostNav(t *testing.T) {
	out := htmx.BoostNav([]htmx.NavLink{
		{Label: "Home", Href: "/"},
		{Label: "About", Href: "/about"},
	})
	if !strings.Contains(out, `hx-boost="true"`) {
		t.Errorf("BoostNav() missing hx-boost: %s", out)
	}
	if strings.Contains(out, "hx-get") || strings.Contains(out, "hx-post") {
		t.Errorf("BoostNav() should not have per-link hx-* attrs: %s", out)
	}
	if !strings.Contains(out, `href="/"`) || !strings.Contains(out, "About") {
		t.Errorf("BoostNav() missing links: %s", out)
	}
}

func TestButtonBoost(t *testing.T) {
	out := htmx.Button("Boost", "default", htmx.Attrs{Boost: true})
	if !strings.Contains(out, `hx-boost="true"`) {
		t.Errorf("Button with Boost missing hx-boost: %s", out)
	}
}

func TestButtonDefaultKind(t *testing.T) {
	out := htmx.Button("Click", "", htmx.Attrs{})
	if !strings.Contains(out, "btn-default") {
		t.Errorf("Button() should use default kind: %s", out)
	}
}

func TestButtonEmptyAttrs(t *testing.T) {
	out := htmx.Button("Plain", "secondary", htmx.Attrs{})
	if strings.Contains(out, "hx-") {
		t.Errorf("Button with empty Attrs should have no hx- attributes: %s", out)
	}
}

func TestButtonHtmxAttrs(t *testing.T) {
	out := htmx.Button("Load", "primary", htmx.Attrs{Get: "/items", Target: "#list", Swap: "innerHTML"})
	for _, want := range []string{`hx-get="/items"`, `hx-target="#list"`, `hx-swap="innerHTML"`, "Load"} {
		if !strings.Contains(out, want) {
			t.Errorf("Button() missing %q: %s", want, out)
		}
	}
}

func TestForm(t *testing.T) {
	out := htmx.Form("/submit", "post", htmx.Attrs{Post: "/submit", Target: "#resp", Swap: "innerHTML"}, "<input>")
	for _, want := range []string{`hx-post="/submit"`, `hx-target="#resp"`, `action="/submit"`, "<input>"} {
		if !strings.Contains(out, want) {
			t.Errorf("Form() missing %q: %s", want, out)
		}
	}
}

func TestFormEmptyAttrs(t *testing.T) {
	out := htmx.Form("/go", "get", htmx.Attrs{}, "<button>")
	if strings.Contains(out, "hx-") {
		t.Errorf("Form with empty Attrs should have no hx- attributes: %s", out)
	}
}

func TestHead(t *testing.T) {
	h := string(htmx.Head())
	if !strings.Contains(h, "htmx.org") {
		t.Errorf("Head() missing htmx.org: %s", h)
	}
	if !strings.Contains(h, "<script") {
		t.Errorf("Head() missing script tag: %s", h)
	}
}

func TestNav(t *testing.T) {
	out := htmx.Nav([]htmx.NavLink{
		{Label: "Home", Href: "/", HX: htmx.Attrs{Get: "/partial/", Target: "#main", Swap: "outerHTML", PushURL: "/"}},
	})
	for _, want := range []string{`hx-get="/partial/"`, `hx-target="#main"`, `hx-push-url="/"`, "Home"} {
		if !strings.Contains(out, want) {
			t.Errorf("Nav() missing %q: %s", want, out)
		}
	}
}

func TestNavMultipleLinks(t *testing.T) {
	out := htmx.Nav([]htmx.NavLink{
		{Label: "A", Href: "/a", HX: htmx.Attrs{Get: "/pa"}},
		{Label: "B", Href: "/b", HX: htmx.Attrs{Get: "/pb"}},
	})
	if !strings.Contains(out, `hx-get="/pa"`) || !strings.Contains(out, `hx-get="/pb"`) {
		t.Errorf("Nav() missing one of the links: %s", out)
	}
}

func TestNavRendersLinkMetadata(t *testing.T) {
	out := htmx.Nav([]htmx.NavLink{
		{
			Label: "Home",
			Href:  "/",
			ID:    "home-link",
			Class: "nav-link",
			Attrs: map[string]string{
				"data-route": "home",
				"aria-label": "Home route",
			},
		},
	})
	for _, want := range []string{`id="home-link"`, `class="nav-link"`, `data-route="home"`, `aria-label="Home route"`} {
		if !strings.Contains(out, want) {
			t.Errorf("Nav() missing %q: %s", want, out)
		}
	}
}

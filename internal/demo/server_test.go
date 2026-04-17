package demo

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBuildDemoPageIncludesComponents(t *testing.T) {
	html, err := BuildDemoPage("none")
	if err != nil {
		t.Fatalf("build failed: %v", err)
	}

	for _, needle := range []string{"Stitch Component Showcase", "Layout &amp; Structure", "Layout Systems", "Content Elements", "Form Elements", "Tables &amp; Navigation", "Navigation Controls", "<table>", "<button", "<nav>", "<img", "<fieldset>", "<details>", "aria-label=\"Breadcrumb\"", "aria-label=\"Pagination\"", "class=\"row\"", "class=\"column col-8\"", "class=\"column col-4\"", "class=\"grid grid-3\"", "class=\"grid-item span-2\"", "class=\"hero\"", "class=\"split\"", "class=\"app-shell app-shell-single\"", "hx-get=\"/provider/stitch\""} {
		if !strings.Contains(html, needle) {
			t.Fatalf("missing component marker %q", needle)
		}
	}
}

func TestBuildDemoPageMinstyleTableClass(t *testing.T) {
	html, err := BuildDemoPage("minstyle")
	if err != nil {
		t.Fatalf("build failed: %v", err)
	}
	if !strings.Contains(html, "class=\"ms-table ms-striped\"") {
		t.Fatal("expected minstyle table classes in demo output")
	}

	htmlNone, err := BuildDemoPage("none")
	if err != nil {
		t.Fatalf("build failed: %v", err)
	}
	if strings.Contains(htmlNone, "ms-table") {
		t.Fatal("did not expect minstyle table classes for non-minstyle provider")
	}
}

func TestBuildDemoPageProviderValidation(t *testing.T) {
	if _, err := BuildDemoPage("missing"); err == nil {
		t.Fatal("expected unknown provider error")
	}
}

func TestBuildDemoPageSectionConsistencyAcrossProviders(t *testing.T) {
	providers := []string{"stitch", "none", "minstyle", "milligram"}
	needles := []string{"Layout &amp; Structure", "Layout Systems", "Content Elements", "Form Elements", "Tables &amp; Navigation", "Navigation Controls"}

	for _, provider := range providers {
		html, err := BuildDemoPage(provider)
		if err != nil {
			t.Fatalf("provider %s build failed: %v", provider, err)
		}
		for _, needle := range needles {
			if !strings.Contains(html, needle) {
				t.Fatalf("provider %s missing section %q", provider, needle)
			}
		}
	}
}

func TestRoutes(t *testing.T) {
	h := NewHandler()
	cases := []struct {
		path   string
		status int
		body   string
	}{
		{path: "/", status: http.StatusOK, body: "Stitch Component Showcase"},
		{path: "/assets/stitch.css", status: http.StatusOK, body: "--stitch-accent"},
		{path: "/branding/stitch.svg", status: http.StatusOK, body: "Stitch Logo"},
		{path: "/branding/stitch-mono.svg", status: http.StatusOK, body: "Stitch Logo Monochrome"},
		{path: "/branding/stitch-mark.svg", status: http.StatusOK, body: "Stitch Mark"},
		{path: "/favicon.svg", status: http.StatusOK, body: "Stitch Mark"},
		{path: "/provider/stitch", status: http.StatusOK, body: "stitch"},
		{path: "/provider/milligram", status: http.StatusOK, body: "milligram"},
		{path: "/provider/nope", status: http.StatusBadRequest, body: "unknown provider"},
	}

	for _, tc := range cases {
		r := httptest.NewRequest(http.MethodGet, tc.path, nil)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		resp := w.Result()
		if resp.StatusCode != tc.status {
			t.Fatalf("path %s expected %d got %d", tc.path, tc.status, resp.StatusCode)
		}
		if !strings.Contains(w.Body.String(), tc.body) {
			t.Fatalf("path %s missing %q", tc.path, tc.body)
		}
	}
}

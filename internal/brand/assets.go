package brand

import (
	"embed"
	"net/http"
)

const (
	Name      = "Stitch"
	Tagline   = "compose UIs, agent-first"
	BrandLine = "Stitch - compose UIs, agent-first"
)

const (
	PathLogo     = "/branding/stitch.svg"
	PathLogoMono = "/branding/stitch-mono.svg"
	PathMark     = "/branding/stitch-mark.svg"
	PathMarkMono = "/branding/stitch-mark-mono.svg"
	PathFavicon  = "/favicon.svg"
)

//go:embed stitch-logo.svg stitch-logo-mono.svg stitch-mark.svg stitch-mark-mono.svg
var assets embed.FS

func LogoSVG() []byte {
	b, _ := assets.ReadFile("stitch-logo.svg")
	return b
}

func LogoMonoSVG() []byte {
	b, _ := assets.ReadFile("stitch-logo-mono.svg")
	return b
}

func MarkSVG() []byte {
	b, _ := assets.ReadFile("stitch-mark.svg")
	return b
}

func MarkMonoSVG() []byte {
	b, _ := assets.ReadFile("stitch-mark-mono.svg")
	return b
}

func FaviconLinkTag() string {
	return `<link rel="icon" type="image/svg+xml" href="` + PathFavicon + `">`
}

func MountRoutes(mux *http.ServeMux) {
	serveSVG := func(w http.ResponseWriter, content []byte) {
		w.Header().Set("Content-Type", "image/svg+xml; charset=utf-8")
		w.Header().Set("Cache-Control", "public, max-age=3600")
		_, _ = w.Write(content)
	}

	mux.HandleFunc(PathLogo, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		serveSVG(w, LogoSVG())
	})
	mux.HandleFunc(PathLogoMono, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		serveSVG(w, LogoMonoSVG())
	})
	mux.HandleFunc(PathMark, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		serveSVG(w, MarkSVG())
	})
	mux.HandleFunc(PathMarkMono, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		serveSVG(w, MarkMonoSVG())
	})
	mux.HandleFunc(PathFavicon, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		serveSVG(w, MarkSVG())
	})
}

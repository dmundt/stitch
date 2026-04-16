package css

import (
	"embed"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"sync"
)

//go:embed assets/*.css
var assetsFS embed.FS

var (
	mu        sync.RWMutex
	providers = map[string]Provider{}
)

type Provider interface {
	Name() string
	HeadHTML() template.HTML
}

type StaticProvider struct {
	name string
	head template.HTML
}

// Assets returns the core stylesheet assets filesystem rooted at asset files.
func Assets() (fs.FS, error) {
	return fs.Sub(assetsFS, "assets")
}

func Get(name string) (Provider, error) {
	mu.RLock()
	defer mu.RUnlock()
	p, ok := providers[name]
	if !ok {
		return nil, fmt.Errorf("unknown provider: %s", name)
	}
	return p, nil
}

func MustRegister(p Provider) {
	if err := Register(p); err != nil {
		panic(err)
	}
}

func NewStaticProvider(name string, head template.HTML) Provider {
	return StaticProvider{name: name, head: head}
}

func None() Provider {
	return StaticProvider{name: "none", head: ""}
}

func Register(p Provider) error {
	if p == nil {
		return errors.New("provider is nil")
	}
	if p.Name() == "" {
		return errors.New("provider name is required")
	}
	mu.Lock()
	defer mu.Unlock()
	providers[p.Name()] = p
	return nil
}

func (p StaticProvider) HeadHTML() template.HTML {
	return p.head
}

func (p StaticProvider) Name() string {
	return p.name
}

// Stitch returns the core Fluent-inspired default provider.
func Stitch() Provider {
	return StaticProvider{name: "stitch", head: template.HTML(`<link rel="stylesheet" href="/assets/stitch.css">`)}
}

func init() {
	MustRegister(None())
	MustRegister(Stitch())
}

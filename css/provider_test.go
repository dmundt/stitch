package css

import (
	"html/template"
	"strings"
	"testing"
)

func TestBuiltInProvidersRegistered(t *testing.T) {
	for _, name := range []string{"none", "stitch"} {
		p, err := Get(name)
		if err != nil {
			t.Fatalf("expected provider %s: %v", name, err)
		}
		if p.Name() != name {
			t.Fatalf("expected name %s, got %s", name, p.Name())
		}
	}
}

func TestProviderHeadHTML(t *testing.T) {
	cases := []struct {
		provider Provider
		needle   string
	}{
		{None(), ""},
		{Stitch(), "/assets/stitch.css"},
		{NewStaticProvider("test", template.HTML(`<link rel="stylesheet" href="https://example.com/test.css">`)), "example.com/test.css"},
	}

	for _, tc := range cases {
		head := string(tc.provider.HeadHTML())
		if tc.needle != "" && !strings.Contains(head, tc.needle) {
			t.Fatalf("provider %s missing needle %q in %q", tc.provider.Name(), tc.needle, head)
		}
	}
}

func TestRegisterAndGetCustomProvider(t *testing.T) {
	provider := NewStaticProvider("custom", template.HTML(`<link rel="stylesheet" href="https://example.com/custom.css">`))
	if err := Register(provider); err != nil {
		t.Fatalf("register failed: %v", err)
	}
	got, err := Get("custom")
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if !strings.Contains(string(got.HeadHTML()), "custom.css") {
		t.Fatalf("unexpected head html: %s", got.HeadHTML())
	}
}

func TestRegisterValidation(t *testing.T) {
	if err := Register(nil); err == nil {
		t.Fatal("expected nil provider error")
	}
	if err := Register(StaticProvider{}); err == nil {
		t.Fatal("expected missing name error")
	}
}

func TestAssets(t *testing.T) {
	fs, err := Assets()
	if err != nil {
		t.Fatalf("Assets() returned error: %v", err)
	}
	if fs == nil {
		t.Fatal("expected non-nil filesystem")
	}
	f, err := fs.Open("stitch.css")
	if err != nil {
		t.Fatalf("expected stitch.css in assets: %v", err)
	}
	_ = f.Close()
}

func TestMustRegisterPanicsOnNil(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected MustRegister(nil) to panic")
		}
	}()
	MustRegister(nil)
}

func TestGetUnknownProvider(t *testing.T) {
	_, err := Get("absolutely-unknown-provider")
	if err == nil {
		t.Fatal("expected error for unknown provider")
	}
	if !strings.Contains(err.Error(), "unknown provider") {
		t.Fatalf("unexpected error: %v", err)
	}
}

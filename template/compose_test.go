package template

import (
	"html/template"
	"strings"
	"testing"
)

func TestComposerAddHTMLAndFragments(t *testing.T) {
	c := NewComposer()
	if err := c.AddHTML(BlockMain, template.HTML("<p>ok</p>")); err != nil {
		t.Fatalf("add html failed: %v", err)
	}
	items, err := c.Fragments(BlockMain)
	if err != nil {
		t.Fatalf("fragments failed: %v", err)
	}
	if len(items) != 1 || string(items[0]) != "<p>ok</p>" {
		t.Fatalf("unexpected fragments: %#v", items)
	}
}

func TestComposerAddStringEscapes(t *testing.T) {
	c := NewComposer()
	if err := c.AddString(BlockMain, "<script>alert(1)</script>"); err != nil {
		t.Fatalf("add string failed: %v", err)
	}
	items, _ := c.Fragments(BlockMain)
	if !strings.Contains(string(items[0]), "&lt;script&gt;") {
		t.Fatalf("expected escaped script, got: %s", items[0])
	}
}

func TestComposerInvalidBlock(t *testing.T) {
	c := NewComposer()
	if err := c.AddHTML("nope", "x"); err == nil {
		t.Fatal("expected invalid block error")
	}
	if _, err := c.Fragments("nope"); err == nil {
		t.Fatal("expected invalid block error")
	}
}

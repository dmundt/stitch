package template

import (
	"errors"
	"html"
	"html/template"
)

// Composer stores HTML fragments grouped by page block.
type Composer struct {
	blocks map[string][]template.HTML
}

// AddHTML appends a trusted HTML fragment to block.
func (c *Composer) AddHTML(block string, fragment template.HTML) error {
	if !IsValidBlock(block) {
		return errors.New("invalid block: " + block)
	}
	c.blocks[block] = append(c.blocks[block], fragment)
	return nil
}

// AddString escapes fragment and appends it to block.
func (c *Composer) AddString(block string, fragment string) error {
	escaped := template.HTML(html.EscapeString(fragment))
	return c.AddHTML(block, escaped)
}

// Fragments returns a copy of block fragments.
func (c *Composer) Fragments(block string) ([]template.HTML, error) {
	if !IsValidBlock(block) {
		return nil, errors.New("invalid block: " + block)
	}
	out := make([]template.HTML, len(c.blocks[block]))
	copy(out, c.blocks[block])
	return out, nil
}

// NewComposer returns a Composer initialized with known blocks.
func NewComposer() *Composer {
	c := &Composer{blocks: map[string][]template.HTML{}}
	for _, block := range OrderedBlocks() {
		c.blocks[block] = []template.HTML{}
	}
	return c
}

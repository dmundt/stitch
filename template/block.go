package template

import (
	"errors"
	"sort"
)

const (
	// BlockHeader is the header composition block name.
	BlockHeader = "header"
	// BlockMain is the main composition block name.
	BlockMain = "main"
	// BlockFooter is the footer composition block name.
	BlockFooter = "footer"
)

var orderedBlocks = []string{BlockHeader, BlockMain, BlockFooter}

// IsValidBlock reports whether name is a known block.
func IsValidBlock(name string) bool {
	for _, block := range orderedBlocks {
		if name == block {
			return true
		}
	}
	return false
}

// OrderedBlocks returns the canonical block order copy.
func OrderedBlocks() []string {
	out := make([]string, len(orderedBlocks))
	copy(out, orderedBlocks)
	return out
}

// SortByBlockOrder sorts blocks in canonical block order.
func SortByBlockOrder(blocks []string) error {
	if err := ValidateBlocks(blocks); err != nil {
		return err
	}

	rank := map[string]int{}
	for idx, name := range orderedBlocks {
		rank[name] = idx
	}

	sort.Slice(blocks, func(i, j int) bool {
		return rank[blocks[i]] < rank[blocks[j]]
	})
	return nil
}

// ValidateBlocks validates that every name is a known block.
func ValidateBlocks(input []string) error {
	for _, name := range input {
		if !IsValidBlock(name) {
			return errors.New("invalid block: " + name)
		}
	}
	return nil
}

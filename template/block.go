package template

import (
	"errors"
	"sort"
)

const (
	BlockHeader = "header"
	BlockMain   = "main"
	BlockFooter = "footer"
)

var orderedBlocks = []string{BlockHeader, BlockMain, BlockFooter}

func IsValidBlock(name string) bool {
	for _, block := range orderedBlocks {
		if name == block {
			return true
		}
	}
	return false
}

func OrderedBlocks() []string {
	out := make([]string, len(orderedBlocks))
	copy(out, orderedBlocks)
	return out
}

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

func ValidateBlocks(input []string) error {
	for _, name := range input {
		if !IsValidBlock(name) {
			return errors.New("invalid block: " + name)
		}
	}
	return nil
}
